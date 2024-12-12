package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"manimatic/internal/api/events"
	"manimatic/internal/api/middleware"
	"net/http"
	"time"
)

type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

func (a *App) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest

	err := ReadJSON(w, r, &req)
	if err != nil || req.Prompt == "" {
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
		result, err := a.manimService.GenerateScript(ctx, req.Prompt, true)
		var msg events.Message
		if err != nil {
			a.logger.Error("failed to generate script", "error", err)
			msg = events.Message{
				Type:      events.MessageTypeScriptUpdate,
				Status:    events.MessageStatusError,
				SessionId: sessionID,
				Content: map[string]any{
					"message": "Failed to generate script",
					"details": map[string]any{
						"reason": err.Error(),
					},
				},
			}
			_ = a.MsgRouter.SendMessage(msg)
			return
		}

		msg = events.Message{
			Type:      events.MessageTypeScriptUpdate,
			SessionId: sessionID,
			Status:    events.MessageStatusSuccess,
			Content:   result.Code,
		}
		a.logger.Info("generated manim script", "session_id", sessionID)
		go func() {
			err := a.queueMgr.EnqeueMsg(context.TODO(), &msg)
			if err != nil {
				slog.Error("failed to enqueue message", "error", err, "message", msg)
			}
		}()
		err = a.MsgRouter.SendMessage(msg)
		if err != nil {
			a.logger.Error("failed to send message to client channel", "session_id", sessionID, "error", err)
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
			}
		}
	}
}
