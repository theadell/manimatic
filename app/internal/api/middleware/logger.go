package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

func HTTPLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			tracker := &ResponseTracker{
				ResponseWriter: w,
				status:         http.StatusOK,
				start:          time.Now(),
			}

			next.ServeHTTP(tracker, r)

			duration := time.Since(tracker.start)
			logFields := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", tracker.status),
				slog.Duration("duration", duration),
				slog.Int("response_size", tracker.size),
			}

			logLevel := determineLogLevel(tracker.status)
			logger.LogAttrs(context.Background(), logLevel, "HTTP Request", logFields...)
		})
	}
}

// ResponseTracker captures response details
type ResponseTracker struct {
	http.ResponseWriter
	status int
	size   int
	start  time.Time
}

func (rt *ResponseTracker) WriteHeader(status int) {
	rt.status = status
	rt.ResponseWriter.WriteHeader(status)
}

func (rt *ResponseTracker) Write(b []byte) (int, error) {
	rt.size += len(b)
	return rt.ResponseWriter.Write(b)
}

// Unwrap for compatibility with ResponseController
func (rt *ResponseTracker) Unwrap() http.ResponseWriter {
	return rt.ResponseWriter
}

// determineLogLevel sets log level based on HTTP status
func determineLogLevel(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
