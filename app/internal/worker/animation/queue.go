package animation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"manimatic/internal/api/events"
	"manimatic/internal/worker/manimexec"
	"manimatic/pkg/queue"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type ResultType int

const (
	ResultTypeSuccess ResultType = iota
	ResultTypeError
)

type Result struct {
	Type      ResultType
	SessionID string
	VideoURL  string // filled only if Type is Success
	Error     error  // filled only if Type is Error
}

type TaskMessage struct {
	E     *events.Event          // Event
	R     *events.CompileRequest // Compile Request
	H     *string                // Recepient Handle
	Valid bool                   //Is valid
}

func NewSuccessResult(sessionID, videoURL string) *Result {
	return &Result{
		Type:      ResultTypeSuccess,
		SessionID: sessionID,
		VideoURL:  videoURL,
	}
}

func NewErrorResult(sessionID string, err error) *Result {
	return &Result{
		Type:      ResultTypeError,
		SessionID: sessionID,
		Error:     err,
	}
}

type MessageQueue interface {
	ReceiveMessage(ctx context.Context) (*types.Message, error)
	DeleteMessage(ctx context.Context, receiptHandle *string) error
	SendMessage(ctx context.Context, body any) error
}

type Queue struct {
	queue MessageQueue
	log   *slog.Logger
}

func NewQueue(queue MessageQueue, log *slog.Logger) *Queue {
	return &Queue{
		queue: queue,
		log:   log,
	}
}
func (q *Queue) ReceiveTask(ctx context.Context) (*TaskMessage, error) {
	msg, err := q.queue.ReceiveMessage(ctx)
	if err != nil {
		if errors.Is(err, queue.ErrNoMessages) {
			return &TaskMessage{}, nil
		}

		return nil, fmt.Errorf("receive failed: %w", err)
	}

	var event events.Event
	if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	if event.Kind != events.KindCompileRequested {
		return nil, fmt.Errorf("unexpected event kind: %s", event.Kind)
	}

	req, ok := event.Data.(events.CompileRequest)
	if !ok {
		return nil, fmt.Errorf("invalid data type for %s event", events.KindCompileRequested)
	}

	return &TaskMessage{E: &event, R: &req, H: msg.ReceiptHandle, Valid: true}, nil
}

func (q *Queue) DeleteTask(ctx context.Context, receiptHandle *string) error {
	return q.queue.DeleteMessage(ctx, receiptHandle)
}

func (q *Queue) PublishResult(ctx context.Context, result *Result) error {

	switch result.Type {

	case ResultTypeSuccess:
		event := events.NewCompileSuccess(result.SessionID, result.VideoURL)
		return q.queue.SendMessage(ctx, event)
	case ResultTypeError:
		return q.publishError(ctx, result.SessionID, result.Error)
	default:
		q.log.Warn("Unknown result type", "type", result.Type, "result", result)
		return nil
	}
}

func (q *Queue) publishError(ctx context.Context, sessionID string, err error) error {
	var execErr *manimexec.ExecutionError

	if !errors.As(err, &execErr) {
		event := events.NewCompileError(
			sessionID,
			err.Error(),
			"", "", 0,
		)
		return q.queue.SendMessage(ctx, event)
	}

	switch execErr.Kind {
	case manimexec.ErrorKindCompilation:
		event := events.NewCompileError(
			sessionID,
			execErr.Message,
			execErr.Stdout,
			execErr.Stderr,
			execErr.Line,
		)
		return q.queue.SendMessage(ctx, event)

	case manimexec.ErrorKindSecurity:
		event := events.NewCompileError(
			sessionID,
			fmt.Sprintf("Security error: %s", execErr.Message),
			execErr.Stdout,
			execErr.Stderr,
			execErr.Line,
		)
		return q.queue.SendMessage(ctx, event)

	case manimexec.ErrorKindTimeout:
		event := events.NewCompileError(
			sessionID,
			"Animation generation timed out",
			execErr.Stdout,
			execErr.Stderr,
			execErr.Line,
		)
		return q.queue.SendMessage(ctx, event)

	default:
		event := events.NewCompileError(
			sessionID,
			execErr.Message,
			execErr.Stdout,
			execErr.Stderr,
			execErr.Line,
		)
		return q.queue.SendMessage(ctx, event)
	}
}
