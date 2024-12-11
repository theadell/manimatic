package main

import (
	"context"
	"fmt"
	"log"
	"manimatic/internal/api"
	"manimatic/internal/api/genmanim"
	"manimatic/internal/config"
	"manimatic/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	cfg := config.LoadConfig()

	logger := logger.NewLogger(cfg.Env)
	manimService, err := genmanim.NewLLMManimService(cfg.OpenAIKey)
	if err != nil {
		log.Fatal(err)
	}
	api := api.New(cfg, logger, manimService)

	server := http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           api,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 2,
		WriteTimeout:      time.Second * 5,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info(fmt.Sprintf("Server is running on port %d", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "err", err.Error())
			os.Exit(1)
		}
	}()

	<-stop
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "err", err.Error())
	}

	logger.Info("Server has shut down gracefully")
}
