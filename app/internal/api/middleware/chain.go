package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"slices"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
)

type Middleware func(http.Handler) http.Handler

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {

	for _, middleware := range slices.Backward(middlewares) {
		handler = middleware(handler)
	}

	return handler
}

const UserSessionTokenKey string = "user_session_token"

func EnsureSessionTokenMiddleware(sm *scs.SessionManager, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := sm.GetString(r.Context(), UserSessionTokenKey)

			if token == "" {
				token = uuid.NewString()
				slog.Debug("created a new session token", "session_id", token)
				sm.Put(r.Context(), UserSessionTokenKey, token)
			}
			ctx := context.WithValue(r.Context(), UserSessionTokenKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
