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

// AuthMethod identifies how the incoming request carried its credentials.
// Emitted as telemetry on every /mcp call so we can eventually retire the
// legacy paths (X-ESCAPE-API-KEY and `Key`) for the publicly-hosted MCP.
type AuthMethod string

const (
	// AuthMethodNone means the request carried no credentials.
	AuthMethodNone AuthMethod = "none"
	// AuthMethodAPIKeyHeader means the request used the legacy X-ESCAPE-API-KEY header.
	AuthMethodAPIKeyHeader AuthMethod = "x-escape-api-key"
	// AuthMethodAuthorizationKey means the request used `Authorization: Key <api-key>`.
	AuthMethodAuthorizationKey AuthMethod = "key"
	// AuthMethodAuthorizationBearer means the request used `Authorization: Bearer <api-key>`,
	// the OAuth 2.1 path.
	AuthMethodAuthorizationBearer AuthMethod = "bearer"
)

// Auth carries the credentials extracted from an incoming MCP request,
// forwarded to the CLI subprocess through sanitized environment variables.
type Auth struct {
	APIKey        string
	Authorization string
	// Method records how the credential was delivered; used for telemetry.
	Method AuthMethod
}

// InjectAuthContext derives a context that carries the credentials extracted
// from the incoming HTTP request so downstream tool handlers can access them.
//
// Precedence: X-ESCAPE-API-KEY wins over Authorization when both are present,
// matching the pre-OAuth behavior. When the credential is extracted from a
// `Bearer` authorization header the raw header is NOT propagated: the backend
// API treats `Bearer <token>` as a JWT (verify), so leaking the header into
// ESCAPE_AUTHORIZATION would cause every downstream call to 401. See
// packages/cli/pkg/env/key.go GetAuthorizationHeader and
// services/api/src/lib/core/context/accessor.ts for the precedence that
// motivates this.
func InjectAuthContext(ctx context.Context, req *http.Request) context.Context {
	if req == nil {
		return ctx
	}
	headerAPIKey := strings.TrimSpace(req.Header.Get("X-ESCAPE-API-KEY"))
	rawAuthorization := strings.TrimSpace(req.Header.Get("Authorization"))

	auth := Auth{Method: AuthMethodNone}

	switch {
	case headerAPIKey != "":
		auth.APIKey = headerAPIKey
		// Drop the Authorization header even when X-ESCAPE-API-KEY wins:
		// executor.go forwards Auth.Authorization as ESCAPE_AUTHORIZATION,
		// and env/key.go prefers that over ESCAPE_API_KEY, so preserving
		// rawAuthorization here would let a stale `Bearer <token>` still
		// reach the backend as a JWT and 401 every tool call.
		auth.Authorization = ""
		auth.Method = AuthMethodAPIKeyHeader
	case rawAuthorization != "":
		scheme, token := splitAuthorization(rawAuthorization)
		switch scheme {
		case "bearer":
			auth.APIKey = token
			// Intentional: drop the Bearer header so the child CLI does not
			// forward it via ESCAPE_AUTHORIZATION. env.GetAuthorizationHeader
			// will rebuild `Key <api-key>` from ESCAPE_API_KEY.
			auth.Authorization = ""
			auth.Method = AuthMethodAuthorizationBearer
		case "key":
			auth.APIKey = token
			auth.Authorization = rawAuthorization
			auth.Method = AuthMethodAuthorizationKey
		default:
			auth.Authorization = rawAuthorization
		}
	}

	return context.WithValue(ctx, authContextKey{}, auth)
}

// AuthFromContext recovers the Auth value previously stored by
// InjectAuthContext and reports an error if the request lacks credentials.
func AuthFromContext(ctx context.Context) (Auth, error) {
	if ctx == nil {
		return Auth{}, errors.New("missing authentication context")
	}
	auth, ok := ctx.Value(authContextKey{}).(Auth)
	if !ok {
		return Auth{}, errors.New("missing authentication context")
	}
	if auth.APIKey == "" && auth.Authorization == "" {
		return Auth{}, errors.New("missing X-ESCAPE-API-KEY or Authorization header")
	}
	return auth, nil
}

// authorizationParts is the number of space-separated parts expected in a
// well-formed Authorization header (scheme + token).
const authorizationParts = 2

// splitAuthorization returns the lowercase scheme and trimmed token from
// a raw Authorization header. Returns ("", "") on malformed input.
func splitAuthorization(authorization string) (scheme, token string) {
	parts := strings.SplitN(authorization, " ", authorizationParts)
	if len(parts) != authorizationParts {
		return "", ""
	}
	return strings.ToLower(strings.TrimSpace(parts[0])), strings.TrimSpace(parts[1])
}

func apiKeyFromAuthorization(authorization string) string {
	scheme, token := splitAuthorization(authorization)
	switch scheme {
	case "key", "bearer":
		return token
	default:
		return ""
	}
}
