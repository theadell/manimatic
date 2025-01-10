package worker

import (
	"context"
	"errors"
	"log/slog"
	"manimatic/internal/config"
	"manimatic/internal/worker/animation"
	"manimatic/internal/worker/manimexec"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var errQueueNotExist = &types.QueueDoesNotExist{}

type VideoStorage interface {
	UploadAndPresign(ctx context.Context, outputPath string, sessionId string) (string, error)
}

type WorkerService struct {
	config        *config.Config
	log           *slog.Logger
	storage       VideoStorage
	queue         *animation.Queue
	workerPool    *WorkerPool
	cancelContext context.Context
	cancelFunc    context.CancelFunc
	executer      *manimexec.Executor
}

func NewWorkerService(cfg *config.Config, queue *animation.Queue, storage VideoStorage, log *slog.Logger) (*WorkerService, error) {

	ctx, cancel := context.WithCancel(context.Background())

	workerPool := NewWorkerPool(cfg.Processing.MaxConcurrency, log)

	return &WorkerService{
		config:        cfg,
		log:           log,
		queue:         queue,
		storage:       storage,
		workerPool:    workerPool,
		cancelContext: ctx,
		cancelFunc:    cancel,
		executer:      manimexec.MustNewExecutor(cfg),
	}, nil
}

func (ws *WorkerService) Cleanup() {
	ws.cancelFunc()
	ws.workerPool.Stop()
}

func (ws *WorkerService) Run() {

	ws.workerPool.Start(ws.processTask)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ws.cancelContext.Done():
				return
			default:
				ws.fetchAndSubmitMessage()
			}
		}
	}()

	<-sigChan
	ws.log.Info("Shutting down gracefully...")
	ws.Cleanup()
	ws.log.Info("Server shutdown")
	os.Exit(0)
}

func (ws *WorkerService) fetchAndSubmitMessage() {
	t, err := ws.queue.ReceiveTask(ws.cancelContext)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			ws.log.Debug("context cancelled - returning")
			return
		}
		ws.handleReceiveMessageError(err)
		return
	}
	if !t.Valid {
		// no message available yet
		return
	}

	ws.workerPool.Submit(Task{
		event:          t.E,
		compileRequest: t.R,
		h:              t.H,
	})
}

func (ws *WorkerService) processTask(task Task) error {

	res, err := ws.executer.ExecuteScript(context.TODO(), task.compileRequest.Script, task.event.SessionID)
	if err != nil {
		ws.log.Error("failed to execute manim script", "error", err.Error())
		if err := ws.queue.PublishResult(ws.cancelContext, animation.NewErrorResult(task.event.SessionID, err)); err != nil {
			ws.log.Error("Failed to enqueue error event", "error", err)
		}
		go func() {
			err := ws.queue.DeleteTask(ws.cancelContext, task.h)
			if err != nil {
				ws.log.Error("failed to delete task", "error", err, "handle", task.h)
			}
		}()
		return nil
	}

	outputPath, taskDir := res.OutputPath, res.WorkingDir
	defer os.RemoveAll(taskDir)

	go func() {
		err := ws.queue.DeleteTask(ws.cancelContext, task.h)
		if err != nil {
			ws.log.Error("failed to delete task", "error", err, "handle", task.h)
		}
	}()
	// upload the video, create presigned url and send a success message to the API
	go func() {

		url, err := ws.storage.UploadAndPresign(context.TODO(), outputPath, task.event.SessionID)
		if err != nil {
			return
		}

		err = ws.queue.PublishResult(ws.cancelContext, animation.NewSuccessResult(task.event.SessionID, url))
		if err != nil {
			ws.log.Error("failed to send message", "err", err)
			return
		}

	}()

	return nil
}

func (ws *WorkerService) handleReceiveMessageError(err error) {
	ws.log.Error("Failed to receive message", "error", err)
	if errors.As(err, &errQueueNotExist) {
		slog.Error("Queue does not exist", "URL", ws.config.AWS.TaskQueueURL)
		os.Exit(1)
	}
}
