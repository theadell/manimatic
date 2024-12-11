package api

import (
	"log/slog"
	"manimatic/internal/api/events"
	"manimatic/internal/api/genmanim"
	"manimatic/internal/api/session"
	"manimatic/internal/config"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type App struct {
	config       *config.Config
	logger       *slog.Logger
	router       http.Handler
	manimService *genmanim.LLMManimService
	sm           *scs.SessionManager
	connMgr      *events.ConnectionManager
}

func New(cfg *config.Config, logger *slog.Logger, manimService *genmanim.LLMManimService) *App {
	app := &App{
		config:       cfg,
		logger:       logger,
		manimService: manimService,
		sm:           session.New(),
		connMgr:      events.NewConnectionManager(),
	}

	h := app.setupRoutes()
	app.router = app.setupMiddleware(h)
	return app
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}
