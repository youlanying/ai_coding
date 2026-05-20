package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"task-scheduler/internal/models"
	"task-scheduler/internal/store"
)

type TaskExecutor func(ctx context.Context, payload map[string]string) (string, error)

type WorkerPool struct {
	store          store.Store
	executor       TaskExecutor
	taskQueue      chan *models.Task
	workers        map[int]context.CancelFunc
	workerMu       sync.Mutex
	workerCount    int32
	activeCount    int32
	nextWorkerID   int
	OnTaskComplete func(taskID string)
}

func NewWorkerPool(st store.Store, executor TaskExecutor, initialWorkers int) *WorkerPool {
	pool := &WorkerPool{
		store:       st,
		executor:    executor,
		taskQueue:   make(chan *models.Task, 10000),
		workers:     make(map[int]context.CancelFunc),
		workerCount: 0,
		activeCount: 0,
	}
	pool.SetWorkerCount(initialWorkers)
	return pool
}

func (p *WorkerPool) SetWorkerCount(count int) {
	if count < 1 {
		count = 1
	}
	p.workerMu.Lock()
	defer p.workerMu.Unlock()

	current := int(p.workerCount)
	if count > current {
		for i := 0; i < count-current; i++ {
			p.startWorker()
		}
	} else if count < current {
		stopCount := current - count
		counter := 0
		for id, cancel := range p.workers {
			if counter >= stopCount {
				break
			}
			cancel()
			delete(p.workers, id)
			counter++
		}
		atomic.StoreInt32(&p.workerCount, int32(count))
	}
}

func (p *WorkerPool) startWorker() {
	p.nextWorkerID++
	id := p.nextWorkerID
	ctx, cancel := context.WithCancel(context.Background())
	p.workers[id] = cancel
	atomic.AddInt32(&p.workerCount, 1)

	go func() {
		defer atomic.AddInt32(&p.workerCount, -1)
		for {
			select {
			case <-ctx.Done():
				return
			case task, ok := <-p.taskQueue:
				if !ok {
					return
				}
				atomic.AddInt32(&p.activeCount, 1)
				p.processTask(task)
				atomic.AddInt32(&p.activeCount, -1)
			}
		}
	}()
}

func (p *WorkerPool) processTask(task *models.Task) {
	startTime := time.Now()
	p.store.UpdateTaskStatus(task.ID, models.TaskStatusRunning)

	ctx := context.Background()
	result, err := p.executor(ctx, task.Payload)

	duration := time.Since(startTime).Milliseconds()

	taskResult := &models.TaskResult{
		TaskID:     task.ID,
		RetryCount: task.RetryCount,
		ExecutedAt: startTime,
		DurationMs: duration,
	}

	if err != nil {
		taskResult.Status = models.TaskStatusFailed
		taskResult.Error = err.Error()
		p.store.UpdateTaskResult(task.ID, "", err.Error())

		if task.RetryCount < task.MaxRetry {
			p.store.IncrementRetry(task.ID)
			delay := p.calculateBackoff(task.RetryCount)
			p.store.UpdateTaskStatus(task.ID, models.TaskStatusPending)
			task.ScheduledAt = time.Now().Add(time.Duration(delay) * time.Millisecond)
			p.store.SaveTask(task)
			go func(t *models.Task, d int64) {
				time.Sleep(time.Duration(d) * time.Millisecond)
				p.Submit(t)
			}(task, delay)
		} else {
			p.store.UpdateTaskStatus(task.ID, models.TaskStatusFailed)
			if p.OnTaskComplete != nil {
				p.OnTaskComplete(task.ID)
			}
		}
	} else {
		taskResult.Status = models.TaskStatusSuccess
		taskResult.Result = result
		p.store.UpdateTaskResult(task.ID, result, "")
		p.store.UpdateTaskStatus(task.ID, models.TaskStatusSuccess)
		if p.OnTaskComplete != nil {
			p.OnTaskComplete(task.ID)
		}
	}

	p.store.SaveResult(taskResult)
}

func (p *WorkerPool) calculateBackoff(retryCount int) int64 {
	base := 100
	factor := 2
	maxBackoff := 30000
	backoff := base * (1 << uint(retryCount*factor))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	return int64(backoff)
}

func (p *WorkerPool) Submit(task *models.Task) {
	p.taskQueue <- task
}

func (p *WorkerPool) GetStats() models.WorkerPoolStats {
	return models.WorkerPoolStats{
		ActiveWorkers: int(atomic.LoadInt32(&p.activeCount)),
		IdleWorkers:   int(atomic.LoadInt32(&p.workerCount) - atomic.LoadInt32(&p.activeCount)),
		TotalWorkers:  int(atomic.LoadInt32(&p.workerCount)),
		PendingTasks:  len(p.taskQueue),
	}
}

func (p *WorkerPool) Shutdown() {
	p.workerMu.Lock()
	defer p.workerMu.Unlock()
	for id, cancel := range p.workers {
		cancel()
		delete(p.workers, id)
	}
	close(p.taskQueue)
}

func DefaultExecutor(ctx context.Context, payload map[string]string) (string, error) {
	if payload == nil {
		return "executed", nil
	}
	if action, ok := payload["action"]; ok && action == "fail" {
		return "", fmt.Errorf("simulated failure")
	}
	if delayStr, ok := payload["delay_ms"]; ok {
		if delay, err := time.ParseDuration(delayStr + "ms"); err == nil {
			time.Sleep(delay)
		}
	}
	return fmt.Sprintf("executed with payload: %v", payload), nil
}
