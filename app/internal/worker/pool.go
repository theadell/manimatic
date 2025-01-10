package worker

import (
	"log/slog"
	"manimatic/internal/api/events"
	"runtime/debug"
	"sync"
)

type Task struct {
	event          *events.Event
	compileRequest *events.CompileRequest
	h              *string
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

			// Protect the entire worker goroutine from panics
			defer func() {
				if r := recover(); r != nil {
					log.Error("worker routine panic recovered",
						"panic", r,
						"stack", string(debug.Stack()))
				}
			}()

			log.Debug("Worker start and ready to receive tasks")

			for task := range wp.tasks {
				// Protect individual task execution from panics
				func(t Task) {
					defer func() {
						if r := recover(); r != nil {
							log.Error("task execution panic recovered",
								"panic", r,
								"stack", string(debug.Stack()))
						}
					}()

					if err := processFunc(t); err != nil {
						log.Error("Task Processing Failed", "error", err)
					} else {
						log.Debug("finished processing message successfully")
					}
				}(task)
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
