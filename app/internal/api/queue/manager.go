package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"manimatic/internal/api/events"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type QueueManager struct {
	client   *sqs.Client
	queueURL string
}

func New(client *sqs.Client, queueURL string) *QueueManager {
	return &QueueManager{client: client, queueURL: queueURL}
}

func (q *QueueManager) EnqeueMsg(ctx context.Context, msg *events.Message) error {
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message %w", err)
	}
	input := &sqs.SendMessageInput{
		QueueUrl:    &q.queueURL,
		MessageBody: aws.String(string(jsonMessage)),
	}
	_, err = q.client.SendMessage(ctx, input)
	return err
}
