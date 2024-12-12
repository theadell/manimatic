package events

import (
	"fmt"
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

// MessageRouter handles routing messages to specific SSE clients
type MessageRouter struct {
	mu      sync.RWMutex
	clients map[string]chan Message
}

// NewMessageRouter creates a new message router
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		clients: make(map[string]chan Message),
	}
}

// AddClient registers a new client for a specific session
func (mr *MessageRouter) AddClient(sessionID string) (<-chan Message, func()) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	messageChan := make(chan Message, 10)
	mr.clients[sessionID] = messageChan

	fmt.Println("added client for session ", sessionID)
	return messageChan, func() {
		mr.mu.Lock()
		defer mr.mu.Unlock()
		fmt.Println("delete client for session", sessionID)
		close(messageChan)
		delete(mr.clients, sessionID)
	}
}

// SendMessage sends a message to a specific session
func (mr *MessageRouter) SendMessage(msg Message) error {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

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
