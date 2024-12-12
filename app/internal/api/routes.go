package api

import (
	"manimatic/internal/api/middleware"
	"net/http"
)

func (a *App) setupRoutes() http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("POST /generate", a.HandleGenerate)
	mux.HandleFunc("GET /events", a.sseHandler)

	return mux

}

func (a *App) setupMiddleware(h http.Handler) http.Handler {
	recovery := middleware.RecoveryMiddleware(a.logger)
	return middleware.Chain(h, recovery, a.sm.LoadAndSave, middleware.EnsureSessionTokenMiddleware(a.sm))

}
