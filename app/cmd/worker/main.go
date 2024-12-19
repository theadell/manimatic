package main

import (
	"fmt"
	"log"
	"manimatic/internal/config"
	"manimatic/internal/worker"
	"os"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config %s \n", err.Error())
	}

	workerService, err := worker.NewWorkerService(cfg)
	if err != nil {
		fmt.Println("Failed to create worker service:", err)
		os.Exit(1)
	}
	workerService.Run()
}
