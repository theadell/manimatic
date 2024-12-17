package worker

import (
	"log/slog"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Task struct {
	Message *types.Message
}

type WorkerPool struct {
	tasks       chan Task
	workerCount int
	log         *slog.Logger
	wg          sync.WaitGroup
}

func NewWorkerPool(workerCount int, log *slog.Logger) *WorkerPool {
	return &WorkerPool{
		tasks:       make(chan Task, workerCount),
		workerCount: workerCount,
		log:         log,
		wg:          sync.WaitGroup{},
	}
}

func (wp *WorkerPool) Start(processFunc func(Task) error) {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go func(workerN int) {
			defer wp.wg.Done()
			log := wp.log.With("Worker", workerN+1)
			log.Debug("Worker start and ready to receive tasks")
			for task := range wp.tasks {
				log.Debug("processing a new message", "message_id", task.Message.MessageId)
				time.Sleep(time.Second * 2)
				if err := processFunc(task); err != nil {
					log.Error("Task Processing Failed", "error", err)
				} else {
					log.Debug("finished processing message successfully", "message_id", task.Message.MessageId)
				}
			}
			log.Debug("Task channel closed. Exiting...")
		}(i)
	}
}

func (wp *WorkerPool) Submit(task Task) {
	wp.tasks <- task
}

func (wp *WorkerPool) Stop() {
	close(wp.tasks)
	wp.wg.Wait()
}
