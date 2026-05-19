package models

import (
	"time"
)

type TaskType string

const (
	TaskTypeDelay    TaskType = "delay"
	TaskTypeCron     TaskType = "cron"
	TaskTypeOneTime  TaskType = "one_time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSuccess   TaskStatus = "success"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID          string            `json:"id"`
	Type        TaskType          `json:"type"`
	Name        string            `json:"name"`
	Payload     map[string]string `json:"payload"`
	CronExpr    string            `json:"cron_expr,omitempty"`
	DelayMs     int64             `json:"delay_ms,omitempty"`
	Status      TaskStatus        `json:"status"`
	RetryCount  int               `json:"retry_count"`
	MaxRetry    int               `json:"max_retry"`
	CreatedAt   time.Time         `json:"created_at"`
	ScheduledAt time.Time         `json:"scheduled_at"`
	StartedAt   time.Time         `json:"started_at,omitempty"`
	FinishedAt  time.Time         `json:"finished_at,omitempty"`
	Result      string            `json:"result,omitempty"`
	Error       string            `json:"error,omitempty"`
}

type TaskResult struct {
	TaskID     string    `json:"task_id"`
	Status     TaskStatus `json:"status"`
	Result     string    `json:"result,omitempty"`
	Error      string    `json:"error,omitempty"`
	RetryCount int       `json:"retry_count"`
	ExecutedAt time.Time `json:"executed_at"`
	DurationMs int64     `json:"duration_ms"`
}

type WorkerPoolStats struct {
	ActiveWorkers int `json:"active_workers"`
	IdleWorkers   int `json:"idle_workers"`
	TotalWorkers  int `json:"total_workers"`
	PendingTasks  int `json:"pending_tasks"`
}
