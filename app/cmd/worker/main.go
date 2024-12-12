package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"manimatic/internal/api/events"
	"manimatic/internal/awsutils"
	"manimatic/internal/config"
	"manimatic/internal/logger"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

const (
	maxConcurrentCompilations = 2
	maxRetries                = 3
	tempDirPrefix             = "manim_scripts_"
)

var errQueueNotExist = &types.QueueDoesNotExist{}

func main() {
	// Setup configuration and logger
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg)

	//  Cancellation context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("Failed to load AWS config", "error", err)
		return
	}

	sqsClient := awsutils.NewSQSClient(*cfg, awsConfig)
	s3Client := awsutils.NewS3Client(*cfg, awsConfig)
	s3Presigner := awsutils.NewS3PreSigner(s3Client, cfg.VideoBucketName)

	uploader := manager.NewUploader(s3Client)

	// Temp directory to store the compiled videos
	tempDir, err := os.MkdirTemp("", tempDirPrefix)
	if err != nil {
		log.Error("Failed to create temp directory", "error", err)
		return
	}
	defer os.RemoveAll(tempDir)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrentCompilations)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				processNextMessage(ctx, log, sqsClient, uploader, cfg.VideoBucketName, s3Presigner, cfg.TaskQueueURL, cfg.ResultQueueURL, tempDir, &wg, semaphore)
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Info("Shutting down gracefully...")
	cancel()
	wg.Wait()
	log.Info("Server shutdown")
}

func processNextMessage(
	ctx context.Context,
	log *slog.Logger,
	client *sqs.Client,
	s3Uploader *manager.Uploader,
	s3Bucket string,
	presigner *awsutils.S3Presigner,
	queueURL string,
	vidoeQueueUrl string,
	tempDir string,
	wg *sync.WaitGroup,
	semaphore chan struct{},
) {
	result, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
	})

	if err != nil {
		log.Error("Failed to receive message", "error", err)
		if errors.As(err, &errQueueNotExist) {
			slog.Error("Queue does not exist", "URL", queueURL, "Base endpoint", client.Options().BaseEndpoint)
			os.Exit(1)
		}
		return
	}

	if len(result.Messages) == 0 {
		return
	}

	msg := result.Messages[0]
	wg.Add(1)

	log.Debug("Message received",
		"messageID", *msg.MessageId,
		"semaphoreAvailable", len(semaphore))

	semaphore <- struct{}{}

	go func(msg *types.Message) {
		defer wg.Done()
		defer func() { <-semaphore }()

		if err := processMessage(ctx, log, client, s3Uploader, s3Bucket, presigner, queueURL, vidoeQueueUrl, msg, tempDir); err != nil {
			log.Error("Message processing failed", "error", err)
		}
	}(&msg)
}

func processMessage(
	ctx context.Context,
	log *slog.Logger,
	client *sqs.Client,
	s3Uploader *manager.Uploader,
	s3Bucket string,
	presigner *awsutils.S3Presigner,
	queueURL string,
	vidoeQueueUrl string,
	msg *types.Message,
	tempDir string,
) error {
	var message events.Message
	if err := json.Unmarshal([]byte(*msg.Body), &message); err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}

	outputPath, err := compileManimScript(message, tempDir)
	if err != nil {
		return fmt.Errorf("compilation failed: %v", err)
	}

	// Delete message after successful processing
	_, err = client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		return fmt.Errorf("message deletion failed: %v", err)
	}

	s3Key := ""
	if s3Bucket != "" {
		videoFile, err := os.Open(outputPath)
		if err != nil {
			return fmt.Errorf("failed to open video: %v", err)
		}
		defer videoFile.Close()

		s3Key = fmt.Sprintf("manim_videos/%s.mp4", message.SessionId)
		_, err = s3Uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(s3Key),
			Body:   videoFile,
		})
		if err != nil {
			return fmt.Errorf("S3 upload failed: %v", err)
		}

		log.Info("Video uploaded to S3", "bucket", s3Bucket, "key", s3Key)
	}

	log.Info("Message processed successfully", "output", outputPath)

	go presignAndEnqueueResult(ctx, log, s3Key, presigner, vidoeQueueUrl, client, message.SessionId)
	return nil
}

func compileManimScript(msg events.Message, tempDir string) (string, error) {
	contentStr, ok := msg.Content.(string)
	if !ok {
		return "", fmt.Errorf("invalid content type: %T", msg.Content)
	}

	// Create secure temp file
	tempFile, err := os.CreateTemp(tempDir, fmt.Sprintf("%s_*.py", msg.SessionId))
	if err != nil {
		return "", fmt.Errorf("temp file creation failed: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(contentStr)); err != nil {
		return "", fmt.Errorf("script write failed: %v", err)
	}
	tempFile.Close()

	// Specify output video path
	outputVideoPath := filepath.Join(tempDir, fmt.Sprintf("%s_output.mp4", msg.SessionId))

	// Execute Manim command
	cmd := exec.Command("manim",
		"-qm",
		"-o", outputVideoPath,
		tempFile.Name(),
	)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("manim execution failed: %v", err)
	}

	return outputVideoPath, nil
}

func presignAndEnqueueResult(
	ctx context.Context,
	logger *slog.Logger,
	key string,
	presigner *awsutils.S3Presigner,
	queueURL string,
	sqsClient *sqs.Client,
	sessionId string) {
	req, err := presigner.PreSignGet(key, 3600)
	if err != nil {
		logger.Error("failed to presign the video url", "err", err)
		return
	}
	updateMsg := events.Message{
		Type:      events.MessageTypeVideoUpdate,
		SessionId: sessionId,
		Status:    events.MessageStatusSuccess,
		Content:   req.URL,
		Details: map[string]any{
			"header": req.SignedHeader,
		},
	}

	bytes, err := json.Marshal(updateMsg)
	if err != nil {
		logger.Error("failed to unmarshal message body", "err", err)
		return
	}

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(bytes))})
	if err != nil {
		logger.Error("failed to send message", "err", err)
		return
	}
	logger.Debug("succefully sent video message to the queue")
}
