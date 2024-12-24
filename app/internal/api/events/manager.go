package events

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
)

type Event struct {
	Kind      string `json:"kind"`       // What type of event this is
	SessionID string `json:"session_id"` // Session this event belongs to
	Data      any    `json:"data"`       // The event payload
}

// All possible event kinds
const (
	// Core events
	KindCompileRequested  = "compile_requested"  // Request to compile a script
	KindCompileSucceeded  = "compile_succeeded"  // Compilation succeeded
	KindCompileFailed     = "compile_failed"     // Compilation failed
	KindGenerateSucceeded = "generate_succeeded" // Script generation succeeded
	KindGenerateFailed    = "generate_failed"
)

// CompileRequest represents a request to compile a script
type CompileRequest struct {
	Script string `json:"script"`
}

// CompileSuccess represents successful compilation
type CompileSuccess struct {
	VideoURL string `json:"video_url"`
}

// CompileError represents a compilation failure
type CompileError struct {
	Message string `json:"message"`        // User-friendly error message
	Stdout  string `json:"stdout"`         // Standard output from compilation
	Stderr  string `json:"stderr"`         // Standard error from compilation
	Line    int    `json:"line,omitempty"` // Line number where error occurred (if available)
}

type GenerateSuccess = CompileRequest

type GenerateError struct {
	Message string `json:"message"`           // User-friendly error message
	Details string `json:"details,omitempty"` // Optional additional context
	Model   string `json:"model"`             // Which model failed (e.g. "gpt-4", "claude", etc)
}

// Helper functions to create events
func NewCompileRequest(sessionID, script string) Event {
	return Event{
		Kind:      KindCompileRequested,
		SessionID: sessionID,
		Data:      CompileRequest{Script: script},
	}
}

func NewCompileSuccess(sessionID, videoURL string) Event {
	return Event{
		Kind:      KindCompileSucceeded,
		SessionID: sessionID,
		Data:      CompileSuccess{VideoURL: videoURL},
	}
}

func NewCompileError(sessionID, message string, stdout, stderr string, line int) Event {
	return Event{
		Kind:      KindCompileFailed,
		SessionID: sessionID,
		Data: CompileError{
			Message: message,
			Stdout:  stdout,
			Stderr:  stderr,
			Line:    line,
		},
	}
}
func NewGenerateSuccess(sessionID, script string) Event {
	return Event{
		Kind:      KindGenerateSucceeded,
		SessionID: sessionID,
		Data: GenerateSuccess{
			Script: script,
		},
	}
}

func NewGenerateError(sessionID, message, details, model string) Event {
	return Event{
		Kind:      KindGenerateFailed,
		SessionID: sessionID,
		Data: GenerateError{
			Message: message,
			Details: details,
			Model:   model,
		},
	}
}

type rawEvent struct {
	Kind      string          `json:"kind"`
	SessionID string          `json:"session_id"`
	Data      json.RawMessage `json:"data"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	var raw rawEvent
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal raw event: %w", err)
	}

	e.Kind = raw.Kind
	e.SessionID = raw.SessionID

	var err error
	switch raw.Kind {
	case KindCompileRequested:
		var d CompileRequest
		err = json.Unmarshal(raw.Data, &d)
		e.Data = d

	case KindCompileSucceeded:
		var d CompileSuccess
		err = json.Unmarshal(raw.Data, &d)
		e.Data = d

	case KindCompileFailed:
		var d CompileError
		err = json.Unmarshal(raw.Data, &d)
		e.Data = d

	case KindGenerateSucceeded:
		var d GenerateSuccess
		err = json.Unmarshal(raw.Data, &d)
		e.Data = d

	case KindGenerateFailed:
		var d GenerateError
		err = json.Unmarshal(raw.Data, &d)
		e.Data = d

	default:
		return fmt.Errorf("unknown event kind: %s", raw.Kind)
	}

	if err != nil {
		return fmt.Errorf("failed to unmarshal event data for kind %s: %w", raw.Kind, err)
	}

	return nil
}

type MessageRouter struct {
	mu      sync.RWMutex
	clients map[string]chan Event
	done    chan struct{}
	log     *slog.Logger
}

func NewMessageRouter(log *slog.Logger) *MessageRouter {
	return &MessageRouter{
		clients: make(map[string]chan Event),
		done:    make(chan struct{}),
		log:     log,
	}
}

func (mr *MessageRouter) AddClient(sessionID string) (<-chan Event, func()) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.log.Debug("Added a new SSE client", "session_id", sessionID)

	// Create a buffered channel with a context
	messageChan := make(chan Event, 10)
	mr.clients[sessionID] = messageChan

	return messageChan, func() {
		mr.mu.Lock()
		defer mr.mu.Unlock()
		close(messageChan)
		delete(mr.clients, sessionID)
		mr.log.Debug("Removed an SSE client", "session_id", sessionID)

	}
}

func (mr *MessageRouter) SendMessage(msg Event) error {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	select {
	case <-mr.done:
		return fmt.Errorf("message router is shutting down")
	default:
		clientChan, exists := mr.clients[msg.SessionID]
		if !exists {
			return fmt.Errorf("no active client for session %s", msg.SessionID)
		}

		select {
		case clientChan <- msg:
			return nil
		default:
			return fmt.Errorf("message channel full for session %s", msg.SessionID)
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
