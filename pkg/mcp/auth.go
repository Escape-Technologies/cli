// Package mcp implements the embedded MCP server that exposes the Escape CLI
// commands as MCP tools over HTTP.
package mcp

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type authContextKey struct{}

// Auth carries the credentials extracted from an incoming MCP request,
// forwarded to the CLI subprocess through sanitized environment variables.
type Auth struct {
	APIKey        string
	Authorization string
}

// InjectAuthContext derives a context that carries the credentials extracted
// from the incoming HTTP request so downstream tool handlers can access them.
func InjectAuthContext(ctx context.Context, req *http.Request) context.Context {
	if req == nil {
		return ctx
	}
	auth := Auth{
		APIKey:        strings.TrimSpace(req.Header.Get("X-ESCAPE-API-KEY")),
		Authorization: strings.TrimSpace(req.Header.Get("Authorization")),
	}
	if auth.Authorization != "" && auth.APIKey == "" {
		auth.APIKey = apiKeyFromAuthorization(auth.Authorization)
	}
	return context.WithValue(ctx, authContextKey{}, auth)
}

// AuthFromContext recovers the Auth value previously stored by
// InjectAuthContext and reports an error if the request lacks credentials.
func AuthFromContext(ctx context.Context) (Auth, error) {
	auth, ok := ctx.Value(authContextKey{}).(Auth)
	if !ok {
		return Auth{}, errors.New("missing authentication context")
	}
	if auth.APIKey == "" && auth.Authorization == "" {
		return Auth{}, errors.New("missing X-ESCAPE-API-KEY or Authorization header")
	}
	return auth, nil
}

func apiKeyFromAuthorization(authorization string) string {
	if !strings.HasPrefix(strings.ToLower(authorization), "key ") {
		return ""
	}

	return strings.TrimSpace(authorization[4:])
}
