package store

import (
	"sync"
	"time"

	"task-scheduler/internal/models"
)

type TaskStore struct {
	mu      sync.RWMutex
	tasks   map[string]*models.Task
	results map[string][]*models.TaskResult
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:   make(map[string]*models.Task),
		results: make(map[string][]*models.TaskResult),
	}
}

func (s *TaskStore) SaveTask(task *models.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

func (s *TaskStore) GetTask(id string) (*models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, exists := s.tasks[id]
	return task, exists
}

func (s *TaskStore) DeleteTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
}

func (s *TaskStore) ListTasks(status models.TaskStatus) []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var tasks []*models.Task
	for _, t := range s.tasks {
		if status == "" || t.Status == status {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

func (s *TaskStore) GetPendingTasks() []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var tasks []*models.Task
	now := time.Now()
	for _, t := range s.tasks {
		if t.Status == models.TaskStatusPending && t.ScheduledAt.Before(now) {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

func (s *TaskStore) UpdateTaskStatus(id string, status models.TaskStatus) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.tasks[id]
	if !exists {
		return false
	}
	task.Status = status
	if status == models.TaskStatusRunning {
		task.StartedAt = time.Now()
	} else if status == models.TaskStatusSuccess || status == models.TaskStatusFailed {
		task.FinishedAt = time.Now()
	}
	return true
}

func (s *TaskStore) UpdateTaskResult(id string, result, err string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.tasks[id]
	if !exists {
		return false
	}
	task.Result = result
	task.Error = err
	return true
}

func (s *TaskStore) IncrementRetry(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.tasks[id]
	if !exists {
		return false
	}
	task.RetryCount++
	return true
}

func (s *TaskStore) SaveResult(result *models.TaskResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[result.TaskID] = append(s.results[result.TaskID], result)
}

func (s *TaskStore) GetTaskResults(taskID string) []*models.TaskResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.results[taskID]
}

func (s *TaskStore) GetLatestResult(taskID string) (*models.TaskResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	results, exists := s.results[taskID]
	if !exists || len(results) == 0 {
		return nil, false
	}
	return results[len(results)-1], true
}
