package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListAIChats returns one page of the copilot chats of the caller.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListAIChats(ctx context.Context, next string, size int) ([]v3.AIChatSummary, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.AIChatsAPI.ListAIChats(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// GetAIChat returns one chat with a preview of its messages. The preview holds
// the first 100 messages; ListAIChatMessages walks the full history.
func GetAIChat(ctx context.Context, chatID string) (*v3.GetAIChat200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.AIChatsAPI.GetAIChat(ctx, chatID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// ListAIChatMessages returns one page of the messages of a chat, oldest first.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListAIChatMessages(ctx context.Context, chatID string, next string, size int) ([]v3.AIChatMessage, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.AIChatsAPI.ListAIChatMessages(ctx, chatID)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// CreateAIChat opens a chat.
func CreateAIChat(ctx context.Context, body v3.CreateAIChatRequest) (*v3.CreateAIChat200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.AIChatsAPI.CreateAIChat(ctx).CreateAIChatRequest(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// CreateAIChatMessage posts a user message to a chat.
func CreateAIChatMessage(ctx context.Context, chatID string, body v3.CreateAIChatMessageRequest) (*v3.CreateAIChatMessage200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.AIChatsAPI.CreateAIChatMessage(ctx, chatID).CreateAIChatMessageRequest(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
