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
	Type    MessageType    `json:"type"`
	Status  MessageStatus  `json:"status"`
	Content any            `json:"content"`
	Details map[string]any `json:"details,omitempty"`
}

type ConnectionManager struct {
	mu          sync.Mutex
	connections map[string]chan Message
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]chan Message),
	}
}

func (cm *ConnectionManager) AddConnection(id string) chan Message {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ch := make(chan Message)
	cm.connections[id] = ch
	return ch
}

func (cm *ConnectionManager) RemoveConnection(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ch, ok := cm.connections[id]; ok {
		close(ch)
		delete(cm.connections, id)
	}
}

func (cm *ConnectionManager) SendMessage(id string, message Message) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ch, ok := cm.connections[id]; ok {
		select {
		case ch <- message:
			return nil
		default:
			return fmt.Errorf("connection %s is unresponsive", id)
		}
	}
	return fmt.Errorf("connection %s not found", id)
}
