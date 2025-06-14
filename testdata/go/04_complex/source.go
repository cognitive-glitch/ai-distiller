// worker.go
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Job struct{ ID int }

type Worker struct {
	id         int
	jobChannel <-chan Job
	wg         *sync.WaitGroup
}

func (w *Worker) start(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case job := <-w.jobChannel:
			fmt.Printf("Worker %d processing job %d\n", w.id, job.ID)
			time.Sleep(100 * time.Millisecond)
		case <-ctx.Done():
			fmt.Printf("Worker %d shutting down.\n", w.id)
			return
		}
	}
}

type Dispatcher struct {
	JobChannel chan Job
	maxWorkers int
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	return &Dispatcher{
		JobChannel: make(chan Job, 100),
		maxWorkers: maxWorkers,
	}
}

func (d *Dispatcher) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 1; i <= d.maxWorkers; i++ {
		wg.Add(1)
		worker := Worker{id: i, jobChannel: d.JobChannel, wg: &wg}
		go worker.start(ctx) // The key goroutine spawn
	}
	wg.Wait()
}