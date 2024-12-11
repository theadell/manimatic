package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	OpenAIKey string
	SQSURL    string
	Env       string
	Host      string
	Port      int
}

func LoadConfig() *Config {
	config := &Config{}

	// Load from environment variables with defaults
	config.Host = getEnvString("HOST", "localhost")
	config.Port = getEnvInt("PORT", 8080)
	config.Env = getEnvString("MANIMATIC_ENV", "dev")
	config.OpenAIKey = getEnvString("OPENAI_API_KEY", "")
	config.SQSURL = getEnvString("SQS_QUEUE_URL", "")

	// Define flags
	flag.StringVar(&config.Host, "host", config.Host, "Server host")
	flag.IntVar(&config.Port, "port", config.Port, "Server port")
	flag.StringVar(&config.Env, "env", config.Env, "Environment (dev, staging, prod)")
	flag.StringVar(&config.OpenAIKey, "openai-api-key", config.OpenAIKey, "OpenAI API key")
	flag.StringVar(&config.SQSURL, "sqs-url", config.SQSURL, "SQS Queue URL")

	// Parse flags
	flag.Parse()

	return config
}

// getEnvString retrieves a string value from the environment or returns a default.
func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an int value from the environment or returns a default.
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
