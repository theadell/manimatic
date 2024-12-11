package api

import (
	"context"
	"encoding/json"
	"fmt"
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

	sessionID := a.sm.GetString(r.Context(), string(middleware.UserSessionTokenKey))
	if sessionID == "" {
		a.serverError(w, fmt.Errorf("invalid, missing or expired session"))
		return
	}

	w.WriteHeader(http.StatusNoContent)

	go func() {
		ctx := context.Background()
		result, err := a.manimService.GenerateScript(ctx, req.Prompt, true)
		if err != nil {
			a.logger.Error("failed to generate script", "error", err)
			_ = a.connMgr.SendMessage(sessionID, events.Message{
				Type:   events.MessageTypeScriptUpdate,
				Status: events.MessageStatusError,
				Content: map[string]any{
					"message": "Failed to generate script",
					"details": map[string]any{
						"reason": err.Error(),
					},
				},
			})
			return
		}
		a.logger.Info("generated manim script", "session_id", sessionID)

		err = a.connMgr.SendMessage(sessionID, events.Message{
			Type:   events.MessageTypeScriptUpdate,
			Status: events.MessageStatusSuccess,
			Content: map[string]string{
				"script": result.Code,
			},
		})
		if err != nil {
			a.logger.Error("failed to send message to client channel", "session_id", sessionID, "error", err)
		}
	}()

}

func (a *App) sseHandler(w http.ResponseWriter, r *http.Request) {

	id := a.sm.Token(r.Context())
	if id == "" {
		a.serverError(w, fmt.Errorf("failed to load session"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)

	noDeadline := time.Time{}
	rc.SetReadDeadline(noDeadline)
	rc.SetWriteDeadline(noDeadline)

	// Add connection to manager
	ch := a.connMgr.AddConnection(id)
	defer a.connMgr.RemoveConnection(id)

	// Send events from the channel to the client
	enc := json.NewEncoder(w)
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				// Channel closed, exit
				return
			}
			err := enc.Encode(msg)
			if err != nil {
				a.serverError(w, err, fmt.Errorf("failed to encode message: %w", err))
				return
			}
			err = rc.Flush()
			if err != nil {
				a.serverError(w, err, fmt.Errorf("failed to flush response: %w", err))
				return
			}

		case <-r.Context().Done():
			return
		}
	}
}
