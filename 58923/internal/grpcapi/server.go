package grpcapi

import (
	"context"
	"encoding/json"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"task-scheduler/internal/models"
	"task-scheduler/internal/service"
)

type TaskServer struct {
	UnimplementedTaskServiceServer
	svc *service.TaskService
}

func NewTaskServer(svc *service.TaskService) *TaskServer {
	return &TaskServer{svc: svc}
}

func (s *TaskServer) AddDelayTask(ctx context.Context, req *AddDelayTaskRequest) (*TaskResponse, error) {
	taskReq := &service.AddDelayTaskRequest{
		Name:     req.Name,
		Payload:  req.Payload,
		DelayMs:  req.DelayMs,
		MaxRetry: int(req.MaxRetry),
	}
	task, err := s.svc.AddDelayTask(taskReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add delay task: %v", err)
	}
	return taskToProto(task), nil
}

func (s *TaskServer) AddCronTask(ctx context.Context, req *AddCronTaskRequest) (*TaskResponse, error) {
	taskReq := &service.AddCronTaskRequest{
		Name:     req.Name,
		Payload:  req.Payload,
		CronExpr: req.CronExpr,
		MaxRetry: int(req.MaxRetry),
	}
	task, err := s.svc.AddCronTask(taskReq)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cron expression: %v", err)
	}
	return taskToProto(task), nil
}

func (s *TaskServer) AddOneTimeTask(ctx context.Context, req *AddOneTimeTaskRequest) (*TaskResponse, error) {
	taskReq := &service.AddOneTimeTaskRequest{
		Name:     req.Name,
		Payload:  req.Payload,
		MaxRetry: int(req.MaxRetry),
	}
	task, err := s.svc.AddOneTimeTask(taskReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add one-time task: %v", err)
	}
	return taskToProto(task), nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *GetTaskRequest) (*TaskResponse, error) {
	task, exists := s.svc.GetTask(req.Id)
	if !exists {
		return nil, status.Errorf(codes.NotFound, "task not found")
	}
	return taskToProto(task), nil
}

func (s *TaskServer) ListTasks(ctx context.Context, req *ListTasksRequest) (*TaskListResponse, error) {
	tasks := s.svc.ListTasks(req.Status)
	protoTasks := make([]*Task, len(tasks))
	for i, t := range tasks {
		protoTasks[i] = taskToProto(t).Task
	}
	return &TaskListResponse{Tasks: protoTasks}, nil
}

func (s *TaskServer) CancelTask(ctx context.Context, req *CancelTaskRequest) (*CancelTaskResponse, error) {
	ok := s.svc.CancelTask(req.Id)
	return &CancelTaskResponse{Success: ok}, nil
}

func (s *TaskServer) GetTaskResults(ctx context.Context, req *GetTaskResultsRequest) (*TaskResultListResponse, error) {
	results := s.svc.GetTaskResults(req.TaskId)
	protoResults := make([]*TaskResult, len(results))
	for i, r := range results {
		protoResults[i] = resultToProto(r)
	}
	return &TaskResultListResponse{Results: protoResults}, nil
}

func (s *TaskServer) GetLatestResult(ctx context.Context, req *GetLatestResultRequest) (*TaskResultResponse, error) {
	result, exists := s.svc.GetLatestResult(req.TaskId)
	if !exists {
		return nil, status.Errorf(codes.NotFound, "no results found for task")
	}
	return &TaskResultResponse{Result: resultToProto(result)}, nil
}

func (s *TaskServer) SetWorkerCount(ctx context.Context, req *SetWorkerCountRequest) (*WorkerPoolStatsResponse, error) {
	s.svc.SetWorkerCount(int(req.Count))
	stats := s.svc.GetWorkerPoolStats()
	return statsToProto(stats), nil
}

func (s *TaskServer) GetWorkerPoolStats(ctx context.Context, req *GetWorkerPoolStatsRequest) (*WorkerPoolStatsResponse, error) {
	stats := s.svc.GetWorkerPoolStats()
	return statsToProto(stats), nil
}

func taskToProto(task *models.Task) *TaskResponse {
	return &TaskResponse{
		Task: &Task{
			Id:          task.ID,
			Type:        string(task.Type),
			Name:        task.Name,
			Payload:     task.Payload,
			CronExpr:    task.CronExpr,
			DelayMs:     task.DelayMs,
			Status:      string(task.Status),
			RetryCount:  int32(task.RetryCount),
			MaxRetry:    int32(task.MaxRetry),
			CreatedAt:   task.CreatedAt.UnixNano(),
			ScheduledAt: task.ScheduledAt.UnixNano(),
			StartedAt:   task.StartedAt.UnixNano(),
			FinishedAt:  task.FinishedAt.UnixNano(),
			Result:      task.Result,
			Error:       task.Error,
		},
	}
}

func resultToProto(r *models.TaskResult) *TaskResult {
	return &TaskResult{
		TaskId:     r.TaskID,
		Status:     string(r.Status),
		Result:     r.Result,
		Error:      r.Error,
		RetryCount: int32(r.RetryCount),
		ExecutedAt: r.ExecutedAt.UnixNano(),
		DurationMs: r.DurationMs,
	}
}

func statsToProto(stats models.WorkerPoolStats) *WorkerPoolStatsResponse {
	return &WorkerPoolStatsResponse{
		Stats: &WorkerPoolStats{
			ActiveWorkers: int32(stats.ActiveWorkers),
			IdleWorkers:   int32(stats.IdleWorkers),
			TotalWorkers:  int32(stats.TotalWorkers),
			PendingTasks:  int32(stats.PendingTasks),
		},
	}
}

func Serve(svc *service.TaskService, addr string) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	RegisterTaskServiceServer(grpcServer, NewTaskServer(svc))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return grpcServer, nil
}

func MustMarshalJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
