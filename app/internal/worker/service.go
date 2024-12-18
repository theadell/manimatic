package worker

import (
	"bytes"
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
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
)

var errQueueNotExist = &types.QueueDoesNotExist{}

type WorkerService struct {
	config        *config.Config
	log           *slog.Logger
	sqsClient     *sqs.Client
	s3Client      *s3.Client
	s3Uploader    *manager.Uploader
	s3Presigner   *awsutils.S3Presigner
	tempDir       string
	workerPool    *WorkerPool
	cancelContext context.Context
	cancelFunc    context.CancelFunc
}

func NewWorkerService(cfg *config.Config) (*WorkerService, error) {

	log := logger.NewLogger(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("Failed to load AWS config", "error", err)
		cancel()
		return nil, err
	}

	sqsClient := awsutils.NewSQSClient(*cfg, awsConfig)
	s3Client := awsutils.NewS3Client(*cfg, awsConfig)
	s3Presigner := awsutils.NewS3PreSigner(s3Client, cfg.VideoBucketName)
	uploader := manager.NewUploader(s3Client)

	workerPool := NewWorkerPool(cfg.MaxConcurrency, log)

	return &WorkerService{
		config:        cfg,
		log:           log,
		sqsClient:     sqsClient,
		s3Client:      s3Client,
		s3Uploader:    uploader,
		s3Presigner:   s3Presigner,
		workerPool:    workerPool,
		cancelContext: ctx,
		cancelFunc:    cancel,
	}, nil
}

func (ws *WorkerService) Cleanup() {
	os.RemoveAll(ws.tempDir)
	ws.cancelFunc()
	ws.workerPool.Stop()
}

func (ws *WorkerService) Run() {

	ws.workerPool.Start(ws.processTask)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ws.cancelContext.Done():
				return
			default:
				ws.fetchAndSubmitMessage()
			}
		}
	}()

	<-sigChan
	ws.log.Info("Shutting down gracefully...")
	ws.Cleanup()
	ws.log.Info("Server shutdown")
	os.Exit(0)
}

func (ws *WorkerService) fetchAndSubmitMessage() {
	result, err := ws.sqsClient.ReceiveMessage(ws.cancelContext, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(ws.config.TaskQueueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
	})

	if err != nil {
		if errors.Is(err, context.Canceled) {
			ws.log.Debug("context cancelled - returning")
			return
		}
		ws.handleReceiveMessageError(err)
		return
	}

	if len(result.Messages) == 0 {
		return
	}

	msg := result.Messages[0]

	ws.log.Debug("Message received", "messageID", *msg.MessageId)

	ws.workerPool.Submit(Task{
		Message: &msg,
	})
}

func (ws *WorkerService) processTask(task Task) error {
	var message events.Message
	if err := json.Unmarshal([]byte(*task.Message.Body), &message); err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}

	outputPath, taskDir, err := ws.compileManimScript(message)
	if err != nil {
		return fmt.Errorf("compilation failed: %v", err)
	}

	defer os.RemoveAll(taskDir)

	_, err = ws.sqsClient.DeleteMessage(ws.cancelContext, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(ws.config.TaskQueueURL),
		ReceiptHandle: task.Message.ReceiptHandle,
	})
	if err != nil {
		return fmt.Errorf("message deletion failed: %v", err)
	}

	s3Key := ""
	if ws.config.VideoBucketName != "" {
		s3Key, err = ws.uploadVideoToS3(outputPath, message)
		if err != nil {
			return err
		}
	}
	go ws.presignAndEnqueueResult(s3Key, message.SessionId)

	return nil
}

func (ws *WorkerService) uploadVideoToS3(outputPath string, message events.Message) (string, error) {
	videoFile, err := os.Open(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to open video: %v", err)
	}
	defer videoFile.Close()

	s3Key := fmt.Sprintf("manim_videos/%s/%d.mp4", message.SessionId, time.Now().UnixNano())
	_, err = ws.s3Uploader.Upload(ws.cancelContext, &s3.PutObjectInput{
		Bucket: aws.String(ws.config.VideoBucketName),
		Key:    aws.String(s3Key),
		Body:   videoFile,
	})
	if err != nil {
		return "", fmt.Errorf("S3 upload failed: %v", err)
	}

	ws.log.Info("Video uploaded to S3", "bucket", ws.config.VideoBucketName, "key", s3Key)
	return s3Key, nil
}

func (ws *WorkerService) compileManimScript(msg events.Message) (string, string, error) {
	compilationID := uuid.New().String()
	contentStr, ok := msg.Content.(string)
	if !ok {
		return "", "", fmt.Errorf("invalid content type: %T", msg.Content)
	}

	taskDir, err := os.MkdirTemp(ws.tempDir, fmt.Sprintf("%s_%s", msg.SessionId, compilationID))
	if err != nil {
		return "", "", fmt.Errorf("failed to create task directory: %v", err)
	}

	scriptFile, err := os.CreateTemp(taskDir, "*.py")
	if err != nil {
		os.RemoveAll(taskDir)
		return "", "", fmt.Errorf("temp file creation failed: %v", err)
	}
	scriptFilePath := scriptFile.Name()

	if _, err := scriptFile.Write([]byte(contentStr)); err != nil {
		scriptFile.Close()
		os.RemoveAll(taskDir)
		return "", "", fmt.Errorf("script write failed: %v", err)
	}
	scriptFile.Close()

	outputVideoPath := filepath.Join(taskDir, "output.mp4")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "manim", "-qm", "-o", outputVideoPath, scriptFilePath)
	var stderr bytes.Buffer
	var stdOut bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Stdout = &stdOut

	if err := cmd.Run(); err != nil {
		os.RemoveAll(taskDir)
		return "", "", fmt.Errorf("manim execution failed: %v, stderr: %s", err, stderr.String())
	}

	return outputVideoPath, taskDir, nil
}

func (ws *WorkerService) presignAndEnqueueResult(key string, sessionId string) {
	req, err := ws.s3Presigner.PreSignGet(key, 3600)
	if err != nil {
		ws.log.Error("failed to presign the video url", "err", err)
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
		ws.log.Error("failed to marshal message body", "err", err)
		return
	}

	_, err = ws.sqsClient.SendMessage(ws.cancelContext, &sqs.SendMessageInput{
		QueueUrl:    aws.String(ws.config.ResultQueueURL),
		MessageBody: aws.String(string(bytes))})
	if err != nil {
		ws.log.Error("failed to send message", "err", err)
		return
	}
}

func (ws *WorkerService) handleReceiveMessageError(err error) {
	ws.log.Error("Failed to receive message", "error", err)
	if errors.As(err, &errQueueNotExist) {
		slog.Error("Queue does not exist", "URL", ws.config.TaskQueueURL, "Base endpoint", ws.sqsClient.Options().BaseEndpoint)
		os.Exit(1)
	}
}
