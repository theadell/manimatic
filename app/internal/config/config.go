package config

import (
	"flag"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	OpenAIKey      string
	SQSTaskURL     string
	SQSResultURL   string
	Env            string
	EnvLocal       bool
	Host           string
	Port           int
	LogLevel       slog.Level
	LogFormat      string
	MaxConcurrency int
	S3Bucket       string
}

func LoadConfig() *Config {
	config := &Config{}

	config.Host = getEnvString("HOST", "localhost")
	config.Port = getEnvInt("PORT", 8080)
	config.Env = getEnvString("MANIMATIC_ENV", "dev")
	config.OpenAIKey = getEnvString("OPENAI_API_KEY", "")
	config.SQSTaskURL = getEnvString("SQS_TASK_QUEUE_URL", "http://sqs.eu-central-1.localhost:4566/000000000000/manim-task-queue")
	config.SQSResultURL = getEnvString("SQS_RESULT_QUEUE_URL", "http://sqs.eu-central-1.localhost:4566/000000000000/manim-result-queue")
	logLevelStr := getEnvString("LOG_LEVEL", "info")
	config.LogFormat = getEnvString("LOG_FORMAT", "text")
	config.LogLevel = parseLogLevel(logLevelStr)
	config.MaxConcurrency = getEnvInt("MAX_CONCURRENT_JOBS", 2)
	config.S3Bucket = getEnvString("S3_BUCKET", "manim-worker-bucket")

	flag.StringVar(&config.Host, "host", config.Host, "Server host")
	flag.IntVar(&config.Port, "port", config.Port, "Server port")
	flag.StringVar(&config.Env, "env", config.Env, "Environment (dev, staging, prod)")
	flag.StringVar(&config.OpenAIKey, "openai-api-key", config.OpenAIKey, "OpenAI API key")
	flag.StringVar(&config.SQSTaskURL, "sqs-task-url", config.SQSTaskURL, "SQS Tasks Queue URL")
	logLevelFlag := flag.String("log-level", logLevelStr, "Logging level (debug, info, warn, error)")
	flag.StringVar(&config.LogFormat, "log-format", config.LogFormat, "Logging format (text or json)")
	flag.IntVar(&config.MaxConcurrency, "max-concurrent", config.MaxConcurrency, "Max concurrent job processing")
	flag.StringVar(&config.S3Bucket, "s3-bucket", config.S3Bucket, "S3 bucket for output videos")

	flag.Parse()

	if *logLevelFlag != logLevelStr {
		config.LogLevel = parseLogLevel(*logLevelFlag)
	}

	if config.Env == "local" || config.Env == "dev" || config.Env == "development" {
		config.EnvLocal = true
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
