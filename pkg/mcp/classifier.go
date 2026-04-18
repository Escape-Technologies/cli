package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	classifierEndpointEnv  = "MCP_CLASSIFIER_ENDPOINT"
	classifierAPIKeyEnv    = "MCP_CLASSIFIER_API_KEY"
	classifierModelEnv     = "MCP_CLASSIFIER_MODEL"
	classifierTimeoutMSEnv = "MCP_CLASSIFIER_TIMEOUT_MS"
	classifierTopKEnv      = "MCP_CLASSIFIER_TOP_K"

	defaultClassifierModel     = "gpt-5-nano"
	defaultClassifierTimeoutMS = 2000
	defaultClassifierTopK      = 15

	// httpClientGracePeriod is added to the request timeout so the underlying
	// http.Client doesn't fire before the per-request context deadline does
	// (the context-driven cancellation produces nicer wrapped errors).
	httpClientGracePeriod = 500 * time.Millisecond

	// classifierErrorBodyPeekBytes bounds how much of an HTTP error body we
	// surface on a non-2xx classifier response.
	classifierErrorBodyPeekBytes = 512

	// classifierUnparseableSnippetChars limits how much of an unparseable
	// classifier reply we echo back in the error string.
	classifierUnparseableSnippetChars = 200

	classifierSystemPrompt = `You rank CLI tools by relevance to a user's chat message and recent history.
Return ONLY a JSON object of the form {"tools":["name1","name2",...]}. Names must be picked from the provided catalog. At most K items, sorted by relevance (most relevant first). No commentary.`
)

// ChatContext is the payload carried in the X-Escape-Chat-Context header. It
// holds the user's current message and a few recent turns, so the classifier
// can resolve ambiguous pronouns and domain paraphrases (e.g. "findings"
// vs "issues").
type ChatContext struct {
	Current string             `json:"current"`
	History []ChatContextEntry `json:"history,omitempty"`
}

// ChatContextEntry is one prior chat turn.
type ChatContextEntry struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolDigest is the compact tool metadata the classifier sees.
type ToolDigest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Classifier ranks tools by their relevance to a chat context. Nil Classifier
// means "not configured"; callers should treat that as "no ranking available"
// and fall back to a compact catalog.
type Classifier interface {
	Rank(ctx context.Context, chatCtx ChatContext, digest []ToolDigest) ([]string, error)
	TopK() int
}

// NewClassifierFromEnv builds a Classifier from the MCP_CLASSIFIER_* env vars.
// If the endpoint or api key are missing, returns (nil, nil) so callers can
// opt-in without hard-failing at boot.
func NewClassifierFromEnv() (Classifier, error) {
	endpoint := strings.TrimSpace(os.Getenv(classifierEndpointEnv))
	apiKey := strings.TrimSpace(os.Getenv(classifierAPIKeyEnv))
	if endpoint == "" || apiKey == "" {
		return nil, nil
	}

	model := strings.TrimSpace(os.Getenv(classifierModelEnv))
	if model == "" {
		model = defaultClassifierModel
	}

	timeout := defaultClassifierTimeoutMS
	if raw := strings.TrimSpace(os.Getenv(classifierTimeoutMSEnv)); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("invalid %s=%q", classifierTimeoutMSEnv, raw)
		}
		timeout = parsed
	}

	topK := defaultClassifierTopK
	if raw := strings.TrimSpace(os.Getenv(classifierTopKEnv)); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("invalid %s=%q", classifierTopKEnv, raw)
		}
		topK = parsed
	}

	return &openAIClassifier{
		endpoint: strings.TrimRight(endpoint, "/"),
		apiKey:   apiKey,
		model:    model,
		timeout:  time.Duration(timeout) * time.Millisecond,
		topK:     topK,
		http: &http.Client{
			Timeout: time.Duration(timeout)*time.Millisecond + httpClientGracePeriod,
		},
	}, nil
}

type openAIClassifier struct {
	endpoint string
	apiKey   string
	model    string
	timeout  time.Duration
	topK     int
	http     *http.Client
}

func (c *openAIClassifier) TopK() int { return c.topK }

type classifierRequest struct {
	Model          string              `json:"model"`
	Messages       []classifierMessage `json:"messages"`
	ResponseFormat map[string]string   `json:"response_format,omitempty"`
	Temperature    float64             `json:"temperature,omitempty"`
}

type classifierMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type classifierResponse struct {
	Choices []struct {
		Message classifierMessage `json:"message"`
	} `json:"choices"`
}

type classifierPayload struct {
	Current string             `json:"current"`
	History []ChatContextEntry `json:"history,omitempty"`
	Catalog []ToolDigest       `json:"catalog"`
	TopK    int                `json:"top_k"`
}

func (c *openAIClassifier) Rank(
	parent context.Context,
	chatCtx ChatContext,
	digest []ToolDigest,
) ([]string, error) {
	if strings.TrimSpace(chatCtx.Current) == "" && len(chatCtx.History) == 0 {
		return nil, errors.New("classifier requires chat context")
	}
	if len(digest) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(parent, c.timeout)
	defer cancel()

	userBody, err := json.Marshal(classifierPayload{
		Current: chatCtx.Current,
		History: chatCtx.History,
		Catalog: digest,
		TopK:    c.topK,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal classifier payload: %w", err)
	}

	reqBody, err := json.Marshal(classifierRequest{
		Model: c.model,
		Messages: []classifierMessage{
			{Role: "system", Content: classifierSystemPrompt},
			{Role: "user", Content: string(userBody)},
		},
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal classifier request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint+"/chat/completions",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("new classifier request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("classifier request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, classifierErrorBodyPeekBytes))
		return nil, fmt.Errorf("classifier http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed classifierResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode classifier response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, errors.New("classifier returned no choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	tools, err := parseClassifierResult(content)
	if err != nil {
		return nil, err
	}
	// The system prompt asks for at most K items, but models occasionally
	// overshoot. Hard-clamp here so intent mode can't accidentally re-expand
	// more schemas than the size budget allows.
	if c.topK > 0 && len(tools) > c.topK {
		tools = tools[:c.topK]
	}
	return tools, nil
}

// parseClassifierResult pulls the tool-name list out of the model's reply.
// The model is asked for a strict JSON object, but we accept a bare array
// too so intermittent protocol drift does not sink the whole call.
func parseClassifierResult(content string) ([]string, error) {
	if content == "" {
		return nil, errors.New("empty classifier response")
	}

	// Treat both `{"tools":[]}` and `[]` as valid "no relevant tool" outcomes
	// rather than parser errors — these are legitimate classifier outputs.
	// The strict shape `{"tools":[...]}` wins over the bare-array fallback so
	// a model that drifts back to the documented format always parses cleanly.
	var objectForm struct {
		Tools []string `json:"tools"`
	}
	if err := json.Unmarshal([]byte(content), &objectForm); err == nil && objectForm.Tools != nil {
		return trimAndDedupe(objectForm.Tools), nil
	}

	var arrayForm []string
	if err := json.Unmarshal([]byte(content), &arrayForm); err == nil && arrayForm != nil {
		return trimAndDedupe(arrayForm), nil
	}

	return nil, fmt.Errorf("unparseable classifier response: %s", truncate(content, classifierUnparseableSnippetChars))
}

func trimAndDedupe(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		name := strings.TrimSpace(value)
		if name == "" {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result
}

func truncate(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}
