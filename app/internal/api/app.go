package api

import (
	"log/slog"
	"manimatic/internal/api/events"
	"manimatic/internal/api/queue"
	"manimatic/internal/api/session"
	"manimatic/internal/config"
	"manimatic/internal/llm"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type App struct {
	config     *config.Config
	logger     *slog.Logger
	router     http.Handler
	llmService *llm.Service
	sm         *scs.SessionManager
	MsgRouter  *events.MessageRouter
	queueMgr   *queue.QueueManager
}

func New(cfg *config.Config, logger *slog.Logger, llmService *llm.Service, sqsClient *sqs.Client) *App {
	app := &App{
		config:     cfg,
		logger:     logger,
		llmService: llmService,
		sm:         session.New(),
		MsgRouter:  events.NewMessageRouter(logger),
		queueMgr:   queue.New(sqsClient, cfg.AWS.TaskQueueURL, cfg.AWS.ResultQueueURL),
	}

	h := app.setupRoutes()
	app.router = app.setupMiddleware(h)
	return app
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}
