package mcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type authContextKey struct{}

type Auth struct {
	APIKey        string
	Authorization string
}

func InjectAuthContext(ctx context.Context, req *http.Request) context.Context {
	auth := Auth{
		APIKey:        strings.TrimSpace(req.Header.Get("X-ESCAPE-API-KEY")),
		Authorization: strings.TrimSpace(req.Header.Get("Authorization")),
	}
	if auth.Authorization != "" && auth.APIKey == "" {
		auth.APIKey = apiKeyFromAuthorization(auth.Authorization)
	}
	return context.WithValue(ctx, authContextKey{}, auth)
}

func AuthFromContext(ctx context.Context) (Auth, error) {
	auth, ok := ctx.Value(authContextKey{}).(Auth)
	if !ok {
		return Auth{}, fmt.Errorf("missing authentication context")
	}
	if auth.APIKey == "" && auth.Authorization == "" {
		return Auth{}, fmt.Errorf("missing X-ESCAPE-API-KEY or Authorization header")
	}
	return auth, nil
}

func apiKeyFromAuthorization(authorization string) string {
	if !strings.HasPrefix(strings.ToLower(authorization), "key ") {
		return ""
	}

	return strings.TrimSpace(authorization[4:])
}
