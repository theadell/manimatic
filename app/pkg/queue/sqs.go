package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQS struct {
	client    *sqs.Client
	taskURL   string
	resultURL string
	log       *slog.Logger
}

var ErrNoMessages = errors.New("no messages available")

func NewSQS(client *sqs.Client, taskURL, resultURL string, log *slog.Logger) *SQS {
	return &SQS{
		client:    client,
		taskURL:   taskURL,
		resultURL: resultURL,
		log:       log,
	}
}

func (q *SQS) ReceiveMessage(ctx context.Context) (*types.Message, error) {
	result, err := q.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.taskURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		return nil, fmt.Errorf("receive failed: %w", err)
	}

	if len(result.Messages) == 0 {
		return nil, ErrNoMessages
	}

	return &result.Messages[0], nil
}

func (q *SQS) DeleteMessage(ctx context.Context, receiptHandle *string) error {
	_, err := q.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.taskURL),
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

func (q *SQS) SendMessage(ctx context.Context, body any) error {
	bytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	_, err = q.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.resultURL),
		MessageBody: aws.String(string(bytes)),
	})
	if err != nil {
		return fmt.Errorf("send failed: %w", err)
	}

	return nil
}
