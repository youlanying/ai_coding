package service

import (
	"task-scheduler/internal/models"
	"task-scheduler/internal/scheduler"
	"task-scheduler/internal/store"
	"task-scheduler/internal/worker"
)

type TaskService struct {
	store      store.Store
	scheduler  *scheduler.Scheduler
	workerPool *worker.WorkerPool
}

func NewTaskService(st store.Store, scheduler *scheduler.Scheduler, workerPool *worker.WorkerPool) *TaskService {
	return &TaskService{
		store:      st,
		scheduler:  scheduler,
		workerPool: workerPool,
	}
}

type AddDelayTaskRequest struct {
	Name      string            `json:"name"`
	Payload   map[string]string `json:"payload"`
	DelayMs   int64             `json:"delay_ms"`
	MaxRetry  int               `json:"max_retry"`
	DependsOn []string          `json:"depends_on,omitempty"`
}

type AddCronTaskRequest struct {
	Name      string            `json:"name"`
	Payload   map[string]string `json:"payload"`
	CronExpr  string            `json:"cron_expr"`
	MaxRetry  int               `json:"max_retry"`
	DependsOn []string          `json:"depends_on,omitempty"`
}

type AddOneTimeTaskRequest struct {
	Name      string            `json:"name"`
	Payload   map[string]string `json:"payload"`
	MaxRetry  int               `json:"max_retry"`
	DependsOn []string          `json:"depends_on,omitempty"`
}

type TaskResponse struct {
	Task *models.Task `json:"task"`
}

type TaskListResponse struct {
	Tasks []*models.Task `json:"tasks"`
}

type TaskResultResponse struct {
	Results []*models.TaskResult `json:"results"`
}

type WorkerPoolResponse struct {
	Stats models.WorkerPoolStats `json:"stats"`
}

type SetWorkerCountRequest struct {
	Count int `json:"count"`
}

func (s *TaskService) AddDelayTask(req *AddDelayTaskRequest) (*models.Task, error) {
	task := s.scheduler.AddDelayTaskWithDeps(req.Name, req.Payload, req.DelayMs, req.MaxRetry, req.DependsOn)
	return task, nil
}

func (s *TaskService) AddCronTask(req *AddCronTaskRequest) (*models.Task, error) {
	task, err := s.scheduler.AddCronTaskWithDeps(req.Name, req.Payload, req.CronExpr, req.MaxRetry, req.DependsOn)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) AddOneTimeTask(req *AddOneTimeTaskRequest) (*models.Task, error) {
	task := s.scheduler.AddOneTimeTaskWithDeps(req.Name, req.Payload, req.MaxRetry, req.DependsOn)
	return task, nil
}

func (s *TaskService) GetTask(id string) (*models.Task, bool) {
	return s.store.GetTask(id)
}

func (s *TaskService) ListTasks(status string) []*models.Task {
	return s.store.ListTasks(models.TaskStatus(status))
}

func (s *TaskService) CancelTask(id string) bool {
	return s.scheduler.CancelTask(id)
}

func (s *TaskService) GetTaskResults(taskID string) []*models.TaskResult {
	return s.store.GetTaskResults(taskID)
}

func (s *TaskService) GetLatestResult(taskID string) (*models.TaskResult, bool) {
	return s.store.GetLatestResult(taskID)
}

func (s *TaskService) SetWorkerCount(count int) {
	s.workerPool.SetWorkerCount(count)
}

func (s *TaskService) GetWorkerPoolStats() models.WorkerPoolStats {
	return s.workerPool.GetStats()
}

func (s *TaskService) Shutdown() {
	s.scheduler.Shutdown()
	s.workerPool.Shutdown()
}
