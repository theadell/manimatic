package main

import (
	"fmt"
	"manimatic/internal/config"
	"manimatic/internal/worker"
	"os"
)

func main() {

	cfg := config.LoadConfig()

	workerService, err := worker.NewWorkerService(cfg)
	if err != nil {
		fmt.Println("Failed to create worker service:", err)
		os.Exit(1)
	}
	workerService.Run()
}
