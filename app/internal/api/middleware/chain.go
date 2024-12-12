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

func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Recovered from panic",
						"error", err,
						"method", r.Method,
						"url", r.URL.String(),
						"remote_addr", r.RemoteAddr,
					)

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

const UserSessionTokenKey string = "user_session_token"

func EnsureSessionTokenMiddleware(sm *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := sm.GetString(r.Context(), UserSessionTokenKey)

			if token == "" {
				token = uuid.NewString()
				sm.Put(r.Context(), UserSessionTokenKey, token)
			}
			ctx := context.WithValue(r.Context(), UserSessionTokenKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
