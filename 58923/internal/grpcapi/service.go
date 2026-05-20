package grpcapi

import (
	"context"

	"google.golang.org/grpc"
)

type Task struct {
	Id          string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Type        string            `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Name        string            `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Payload     map[string]string `protobuf:"bytes,4,rep,name=payload,proto3" json:"payload,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	DependsOn   []string          `protobuf:"bytes,16,rep,name=depends_on,json=dependsOn,proto3" json:"depends_on,omitempty"`
	CronExpr    string            `protobuf:"bytes,5,opt,name=cron_expr,json=cronExpr,proto3" json:"cron_expr,omitempty"`
	DelayMs     int64             `protobuf:"varint,6,opt,name=delay_ms,json=delayMs,proto3" json:"delay_ms,omitempty"`
	Status      string            `protobuf:"bytes,7,opt,name=status,proto3" json:"status,omitempty"`
	RetryCount  int32             `protobuf:"varint,8,opt,name=retry_count,json=retryCount,proto3" json:"retry_count,omitempty"`
	MaxRetry    int32             `protobuf:"varint,9,opt,name=max_retry,json=maxRetry,proto3" json:"max_retry,omitempty"`
	CreatedAt   int64             `protobuf:"varint,10,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	ScheduledAt int64             `protobuf:"varint,11,opt,name=scheduled_at,json=scheduledAt,proto3" json:"scheduled_at,omitempty"`
	StartedAt   int64             `protobuf:"varint,12,opt,name=started_at,json=startedAt,proto3" json:"started_at,omitempty"`
	FinishedAt  int64             `protobuf:"varint,13,opt,name=finished_at,json=finishedAt,proto3" json:"finished_at,omitempty"`
	Result      string            `protobuf:"bytes,14,opt,name=result,proto3" json:"result,omitempty"`
	Error       string            `protobuf:"bytes,15,opt,name=error,proto3" json:"error,omitempty"`
}

type TaskResult struct {
	TaskId     string `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	Status     string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Result     string `protobuf:"bytes,3,opt,name=result,proto3" json:"result,omitempty"`
	Error      string `protobuf:"bytes,4,opt,name=error,proto3" json:"error,omitempty"`
	RetryCount int32  `protobuf:"varint,5,opt,name=retry_count,json=retryCount,proto3" json:"retry_count,omitempty"`
	ExecutedAt int64  `protobuf:"varint,6,opt,name=executed_at,json=executedAt,proto3" json:"executed_at,omitempty"`
	DurationMs int64  `protobuf:"varint,7,opt,name=duration_ms,json=durationMs,proto3" json:"duration_ms,omitempty"`
}

type WorkerPoolStats struct {
	ActiveWorkers int32 `protobuf:"varint,1,opt,name=active_workers,json=activeWorkers,proto3" json:"active_workers,omitempty"`
	IdleWorkers   int32 `protobuf:"varint,2,opt,name=idle_workers,json=idleWorkers,proto3" json:"idle_workers,omitempty"`
	TotalWorkers  int32 `protobuf:"varint,3,opt,name=total_workers,json=totalWorkers,proto3" json:"total_workers,omitempty"`
	PendingTasks  int32 `protobuf:"varint,4,opt,name=pending_tasks,json=pendingTasks,proto3" json:"pending_tasks,omitempty"`
}

type AddDelayTaskRequest struct {
	Name      string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Payload   map[string]string `protobuf:"bytes,2,rep,name=payload,proto3" json:"payload,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	DelayMs   int64             `protobuf:"varint,3,opt,name=delay_ms,json=delayMs,proto3" json:"delay_ms,omitempty"`
	MaxRetry  int32             `protobuf:"varint,4,opt,name=max_retry,json=maxRetry,proto3" json:"max_retry,omitempty"`
	DependsOn []string          `protobuf:"bytes,5,rep,name=depends_on,json=dependsOn,proto3" json:"depends_on,omitempty"`
}

type AddCronTaskRequest struct {
	Name      string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Payload   map[string]string `protobuf:"bytes,2,rep,name=payload,proto3" json:"payload,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	CronExpr  string            `protobuf:"bytes,3,opt,name=cron_expr,json=cronExpr,proto3" json:"cron_expr,omitempty"`
	MaxRetry  int32             `protobuf:"varint,4,opt,name=max_retry,json=maxRetry,proto3" json:"max_retry,omitempty"`
	DependsOn []string          `protobuf:"bytes,5,rep,name=depends_on,json=dependsOn,proto3" json:"depends_on,omitempty"`
}

type AddOneTimeTaskRequest struct {
	Name      string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Payload   map[string]string `protobuf:"bytes,2,rep,name=payload,proto3" json:"payload,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	MaxRetry  int32             `protobuf:"varint,3,opt,name=max_retry,json=maxRetry,proto3" json:"max_retry,omitempty"`
	DependsOn []string          `protobuf:"bytes,4,rep,name=depends_on,json=dependsOn,proto3" json:"depends_on,omitempty"`
}

