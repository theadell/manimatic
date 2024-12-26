package api

import (
	"context"
	"encoding/json"
	"errors"
	"manimatic/internal/api/events"
	"time"
)

func (a *App) StartMessageProcessor(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				a.logger.Info("Message processor shutting down")
				return
			default:
				if err := a.processNextVideoUpdateMessage(ctx); err != nil {
					if err == context.Canceled {
						return
					}
					if err != ErrNoMessagesAvailable {
						a.logger.Error("Error processing message", "error", err)
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}

var ErrNoMessagesAvailable = errors.New("no messages available")

func (a *App) processNextVideoUpdateMessage(ctx context.Context) error {
	// Receive a single message with a short wait time
	messages, err := a.queueMgr.ReceiveSingleMessage(ctx)
	if err != nil {
		return err
	}

	// If no message received, return special error
	if len(messages) == 0 {
		return ErrNoMessagesAvailable
	}

	msg := messages[0]
	var ev events.Event
	err = json.Unmarshal([]byte(*msg.Body), &ev)
	if err != nil {
		a.logger.Error("Failed to unmarshal message", "error", err)
		_ = a.queueMgr.DeleteMessage(ctx, msg)
		return err
	}
	a.logger.Debug("processing event", "kind", ev.Kind, "session_id", ev.SessionID)

	err = a.MsgRouter.SendMessage(ev)
	if err != nil {
		a.logger.Error("Failed to send message to connection manager",
			"session_id", ev.SessionID,
			"error", err)
	}

	return a.queueMgr.DeleteMessage(ctx, msg)
}
