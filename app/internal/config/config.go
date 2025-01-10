package config

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"manimatic/internal/api/features"
	"os"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type ServerConfig struct {
	Host string
	Port int
}

type AWSConfig struct {
	EndpointURL     string
	TaskQueueURL    string
	ResultQueueURL  string
	VideoBucketName string
}

type LogConfig struct {
	Level  slog.Level
	Format string
}

type ProcessingConfig struct {
	MaxConcurrency   int
	EnableModeration bool
	FeaturesFlag     string
	Features         *features.Features
}

type APIKeyConfig struct {
	Key        string
	keyFile    string
	keySSMPath string
	IsSet      bool
}

type WorkerMediaConfig struct {
	BaseDir string
}

type Config struct {
	Server     ServerConfig
	AWS        AWSConfig
	Logging    LogConfig
	Processing ProcessingConfig
	OpenAI     APIKeyConfig
	XAI        APIKeyConfig
	Worker     WorkerMediaConfig
}

func (c *Config) registerServerConfig(r *Register) {
	r.String(&c.Server.Host, "HOST", "Server host", "0.0.0.0")
	r.Int(&c.Server.Port, "PORT", "Server port", 8080)
}

func (c *Config) registerAWSConfig(r *Register) {
	r.String(&c.AWS.EndpointURL, "AWS_ENDPOINT_URL", "Custom AWS endpoint URL", "")
	r.String(&c.AWS.TaskQueueURL, "TASK_QUEUE_URL", "Task Queue URL",
		"http://sqs.eu-central-1.localhost:4566/000000000000/manim-task-queue")
	r.String(&c.AWS.ResultQueueURL, "RESULT_QUEUE_URL", "Result Queue URL",
		"http://sqs.eu-central-1.localhost:4566/000000000000/manim-result-queue")
	r.String(&c.AWS.VideoBucketName, "VIDEO_BUCKET_NAME", "S3 bucket for output videos", "manim-worker-bucket")
}

func (c *Config) registerLoggingConfig(r *Register) {
	r.String(&c.Logging.Format, "LOG_FORMAT", "Logging format (text or json)", "text")

	logLevel := "info"
	r.String(&logLevel, "LOG_LEVEL", "Log level (debug, info, warn, error)", "info")
	c.Logging.Level = parseLogLevel(logLevel)
}

func (c *Config) registerProcessingConfig(r *Register) {
	r.Int(&c.Processing.MaxConcurrency, "MAX_CONCURRENCY", "Max concurrent job processing", runtime.NumCPU())
	r.Bool(&c.Processing.EnableModeration, "ENABLE_MODERATION", "Use the OpenAI moderation endpoint", false)
	r.String(&c.Processing.FeaturesFlag, "FEATURES", "Comma-separated list of features to enable", "")
}

func (c *Config) registerAPIKeys(r *Register) {
	// OpenAI
	r.String(&c.OpenAI.Key, "OPENAI_API_KEY", "OpenAI API key", "")
	r.String(&c.OpenAI.keyFile, "OPENAI_API_KEY_FILE", "Path to Docker secret file containing OpenAI key", "")
	r.String(&c.OpenAI.keySSMPath, "OPENAI_API_KEY_SSM_PATH", "AWS SSM Parameter Store path for OpenAI key", "")

	// XAI
	r.String(&c.XAI.Key, "XAI_API_KEY", "XAI API key", "")
	r.String(&c.XAI.keyFile, "XAI_API_KEY_FILE", "Path to Docker secret file containing XAI key", "")
	r.String(&c.XAI.keySSMPath, "XAI_API_KEY_SSM_PATH", "AWS SSM Parameter Store path for XAI key", "")
}

func (c *Config) registerWorkerConfig(r *Register) {
	r.String(&c.Worker.BaseDir, "WORKER_DIR", "Directory for worker temporary files", os.TempDir())
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	r := &Register{}

	// Register all config groups
	config.registerServerConfig(r)
	config.registerAWSConfig(r)
	config.registerLoggingConfig(r)
	config.registerProcessingConfig(r)
	config.registerAPIKeys(r)
	config.registerWorkerConfig(r)

	flag.Parse()

	// Load API keys
	if err := config.loadAPIKeys(); err != nil {
		fmt.Println(err.Error())
	}

	// Validate
	if err := config.validate(); err != nil {
		return nil, err
	}

	if lvl := os.Getenv("LOG_LEVEL"); strings.ToLower(lvl) == "debug" {
		fmt.Print(config.Debug())
	}

	config.Processing.Features = features.New(config.Processing.FeaturesFlag)

	return config, nil
}
func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
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

