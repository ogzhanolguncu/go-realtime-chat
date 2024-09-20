package threadpool

// This is experimentally add to practice threadpools, currently used to block number of connections to server.

import (
	"sync"
	"sync/atomic"
	"time"
)

// Task represents a function to be executed by a worker
type Task func()

// Threadpool manages a pool of workers
type Threadpool struct {
	workers   []*Worker
	taskQueue chan Task
	wg        sync.WaitGroup
	running   atomic.Bool
}

// Worker represents a goroutine that executes tasks
type Worker struct {
	id        int
	taskQueue chan Task
	quit      chan bool
}

// Create a new threadpool with the given number of workers
func NewThreadpool(numWorkers int) *Threadpool {
	tp := &Threadpool{
		workers:   make([]*Worker, numWorkers),
		taskQueue: make(chan Task, numWorkers*2),
	}
	tp.running.Store(true)

	for i := 0; i < numWorkers; i++ {
		worker := &Worker{
			id:        i,
			taskQueue: tp.taskQueue,
			quit:      make(chan bool),
		}
		tp.workers[i] = worker
		tp.wg.Add(1)
		go worker.Start(&tp.wg)
	}
	return tp
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-w.taskQueue:
			if !ok {
				return // taskQueue was closed
			}
			if task != nil {
				task()
			}
		case <-w.quit:
			return
		}
	}
}

// Adds a task to the threadpool
func (tp *Threadpool) Submit(task Task) {
	if tp.running.Load() {
		tp.taskQueue <- task
	}
}

func (tp *Threadpool) SubmitWithTimeout(task Task, timeout time.Duration) bool {
	if !tp.running.Load() {
		return false
	}
	select {
	case tp.taskQueue <- task:
		return true
	case <-time.After(timeout):
		return false
	}
}

// Stops all workers and waits for them to finish
func (tp *Threadpool) Stop() {
	tp.running.Store(false)
	close(tp.taskQueue)
	for _, worker := range tp.workers {
		close(worker.quit)
	}
	tp.wg.Wait()
}
