package store

import (
	"task-scheduler/internal/models"
)

type Store interface {
	SaveTask(task *models.Task)
	GetTask(id string) (*models.Task, bool)
	DeleteTask(id string)
	ListTasks(status models.TaskStatus) []*models.Task
	GetPendingTasks() []*models.Task
	GetWaitingTasks() []*models.Task
	UpdateTaskStatus(id string, status models.TaskStatus) bool
	UpdateTaskResult(id string, result, err string) bool
	IncrementRetry(id string) bool
	SaveResult(result *models.TaskResult)
	GetTaskResults(taskID string) []*models.TaskResult
	GetLatestResult(taskID string) (*models.TaskResult, bool)
	GetDependents(taskID string) []*models.Task
	CheckDependencies(taskID string) bool
	LoadAllTasks() ([]*models.Task, error)
	LoadAllResults() ([]*models.TaskResult, error)
	Close() error
}
