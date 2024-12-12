package events

import (
	"fmt"
	"log/slog"
	"sync"
)

type MessageType string

const (
	MessageTypeScriptUpdate MessageType = "script"
	MessageTypeVideoUpdate  MessageType = "video"
)

type MessageStatus string

const (
	MessageStatusSuccess MessageStatus = "success"
	MessageStatusError   MessageStatus = "error"
)

type Message struct {
	Type      MessageType    `json:"type"`
	SessionId string         `json:"session_id"`
	Status    MessageStatus  `json:"status"`
	Content   any            `json:"content"`
	Details   map[string]any `json:"details,omitempty"`
}

type MessageRouter struct {
	mu      sync.RWMutex
	clients map[string]chan Message
	done    chan struct{}
	log     *slog.Logger
}

func NewMessageRouter(log *slog.Logger) *MessageRouter {
	return &MessageRouter{
		clients: make(map[string]chan Message),
		done:    make(chan struct{}),
		log:     log,
	}
}

func (mr *MessageRouter) AddClient(sessionID string) (<-chan Message, func()) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.log.Debug("Added a new SSE client", "session_id", sessionID)

	// Create a buffered channel with a context
	messageChan := make(chan Message, 10)
	mr.clients[sessionID] = messageChan

	return messageChan, func() {
		mr.mu.Lock()
		defer mr.mu.Unlock()
		close(messageChan)
		delete(mr.clients, sessionID)
		mr.log.Debug("Removed an SSE client", "session_id", sessionID)

	}
}

func (mr *MessageRouter) SendMessage(msg Message) error {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	select {
	case <-mr.done:
		return fmt.Errorf("message router is shutting down")
	default:
		clientChan, exists := mr.clients[msg.SessionId]
		if !exists {
			return fmt.Errorf("no active client for session %s", msg.SessionId)
		}

		select {
		case clientChan <- msg:
			return nil
		default:
			return fmt.Errorf("message channel full for session %s", msg.SessionId)
		}
	}
}

func (mr *MessageRouter) Shutdown() {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	// Signal that no more messages should be sent
	close(mr.done)

	// Close all client channels
	for sessionID, ch := range mr.clients {
		close(ch)
		delete(mr.clients, sessionID)
	}
}
