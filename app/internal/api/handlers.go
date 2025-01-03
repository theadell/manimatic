package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"manimatic/internal/api/events"
	"manimatic/internal/api/middleware"
	"manimatic/internal/llm"
	"net/http"
	"time"
)

type GenerateRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}
type CompileRequest struct {
	Script string `json:"script"`
}

func (a *App) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest

	err := ReadJSON(w, r, &req)
	if err != nil || len(req.Prompt) < 8 {
		a.badRequestResponse(w, "invalid request body")
	}

	sessionID := a.sm.GetString(r.Context(), middleware.UserSessionTokenKey)
	if sessionID == "" {
		a.serverError(w, fmt.Errorf("invalid, missing or expired session"))
		return
	}

	w.WriteHeader(http.StatusNoContent)

	go func() {
		ctx := context.Background()
		result, err := a.llmService.Generate(ctx, req.Prompt, req.Model)
		var msg events.Event
		if err != nil {
			a.logger.Error("failed to generate script", "error", err)
			msg = events.NewGenerateError(sessionID, "failed to generate script", err.Error(), "openai-4o")
			_ = a.MsgRouter.SendMessage(msg)
			return
		}
		if !result.ValidInput || result.Code == "" {
			a.logger.Info("generated script flagged as invalid or empty", "prompt", req.Prompt)
			msg = events.NewGenerateError(sessionID, "failed to generate scene for the given prompt", result.Warnings, "openai-4o")
			_ = a.MsgRouter.SendMessage(msg)
			return
		}

		clientUpdate := events.NewGenerateSuccess(sessionID, result.Code)
		workerTask := events.NewCompileRequest(sessionID, result.Code)
		a.logger.Info("generated manim script", "session_id", sessionID)
		go func() {
			err := a.queueMgr.EnqeueMsg(context.TODO(), &workerTask)
			if err != nil {
				slog.Error("failed to enqueue message", "error", err, "message", msg)
			}
		}()
		err = a.MsgRouter.SendMessage(clientUpdate)
		if err != nil {
			a.logger.Error("failed to send message to client channel", "session_id", sessionID, "error", err)
		}
	}()

}

func (a *App) handleCompile(w http.ResponseWriter, r *http.Request) {
	var req CompileRequest

	err := ReadJSON(w, r, &req)
	if err != nil || len(req.Script) < 8 {
		a.badRequestResponse(w, "invalid request body")
	}

	sessionID := a.sm.GetString(r.Context(), middleware.UserSessionTokenKey)
	if sessionID == "" {
		a.serverError(w, fmt.Errorf("invalid, missing or expired session"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	go func() {
		msg := events.NewCompileRequest(sessionID, req.Script)
		err := a.queueMgr.EnqeueMsg(context.TODO(), &msg)
		if err != nil {
			slog.Error("failed to enqueue message", "error", err, "message", msg)
		}
	}()
}

func (a *App) sseHandler(w http.ResponseWriter, r *http.Request) {
	id := a.sm.GetString(r.Context(), string(middleware.UserSessionTokenKey))
	if id == "" {
		a.serverError(w, fmt.Errorf("invalid, missing or expired session"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	rc := http.NewResponseController(w)

	noDeadline := time.Time{}
	rc.SetReadDeadline(noDeadline)
	rc.SetWriteDeadline(noDeadline)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	messageChan, cleanup := a.MsgRouter.AddClient(id)
	defer cleanup()

	// Event loop
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-messageChan:
			if !ok {
				a.logger.Debug("Client Event channel Closed")
				return
			}

			jsonData, err := json.Marshal(msg)
			if err != nil {
				a.logger.Error("failed to serialize message", "error", err)
				continue
			}

			_, err = fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
			if err != nil {
				a.logger.Error("failed to write event into SSE connection")
			}
			if err = rc.Flush(); err != nil {
				a.logger.Error("failed to flush event message into client", "err", err)
				return
			}
		}
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (a *App) featuresHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, a.config.Processing.Features)
}

func (a *App) modelsHandler(w http.ResponseWriter, _ *http.Request) {
	response := llm.ModelsResponse{
		Models:       a.llmService.AvailableModels(),
		DefaultModel: a.llmService.DefaultModel(),
	}
	WriteJSON(w, http.StatusOK, response)
}
