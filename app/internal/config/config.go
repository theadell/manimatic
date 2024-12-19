package config

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Config struct {
	OpenAIKey       string
	TaskQueueURL    string
	ResultQueueURL  string
	UseLocalStack   bool
	AWSEndpointURL  string
	Host            string
	Port            int
	LogLevel        slog.Level
	LogFormat       string
	MaxConcurrency  int
	VideoBucketName string

	OpenAIKeyFile    string // Path to Docker secret file
	OpenAIKeySSMPath string // Path in AWS Parameter Store
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	config.Host = getEnvString("HOST", "0.0.0.0")
	config.Port = getEnvInt("PORT", 8080)
	config.UseLocalStack = getEnvBool("LOCALSTACK")
	config.AWSEndpointURL = getEnvString("AWS_ENDPOINT_URL", "")
	config.OpenAIKey = getEnvString("OPENAI_API_KEY", "")
	config.OpenAIKeyFile = getEnvString("OPENAI_API_KEY_FILE", "")
	config.OpenAIKeySSMPath = getEnvString("OPENAI_API_KEY_SSM_PATH", "")
	config.TaskQueueURL = getEnvString("TASK_QUEUE_URL", "http://sqs.eu-central-1.localhost:4566/000000000000/manim-task-queue")
	config.ResultQueueURL = getEnvString("RESULT_QUEUE_URL", "http://sqs.eu-central-1.localhost:4566/000000000000/manim-result-queue")
	logLevelStr := getEnvString("LOG_LEVEL", "info")
	config.LogFormat = getEnvString("LOG_FORMAT", "text")
	config.LogLevel = parseLogLevel(logLevelStr)
	config.MaxConcurrency = getEnvInt("MAX_CONCURRENCY", runtime.NumCPU())
	if config.MaxConcurrency <= 0 || config.MaxConcurrency > 20 {
		config.MaxConcurrency = runtime.NumCPU()
	}
	config.VideoBucketName = getEnvString("VIDEO_BUCKET_NAME", "manim-worker-bucket")

	flag.StringVar(&config.Host, "host", config.Host, "Server host")
	flag.IntVar(&config.Port, "port", config.Port, "Server port")
	flag.BoolVar(&config.UseLocalStack, "localstack", config.UseLocalStack, "Use localstack")
	flag.StringVar(&config.AWSEndpointURL, "aws-endpoint-url", config.AWSEndpointURL, "Custom AWS endpoint URL (e.g., for LocalStack or other mock services)")
	flag.StringVar(&config.OpenAIKey, "openai-api-key", config.OpenAIKey, "OpenAI API key")
	flag.StringVar(&config.OpenAIKeyFile, "openai-api-key-file", config.OpenAIKeyFile, "Path to Docker secret file containing OpenAI key")
	flag.StringVar(&config.OpenAIKeySSMPath, "openai-api-key-ssm-path", config.OpenAIKeySSMPath, "AWS SSM Parameter Store path for OpenAI key")
	flag.StringVar(&config.TaskQueueURL, "task-queue-url", config.TaskQueueURL, "Task Queue URL")
	flag.StringVar(&config.ResultQueueURL, "result-queue-url", config.ResultQueueURL, "Result Queue URL")

	logLevelFlag := flag.String("log-level", logLevelStr, "Logging level (debug, info, warn, error)")
	flag.StringVar(&config.LogFormat, "log-format", config.LogFormat, "Logging format (text or json)")
	flag.IntVar(&config.MaxConcurrency, "max-concurrency", config.MaxConcurrency, "Max concurrent job processing")
	flag.StringVar(&config.VideoBucketName, "video-bucket-name", config.VideoBucketName, "S3 bucket for output videos")

	flag.Parse()

	if *logLevelFlag != logLevelStr {
		config.LogLevel = parseLogLevel(*logLevelFlag)
	}

	var err error
	if config.OpenAIKeyFile != "" {
		config.OpenAIKey, err = readSecretFile(config.OpenAIKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read OpenAI key from Docker secret: %w", err)
		}
	} else if config.OpenAIKeySSMPath != "" && !config.UseLocalStack {
		config.OpenAIKey, err = readAWSParameter(config.OpenAIKeySSMPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read OpenAI key from AWS Parameter Store: %w", err)
		}
	}

	return config, nil
}

func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseLogLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnvBool(key string) bool {
	vString := getEnvString(key, "false")
	val, err := strconv.ParseBool(vString)
	if err != nil {
		return false
	}
	return val
}

func readSecretFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read secret file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func readAWSParameter(path string) (string, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := ssm.NewFromConfig(cfg)
	input := &ssm.GetParameterInput{
		Name:           &path,
		WithDecryption: aws.Bool(true),
	}

	result, err := client.GetParameter(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get AWS SSM parameter: %w", err)
	}

	return *result.Parameter.Value, nil
}