type TaskResponse struct {
	Task *Task `protobuf:"bytes,1,opt,name=task,proto3" json:"task,omitempty"`
}

type TaskListResponse struct {
	Tasks []*Task `protobuf:"bytes,1,rep,name=tasks,proto3" json:"tasks,omitempty"`
}

type TaskResultListResponse struct {
	Results []*TaskResult `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
}

type TaskResultResponse struct {
	Result *TaskResult `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
}

type WorkerPoolStatsResponse struct {
	Stats *WorkerPoolStats `protobuf:"bytes,1,opt,name=stats,proto3" json:"stats,omitempty"`
}

type GetTaskRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

type ListTasksRequest struct {
	Status string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

type CancelTaskRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

type CancelTaskResponse struct {
	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

type GetTaskResultsRequest struct {
	TaskId string `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
}

type GetLatestResultRequest struct {
	TaskId string `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
}

type SetWorkerCountRequest struct {
	Count int32 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
}

type GetWorkerPoolStatsRequest struct{}

type TaskServiceServer interface {
	AddDelayTask(context.Context, *AddDelayTaskRequest) (*TaskResponse, error)
	AddCronTask(context.Context, *AddCronTaskRequest) (*TaskResponse, error)
	AddOneTimeTask(context.Context, *AddOneTimeTaskRequest) (*TaskResponse, error)
	GetTask(context.Context, *GetTaskRequest) (*TaskResponse, error)
	ListTasks(context.Context, *ListTasksRequest) (*TaskListResponse, error)
	CancelTask(context.Context, *CancelTaskRequest) (*CancelTaskResponse, error)
	GetTaskResults(context.Context, *GetTaskResultsRequest) (*TaskResultListResponse, error)
	GetLatestResult(context.Context, *GetLatestResultRequest) (*TaskResultResponse, error)
	SetWorkerCount(context.Context, *SetWorkerCountRequest) (*WorkerPoolStatsResponse, error)
	GetWorkerPoolStats(context.Context, *GetWorkerPoolStatsRequest) (*WorkerPoolStatsResponse, error)
}

type UnimplementedTaskServiceServer struct{}

func (UnimplementedTaskServiceServer) AddDelayTask(context.Context, *AddDelayTaskRequest) (*TaskResponse, error) {
	return nil, nil
}
func (UnimplementedTaskServiceServer) AddCronTask(context.Context, *AddCronTaskRequest) (*TaskResponse, error) {
	return nil, nil
}
func (UnimplementedTaskServiceServer) AddOneTimeTask(context.Context, *AddOneTimeTaskRequest) (*TaskResponse, error) {
	return nil, nil
}
func (UnimplementedTaskServiceServer) GetTask(context.Context, *GetTaskRequest) (*TaskResponse, error) {
	return nil, nil
}
func (UnimplementedTaskServiceServer) ListTasks(context.Context, *ListTasksRequest) (*TaskListResponse, error) {
	return nil, nil
}
func (UnimplementedTaskServiceServer) CancelTask(context.Context, *CancelTaskRequest) (*CancelTaskResponse, error) {
	return nil, nil
}
func (UnimplementedTaskServiceServer) GetTaskResults(context.Context, *GetTaskResultsRequest) (*TaskResultListResponse, error) {
	return nil, nil
}
func (UnimplementedTaskServiceServer) GetLatestResult(context.Context, *GetLatestResultRequest) (*TaskResultResponse, error) {
	return nil, nil
}
func (UnimplementedTaskServiceServer) SetWorkerCount(context.Context, *SetWorkerCountRequest) (*WorkerPoolStatsResponse, error) {
	return nil, nil
}
func (UnimplementedTaskServiceServer) GetWorkerPoolStats(context.Context, *GetWorkerPoolStatsRequest) (*WorkerPoolStatsResponse, error) {
	return nil, nil
}

func RegisterTaskServiceServer(s *grpc.Server, srv TaskServiceServer) {
	s.RegisterService(&TaskService_ServiceDesc, srv)
}

var TaskService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "task_scheduler.TaskService",
	HandlerType: (*TaskServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddDelayTask",
			Handler:    _TaskService_AddDelayTask_Handler,
		},
		{
			MethodName: "AddCronTask",
			Handler:    _TaskService_AddCronTask_Handler,
		},
		{
			MethodName: "AddOneTimeTask",
			Handler:    _TaskService_AddOneTimeTask_Handler,
		},
		{
			MethodName: "GetTask",
			Handler:    _TaskService_GetTask_Handler,
		},
		{
			MethodName: "ListTasks",
			Handler:    _TaskService_ListTasks_Handler,
		},
		{
			MethodName: "CancelTask",
			Handler:    _TaskService_CancelTask_Handler,
		},
		{
			MethodName: "GetTaskResults",
			Handler:    _TaskService_GetTaskResults_Handler,
		},
		{
			MethodName: "GetLatestResult",
			Handler:    _TaskService_GetLatestResult_Handler,
		},
		{
			MethodName: "SetWorkerCount",
			Handler:    _TaskService_SetWorkerCount_Handler,
		},
		{
			MethodName: "GetWorkerPoolStats",
			Handler:    _TaskService_GetWorkerPoolStats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "task_scheduler.proto",
}

func _TaskService_AddDelayTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddDelayTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).AddDelayTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/AddDelayTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).AddDelayTask(ctx, req.(*AddDelayTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_AddCronTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddCronTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).AddCronTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/AddCronTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).AddCronTask(ctx, req.(*AddCronTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_AddOneTimeTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddOneTimeTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).AddOneTimeTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/AddOneTimeTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).AddOneTimeTask(ctx, req.(*AddOneTimeTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_GetTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).GetTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/GetTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).GetTask(ctx, req.(*GetTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_ListTasks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTasksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).ListTasks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/ListTasks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).ListTasks(ctx, req.(*ListTasksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_CancelTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).CancelTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/CancelTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).CancelTask(ctx, req.(*CancelTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_GetTaskResults_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTaskResultsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).GetTaskResults(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/GetTaskResults",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).GetTaskResults(ctx, req.(*GetTaskResultsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_GetLatestResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLatestResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).GetLatestResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/GetLatestResult",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).GetLatestResult(ctx, req.(*GetLatestResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_SetWorkerCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetWorkerCountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).SetWorkerCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/SetWorkerCount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).SetWorkerCount(ctx, req.(*SetWorkerCountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_GetWorkerPoolStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorkerPoolStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).GetWorkerPoolStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task_scheduler.TaskService/GetWorkerPoolStats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).GetWorkerPoolStats(ctx, req.(*GetWorkerPoolStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

type TaskServiceClient interface {
	AddDelayTask(ctx context.Context, in *AddDelayTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error)
	AddCronTask(ctx context.Context, in *AddCronTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error)
	AddOneTimeTask(ctx context.Context, in *AddOneTimeTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error)
	GetTask(ctx context.Context, in *GetTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error)
	ListTasks(ctx context.Context, in *ListTasksRequest, opts ...grpc.CallOption) (*TaskListResponse, error)
	CancelTask(ctx context.Context, in *CancelTaskRequest, opts ...grpc.CallOption) (*CancelTaskResponse, error)
	GetTaskResults(ctx context.Context, in *GetTaskResultsRequest, opts ...grpc.CallOption) (*TaskResultListResponse, error)
	GetLatestResult(ctx context.Context, in *GetLatestResultRequest, opts ...grpc.CallOption) (*TaskResultResponse, error)
	SetWorkerCount(ctx context.Context, in *SetWorkerCountRequest, opts ...grpc.CallOption) (*WorkerPoolStatsResponse, error)
	GetWorkerPoolStats(ctx context.Context, in *GetWorkerPoolStatsRequest, opts ...grpc.CallOption) (*WorkerPoolStatsResponse, error)
}

type taskServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTaskServiceClient(cc grpc.ClientConnInterface) TaskServiceClient {
	return &taskServiceClient{cc}
}

func (c *taskServiceClient) AddDelayTask(ctx context.Context, in *AddDelayTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/AddDelayTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) AddCronTask(ctx context.Context, in *AddCronTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/AddCronTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) AddOneTimeTask(ctx context.Context, in *AddOneTimeTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/AddOneTimeTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) GetTask(ctx context.Context, in *GetTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/GetTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) ListTasks(ctx context.Context, in *ListTasksRequest, opts ...grpc.CallOption) (*TaskListResponse, error) {
	out := new(TaskListResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/ListTasks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) CancelTask(ctx context.Context, in *CancelTaskRequest, opts ...grpc.CallOption) (*CancelTaskResponse, error) {
	out := new(CancelTaskResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/CancelTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) GetTaskResults(ctx context.Context, in *GetTaskResultsRequest, opts ...grpc.CallOption) (*TaskResultListResponse, error) {
	out := new(TaskResultListResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/GetTaskResults", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) GetLatestResult(ctx context.Context, in *GetLatestResultRequest, opts ...grpc.CallOption) (*TaskResultResponse, error) {
	out := new(TaskResultResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/GetLatestResult", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) SetWorkerCount(ctx context.Context, in *SetWorkerCountRequest, opts ...grpc.CallOption) (*WorkerPoolStatsResponse, error) {
	out := new(WorkerPoolStatsResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/SetWorkerCount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) GetWorkerPoolStats(ctx context.Context, in *GetWorkerPoolStatsRequest, opts ...grpc.CallOption) (*WorkerPoolStatsResponse, error) {
	out := new(WorkerPoolStatsResponse)
	err := c.cc.Invoke(ctx, "/task_scheduler.TaskService/GetWorkerPoolStats", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
