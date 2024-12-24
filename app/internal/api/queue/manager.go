package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"manimatic/internal/api/events"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type QueueManager struct {
	client         *sqs.Client
	tasQueueURL    string
	resultQueueURL string
}

func New(client *sqs.Client, queueURL, resultQueueURL string) *QueueManager {
	return &QueueManager{client: client, tasQueueURL: queueURL, resultQueueURL: resultQueueURL}
}

func (q *QueueManager) EnqeueMsg(ctx context.Context, msg *events.Event) error {
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message %w", err)
	}
	input := &sqs.SendMessageInput{
		QueueUrl:    &q.tasQueueURL,
		MessageBody: aws.String(string(jsonMessage)),
	}
	_, err = q.client.SendMessage(ctx, input)
	return err
}

func (qm *QueueManager) ReceiveSingleMessage(ctx context.Context) ([]types.Message, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(qm.resultQueueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
	}

	resp, err := qm.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, err
	}

	return resp.Messages, nil
}

func (qm *QueueManager) DeleteMessage(ctx context.Context, msg types.Message) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(qm.resultQueueURL),
		ReceiptHandle: msg.ReceiptHandle,
	}

	_, err := qm.client.DeleteMessage(ctx, input)
	return err
}
