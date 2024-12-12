package config

import (
	"flag"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	OpenAIKey       string
	TaskQueueURL    string
	ResultQueueURL  string
	Environment     string
	IsLocal         bool
	Host            string
	Port            int
	LogLevel        slog.Level
	LogFormat       string
	MaxConcurrency  int
	VideoBucketName string
}

func LoadConfig() *Config {
	config := &Config{}

	config.Host = getEnvString("HOST", "localhost")
	config.Port = getEnvInt("PORT", 8080)
	config.Environment = getEnvString("MANIMATIC_ENVIRONMENT", "dev")
	config.OpenAIKey = getEnvString("OPENAI_API_KEY", "")
	config.TaskQueueURL = getEnvString("TASK_QUEUE_URL", "http://sqs.eu-central-1.localhost:4566/000000000000/manim-task-queue")
	config.ResultQueueURL = getEnvString("RESULT_QUEUE_URL", "http://sqs.eu-central-1.localhost:4566/000000000000/manim-result-queue")
	logLevelStr := getEnvString("LOG_LEVEL", "info")
	config.LogFormat = getEnvString("LOG_FORMAT", "text")
	config.LogLevel = parseLogLevel(logLevelStr)
	config.MaxConcurrency = getEnvInt("MAX_CONCURRENT_JOBS", 2)
	config.VideoBucketName = getEnvString("VIDEO_BUCKET_NAME", "manim-videos-bucket")

	flag.StringVar(&config.Host, "host", config.Host, "Server host")
	flag.IntVar(&config.Port, "port", config.Port, "Server port")
	flag.StringVar(&config.Environment, "env", config.Environment, "Environment (dev, staging, prod)")
	flag.StringVar(&config.OpenAIKey, "openai-api-key", config.OpenAIKey, "OpenAI API key")
	flag.StringVar(&config.TaskQueueURL, "task-queue-url", config.TaskQueueURL, "Task Queue URL")
	flag.StringVar(&config.ResultQueueURL, "result-queue-url", config.ResultQueueURL, "Result Queue URL")

	logLevelFlag := flag.String("log-level", logLevelStr, "Logging level (debug, info, warn, error)")
	flag.StringVar(&config.LogFormat, "log-format", config.LogFormat, "Logging format (text or json)")
	flag.IntVar(&config.MaxConcurrency, "max-concurrent", config.MaxConcurrency, "Max concurrent job processing")
	flag.StringVar(&config.VideoBucketName, "video-bucket-name", config.VideoBucketName, "S3 bucket for output videos")

	flag.Parse()

	if *logLevelFlag != logLevelStr {
		config.LogLevel = parseLogLevel(*logLevelFlag)
	}

	if config.Environment == "local" || config.Environment == "dev" || config.Environment == "development" {
		config.IsLocal = true
	}

	return config
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
