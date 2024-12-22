package api

import (
	"manimatic/internal/api/features"
	"manimatic/internal/api/middleware"
	"net/http"
	"strings"

	"github.com/rs/cors"
)

const (
	localhostPrefixHTTP = "http://localhost"
	domain              = ".adelh.dev"
)

func (a *App) setupRoutes() http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("POST /generate", a.HandleGenerate)
	mux.HandleFunc("GET /events", a.sseHandler)
	mux.HandleFunc("GET /healthz", healthCheckHandler)
	mux.HandleFunc("GET /features", a.featuresHandler)

	if a.config.Features.IsEnabled(features.UserCompile) {
		mux.HandleFunc("POST /compile", a.handleCompile)
	}

	return mux

}

func (a *App) setupMiddleware(h http.Handler) http.Handler {
	recovery := middleware.PanicRecovery(a.logger)
	requestLogger := middleware.HTTPLogger(a.logger)
	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {

			if strings.HasPrefix(origin, localhostPrefixHTTP) {
				return true
			}
			return strings.HasSuffix(origin, domain)
		},
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut},
		AllowedHeaders:   []string{"*"},
	})

	handler := c.Handler(h)
	return middleware.Chain(handler, recovery, middleware.RealIP, requestLogger, a.sm.LoadAndSave, middleware.EnsureSessionTokenMiddleware(a.sm, a.logger))

}