func (c *Config) loadAPIKeys() error {
	var loadErrors []error

	if err := c.loadKey(&c.OpenAI, "OpenAI"); err != nil {
		loadErrors = append(loadErrors, err)
	}

	if err := c.loadKey(&c.XAI, "XAI"); err != nil {
		loadErrors = append(loadErrors, err)
	}

	if !c.OpenAI.IsSet && !c.XAI.IsSet {
		return fmt.Errorf("no valid API keys provided. Errors: %v", loadErrors)
	}

	return nil
}
func (c *Config) loadKey(keyConfig *APIKeyConfig, keyName string) error {
	if keyConfig.keyFile != "" {
		key, err := readSecretFile(keyConfig.keyFile)
		if err != nil {
			return fmt.Errorf("failed to read %s key from Docker secret: %w", keyName, err)
		}
		keyConfig.Key = key
		keyConfig.IsSet = true
		return nil
	}

	if keyConfig.keySSMPath != "" {
		key, err := readAWSParameter(keyConfig.keySSMPath)
		if err != nil {
			return fmt.Errorf("failed to read %s key from AWS Parameter Store: %w", keyName, err)
		}
		keyConfig.Key = key
		keyConfig.IsSet = true
		return nil
	}

	if keyConfig.Key != "" {
		keyConfig.IsSet = true
	}

	return nil
}

func (c *Config) validate() error {
	// Server validation
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Server.Port)
	}

	// Processing validation
	if c.Processing.MaxConcurrency <= 0 || c.Processing.MaxConcurrency > 20 {
		c.Processing.MaxConcurrency = runtime.NumCPU()
	}

	// AWS validation
	if c.AWS.TaskQueueURL == "" {
		return fmt.Errorf("task queue URL is required")
	}
	if c.AWS.ResultQueueURL == "" {
		return fmt.Errorf("result queue URL is required")
	}
	if c.AWS.VideoBucketName == "" {
		return fmt.Errorf("video bucket name is required")
	}

	// Log format validation
	if c.Logging.Format != "text" && c.Logging.Format != "json" {
		c.Logging.Format = "json"
	}

	return nil
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
		Name:           aws.String(path),
		WithDecryption: aws.Bool(true),
	}

	result, err := client.GetParameter(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get AWS SSM parameter: %w", err)
	}

	return *result.Parameter.Value, nil
}

func (c *Config) Debug() string {
	var b strings.Builder

	b.WriteString("\n=== Configuration ===\n\n")

	// Server Config
	b.WriteString("🖥️  Server:\n")
	b.WriteString(fmt.Sprintf("  ├─ Host: %s\n", c.Server.Host))
	b.WriteString(fmt.Sprintf("  └─ Port: %d\n\n", c.Server.Port))

	// AWS Config
	b.WriteString("☁️  AWS:\n")
	b.WriteString(fmt.Sprintf("  ├─ Endpoint URL: %s\n", valueOrEmpty(c.AWS.EndpointURL)))
	b.WriteString(fmt.Sprintf("  ├─ Task Queue URL: %s\n", c.AWS.TaskQueueURL))
	b.WriteString(fmt.Sprintf("  ├─ Result Queue URL: %s\n", c.AWS.ResultQueueURL))
	b.WriteString(fmt.Sprintf("  └─ Video Bucket: %s\n\n", c.AWS.VideoBucketName))

	// Logging Config
	b.WriteString("📝 Logging:\n")
	b.WriteString(fmt.Sprintf("  ├─ Level: %s\n", c.Logging.Level))
	b.WriteString(fmt.Sprintf("  └─ Format: %s\n\n", c.Logging.Format))

	// Processing Config
	b.WriteString("⚙️  Processing:\n")
	b.WriteString(fmt.Sprintf("  ├─ Max Concurrency: %d\n", c.Processing.MaxConcurrency))
	b.WriteString(fmt.Sprintf("  ├─ Moderation Enabled: %v\n", c.Processing.EnableModeration))
	b.WriteString(fmt.Sprintf("  └─ Base Dir: %s\n", valueOrEmpty(c.Worker.BaseDir)))
	b.WriteString(fmt.Sprintf("  └─ Features: %s\n\n", valueOrEmpty(c.Processing.FeaturesFlag)))

	// API Keys (safely)
	b.WriteString("🔑 API Keys:\n")
	b.WriteString(fmt.Sprintf("  ├─ OpenAI:\n"))
	b.WriteString(fmt.Sprintf("  │  ├─ Key Set: %v\n", c.OpenAI.IsSet))
	b.WriteString(fmt.Sprintf("  │  ├─ Key File: %s\n", valueOrEmpty(c.OpenAI.keyFile)))
	b.WriteString(fmt.Sprintf("  │  └─ SSM Path: %s\n", valueOrEmpty(c.OpenAI.keySSMPath)))
	b.WriteString(fmt.Sprintf("  └─ XAI:\n"))
	b.WriteString(fmt.Sprintf("     ├─ Key Set: %v\n", c.XAI.IsSet))
	b.WriteString(fmt.Sprintf("     ├─ Key File: %s\n", valueOrEmpty(c.XAI.keyFile)))
	b.WriteString(fmt.Sprintf("     └─ SSM Path: %s\n", valueOrEmpty(c.XAI.keySSMPath)))

	return b.String()
}

// valueOrEmpty returns "<not set>" for empty strings
func valueOrEmpty(s string) string {
	if s == "" {
		return "<not set>"
	}
	return s
}
