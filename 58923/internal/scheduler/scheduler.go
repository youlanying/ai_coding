package scheduler

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"task-scheduler/internal/models"
	"task-scheduler/internal/store"
	"task-scheduler/internal/worker"
)

type Scheduler struct {
	store      *store.TaskStore
	workerPool *worker.WorkerPool
	cron       *cron.Cron
	cronJobs   map[string]cron.EntryID
	cronMu     sync.Mutex
	stopChan   chan struct{}
}

func NewScheduler(store *store.TaskStore, workerPool *worker.WorkerPool) *Scheduler {
	s := &Scheduler{
		store:      store,
		workerPool: workerPool,
		cron:       cron.New(cron.WithSeconds()),
		cronJobs:   make(map[string]cron.EntryID),
		stopChan:   make(chan struct{}),
	}
	s.cron.Start()
	go s.dispatchLoop()
	return s
}

func (s *Scheduler) generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Scheduler) AddDelayTask(name string, payload map[string]string, delayMs int64, maxRetry int) *models.Task {
	task := &models.Task{
		ID:          s.generateID(),
		Type:        models.TaskTypeDelay,
		Name:        name,
		Payload:     payload,
		DelayMs:     delayMs,
		Status:      models.TaskStatusPending,
		RetryCount:  0,
		MaxRetry:    maxRetry,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now().Add(time.Duration(delayMs) * time.Millisecond),
	}
	s.store.SaveTask(task)
	return task
}

func (s *Scheduler) AddCronTask(name string, payload map[string]string, cronExpr string, maxRetry int) (*models.Task, error) {
	task := &models.Task{
		ID:          s.generateID(),
		Type:        models.TaskTypeCron,
		Name:        name,
		Payload:     payload,
		CronExpr:    cronExpr,
		Status:      models.TaskStatusPending,
		RetryCount:  0,
		MaxRetry:    maxRetry,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
	}

	s.cronMu.Lock()
	defer s.cronMu.Unlock()

	entryID, err := s.cron.AddFunc(cronExpr, func() {
		taskCopy := *task
		taskCopy.ID = s.generateID()
		taskCopy.Status = models.TaskStatusPending
		taskCopy.RetryCount = 0
		taskCopy.CreatedAt = time.Now()
		taskCopy.ScheduledAt = time.Now()
		s.store.SaveTask(&taskCopy)
		s.workerPool.Submit(&taskCopy)
	})
	if err != nil {
		return nil, err
	}

	s.cronJobs[task.ID] = entryID
	s.store.SaveTask(task)
	return task, nil
}

func (s *Scheduler) AddOneTimeTask(name string, payload map[string]string, maxRetry int) *models.Task {
	task := &models.Task{
		ID:          s.generateID(),
		Type:        models.TaskTypeOneTime,
		Name:        name,
		Payload:     payload,
		Status:      models.TaskStatusPending,
		RetryCount:  0,
		MaxRetry:    maxRetry,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
	}
	s.store.SaveTask(task)
	return task
}

func (s *Scheduler) CancelTask(taskID string) bool {
	s.cronMu.Lock()
	defer s.cronMu.Unlock()

	if entryID, exists := s.cronJobs[taskID]; exists {
		s.cron.Remove(entryID)
		delete(s.cronJobs, taskID)
	}

	task, exists := s.store.GetTask(taskID)
	if !exists {
		return false
	}

	if task.Status == models.TaskStatusPending || task.Status == models.TaskStatusRunning {
		s.store.UpdateTaskStatus(taskID, models.TaskStatusCancelled)
	}
	return true
}

func (s *Scheduler) dispatchLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			tasks := s.store.GetPendingTasks()
			for _, task := range tasks {
				if task.Type == models.TaskTypeCron {
					continue
				}
				s.workerPool.Submit(task)
			}
		}
	}
}

func (s *Scheduler) Shutdown() {
	close(s.stopChan)
	s.cron.Stop()
}
