// Package escape provides the API client for the Escape Platform
package escape

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/env"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/version"
)

// Debug is a flag to enable debug mode for the API client
var Debug = false

// arrayQueryContextKey carries the query values a caller attached with
// withArrayQuery. Unexported, so only this package can produce one.
type arrayQueryContextKey struct{}

// rawBodyContextKey carries the JSON body a caller attached with withRawBody.
// Unexported, so only this package can produce one.
type rawBodyContextKey struct{}

// withArrayQuery attaches query values to ctx for the next request the
// generated client builds. Empty values leave ctx unchanged.
func withArrayQuery(ctx context.Context, values url.Values) context.Context {
	if len(values) == 0 {
		return ctx
	}

	return context.WithValue(ctx, arrayQueryContextKey{}, values)
}

// addArrayQuery records key repeated once per value. An empty list leaves the
// key unset.
func addArrayQuery(values url.Values, key string, list []string) {
	if len(list) > 0 {
		values[key] = list
	}
}

// addScalarQuery records key with a single value. An empty value leaves the
// key unset.
func addScalarQuery(values url.Values, key, value string) {
	if value != "" {
		values[key] = []string{value}
	}
}

// withRawBody attaches the JSON body the next generated request must send.
// The generated client marshals its own body and drops explicit nulls. The
// transport substitutes this body, so the wire stays the caller's JSON while
// Execute still returns *v3.GenericOpenAPIError.
func withRawBody(ctx context.Context, body []byte) context.Context {
	copied := make([]byte, len(body))
	copy(copied, body)

	return context.WithValue(ctx, rawBodyContextKey{}, copied)
}

// arrayQueryTransport rewrites a generated request before it is sent. Two
// things the generator cannot express travel on the request context: repeated
// query keys (withArrayQuery) and a raw JSON body (withRawBody). Query keys
// are repeated because parameterAddToHeaderOrQuery has no branch for a
// oneOf string | string[] and would write "<Go type> value". The body override
// is how UpdateProfile keeps explicit nulls that ToMap would drop.
type arrayQueryTransport struct {
	base http.RoundTripper
}

func (t arrayQueryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if overlay, ok := req.Context().Value(arrayQueryContextKey{}).(url.Values); ok && len(overlay) > 0 {
		req = withQueryOverlay(req, overlay)
	}

	if body, ok := req.Context().Value(rawBodyContextKey{}).([]byte); ok {
		req = requestWithRawBody(req, body)
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("transport request: %w", err)
	}

	return resp, nil
}

// withQueryOverlay returns a copy of req whose query carries overlay, replacing
// any value the generated client added for the same key.
func withQueryOverlay(req *http.Request, overlay url.Values) *http.Request {
	query := req.URL.Query()
	for key, values := range overlay {
		query.Del(key)
		for _, value := range values {
			query.Add(key, value)
		}
	}

	cloned := req.Clone(req.Context())
	cloned.URL.RawQuery = query.Encode()

	return cloned
}

// requestWithRawBody returns a copy of req whose body is body. ContentLength
// and GetBody follow the new bytes: the generated client has already set them
// from the body it marshaled, which is shorter once explicit nulls are dropped.
func requestWithRawBody(req *http.Request, body []byte) *http.Request {
	cloned := req.Clone(req.Context())
	if cloned.Body != nil && cloned.Body != req.Body {
		_ = cloned.Body.Close()
	}

	if req.Body != nil {
		_ = req.Body.Close()
	}

	cloned.Body = io.NopCloser(bytes.NewReader(body))
	cloned.ContentLength = int64(len(body))
	cloned.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	cloned.Header.Del("Content-Length")

	return cloned
}

// newAPIV3Client creates a new API v3 client
func newAPIV3Client() (*v3.APIClient, error) {
	log.Trace("Initializing v3 client")
	url, err := env.GetAPIURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get API URL: %w", err)
	}

	authorization, err := env.GetAuthorizationHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization header: %w", err)
	}

	httpClient := env.GetHTTPClient()
	base := httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}

	httpClient.Transport = arrayQueryTransport{base: base}

	cfg := v3.Configuration{
		Host:   url.Host,
		Scheme: url.Scheme,
		DefaultHeader: map[string]string{
			"Authorization": authorization,
		},
		UserAgent:  version.GetVersion().UserAgent(),
		Debug:      Debug,
		HTTPClient: httpClient,
		Servers: []v3.ServerConfiguration{
			{
				URL: url.String() + "/v3",
			},
		},
	}

	return v3.NewAPIClient(&cfg), nil
}
