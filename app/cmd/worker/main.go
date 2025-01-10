package main

import (
	"context"
	"fmt"
	"log"
	"manimatic/internal/awsutils"
	"manimatic/internal/config"
	"manimatic/internal/logger"
	"manimatic/internal/worker"
	"manimatic/internal/worker/animation"
	"manimatic/pkg/queue"
	"manimatic/pkg/storage"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config %s \n", err.Error())
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("failed to load AWS config", err)
	}
	log := logger.NewLogger(cfg)

	s3Client := awsutils.NewS3Client(*cfg, awsConfig)
	sqsClient := awsutils.NewSQSClient(*cfg, awsConfig)
	S3Storage := storage.NewS3(s3Client, cfg.AWS.VideoBucketName, log)
	msgQueue := queue.NewSQS(sqsClient, cfg.AWS.TaskQueueURL, cfg.AWS.ResultQueueURL, log)
	q := animation.NewQueue(msgQueue, log)
	workerService, err := worker.NewWorkerService(cfg, q, S3Storage, log)
	if err != nil {
		fmt.Println("Failed to create worker service:", err)
		os.Exit(1)
	}
	workerService.Run()
}
