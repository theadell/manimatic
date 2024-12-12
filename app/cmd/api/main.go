package main

import (
	"context"
	"fmt"
	"log"
	"manimatic/internal/api"
	"manimatic/internal/api/genmanim"
	"manimatic/internal/awsutils"
	"manimatic/internal/config"
	"manimatic/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	cfg := config.LoadConfig()

	logger := logger.NewLogger(cfg)
	manimService, err := genmanim.NewLLMManimService(cfg.OpenAIKey)
	if err != nil {
		log.Fatal(err)
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	sqsClient := awsutils.NewSQSClient(*cfg, awsConfig)
	api := api.New(cfg, logger, manimService, sqsClient)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	api.StartMessageProcessor(ctx)

	server := http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           api,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 2,
		WriteTimeout:      time.Second * 5,
	}

	go func() {
		logger.Info(fmt.Sprintf("Server is running on port %d", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "err", err.Error())
			stop()
		}
	}()

	// Wait for stop signal
	<-ctx.Done()
	logger.Info("Shutting down server...")

	api.MsgRouter.Shutdown()

	// Create a shutdown context with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "err", err.Error())
	}

	logger.Info("Server has shut down gracefully")
}
