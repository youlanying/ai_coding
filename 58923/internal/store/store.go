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

func (s *TaskStore) GetWaitingTasks() []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var tasks []*models.Task
	for _, t := range s.tasks {
		if t.Status == models.TaskStatusWaiting {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

func (s *TaskStore) GetDependents(taskID string) []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var dependents []*models.Task
	for _, t := range s.tasks {
		if t.Status == models.TaskStatusWaiting {
			for _, dep := range t.DependsOn {
				if dep == taskID {
					dependents = append(dependents, t)
					break
				}
			}
		}
	}
	return dependents
}

func (s *TaskStore) CheckDependencies(taskID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, exists := s.tasks[taskID]
	if !exists {
		return false
	}
	if len(task.DependsOn) == 0 {
		return true
	}
	for _, depID := range task.DependsOn {
		depTask, exists := s.tasks[depID]
		if !exists {
			return false
		}
		if depTask.Status != models.TaskStatusSuccess {
			return false
		}
	}
	return true
}

func (s *TaskStore) LoadAllTasks() ([]*models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]*models.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *TaskStore) LoadAllResults() ([]*models.TaskResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []*models.TaskResult
	for _, rs := range s.results {
		results = append(results, rs...)
	}
	return results, nil
}

func (s *TaskStore) Close() error {
	return nil
}
