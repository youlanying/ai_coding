package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"task-scheduler/internal/models"
)

type SQLiteStore struct {
	db    *sql.DB
	cache *TaskStore
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	store := &SQLiteStore{
		db:    db,
		cache: NewTaskStore(),
	}

	if err := store.initSchema(); err != nil {
		return nil, err
	}

	if err := store.loadFromDB(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		payload TEXT,
		cron_expr TEXT,
		delay_ms INTEGER,
		status TEXT NOT NULL,
		retry_count INTEGER DEFAULT 0,
		max_retry INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		scheduled_at INTEGER NOT NULL,
		started_at INTEGER,
		finished_at INTEGER,
		result TEXT,
		error TEXT
	);

	CREATE TABLE IF NOT EXISTS task_dependencies (
		task_id TEXT NOT NULL,
		depends_on TEXT NOT NULL,
		PRIMARY KEY (task_id, depends_on),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS task_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT NOT NULL,
		status TEXT NOT NULL,
		result TEXT,
		error TEXT,
		retry_count INTEGER DEFAULT 0,
		executed_at INTEGER NOT NULL,
		duration_ms INTEGER NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_results_task_id ON task_results(task_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) loadFromDB() error {
	tasks, err := s.LoadAllTasks()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		s.cache.SaveTask(task)
	}

	results, err := s.LoadAllResults()
	if err != nil {
		return err
	}
	for _, r := range results {
		s.cache.SaveResult(r)
	}

	return nil
}

func (s *SQLiteStore) SaveTask(task *models.Task) {
	s.cache.SaveTask(task)

	payloadJSON, _ := json.Marshal(task.Payload)

	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT OR REPLACE INTO tasks 
		(id, type, name, payload, cron_expr, delay_ms, status, retry_count, max_retry, 
		 created_at, scheduled_at, started_at, finished_at, result, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.Type, task.Name, string(payloadJSON), task.CronExpr, task.DelayMs,
		task.Status, task.RetryCount, task.MaxRetry,
		task.CreatedAt.UnixNano(), task.ScheduledAt.UnixNano(),
		nilTime(task.StartedAt), nilTime(task.FinishedAt),
		task.Result, task.Error)
	if err != nil {
		return
	}

	_, err = tx.Exec("DELETE FROM task_dependencies WHERE task_id = ?", task.ID)
	if err != nil {
		return
	}

	for _, depID := range task.DependsOn {
		_, err = tx.Exec("INSERT OR IGNORE INTO task_dependencies (task_id, depends_on) VALUES (?, ?)", task.ID, depID)
		if err != nil {
			return
		}
	}

	tx.Commit()
}

func (s *SQLiteStore) GetTask(id string) (*models.Task, bool) {
	return s.cache.GetTask(id)
}

func (s *SQLiteStore) DeleteTask(id string) {
	s.cache.DeleteTask(id)
	tx, _ := s.db.Begin()
	tx.Exec("DELETE FROM task_dependencies WHERE task_id = ?", id)
	tx.Exec("DELETE FROM task_results WHERE task_id = ?", id)
	tx.Exec("DELETE FROM tasks WHERE id = ?", id)
	tx.Commit()
}

func (s *SQLiteStore) ListTasks(status models.TaskStatus) []*models.Task {
	return s.cache.ListTasks(status)
}

func (s *SQLiteStore) GetPendingTasks() []*models.Task {
	return s.cache.GetPendingTasks()
}

func (s *SQLiteStore) GetWaitingTasks() []*models.Task {
	return s.cache.GetWaitingTasks()
}

func (s *SQLiteStore) UpdateTaskStatus(id string, status models.TaskStatus) bool {
	if !s.cache.UpdateTaskStatus(id, status) {
		return false
	}

	task, _ := s.cache.GetTask(id)
	var err error

	switch status {
	case models.TaskStatusRunning:
		_, err = s.db.Exec("UPDATE tasks SET status = ?, started_at = ? WHERE id = ?",
			status, task.StartedAt.UnixNano(), id)
	case models.TaskStatusSuccess, models.TaskStatusFailed:
		_, err = s.db.Exec("UPDATE tasks SET status = ?, finished_at = ?, result = ?, error = ? WHERE id = ?",
			status, task.FinishedAt.UnixNano(), task.Result, task.Error, id)
	default:
		_, err = s.db.Exec("UPDATE tasks SET status = ? WHERE id = ?", status, id)
	}

	return err == nil
}

func (s *SQLiteStore) UpdateTaskResult(id string, result, errStr string) bool {
	if !s.cache.UpdateTaskResult(id, result, errStr) {
		return false
	}
	_, err := s.db.Exec("UPDATE tasks SET result = ?, error = ? WHERE id = ?", result, errStr, id)
	return err == nil
}

func (s *SQLiteStore) IncrementRetry(id string) bool {
	if !s.cache.IncrementRetry(id) {
		return false
	}
	task, _ := s.cache.GetTask(id)
	_, err := s.db.Exec("UPDATE tasks SET retry_count = ? WHERE id = ?", task.RetryCount, id)
	return err == nil
}

func (s *SQLiteStore) SaveResult(result *models.TaskResult) {
	s.cache.SaveResult(result)
	_, err := s.db.Exec(`
		INSERT INTO task_results (task_id, status, result, error, retry_count, executed_at, duration_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		result.TaskID, result.Status, result.Result, result.Error,
		result.RetryCount, result.ExecutedAt.UnixNano(), result.DurationMs)
	_ = err
}

func (s *SQLiteStore) GetTaskResults(taskID string) []*models.TaskResult {
	return s.cache.GetTaskResults(taskID)
}

func (s *SQLiteStore) GetLatestResult(taskID string) (*models.TaskResult, bool) {
	return s.cache.GetLatestResult(taskID)
}

func (s *SQLiteStore) GetDependents(taskID string) []*models.Task {
	return s.cache.GetDependents(taskID)
}

func (s *SQLiteStore) CheckDependencies(taskID string) bool {
	return s.cache.CheckDependencies(taskID)
}

func (s *SQLiteStore) LoadAllTasks() ([]*models.Task, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.type, t.name, t.payload, t.cron_expr, t.delay_ms, t.status,
		       t.retry_count, t.max_retry, t.created_at, t.scheduled_at, t.started_at,
		       t.finished_at, t.result, t.error,
		       GROUP_CONCAT(td.depends_on, ',') as dependencies
		FROM tasks t
		LEFT JOIN task_dependencies td ON t.id = td.task_id
		GROUP BY t.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var payloadJSON, dependenciesStr sql.NullString
		var startedAt, finishedAt sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.Type, &task.Name, &payloadJSON, &task.CronExpr, &task.DelayMs,
			&task.Status, &task.RetryCount, &task.MaxRetry,
			new(int64), new(int64), &startedAt, &finishedAt,
			&task.Result, &task.Error, &dependenciesStr)
		if err != nil {
			return nil, err
		}

		if payloadJSON.Valid && payloadJSON.String != "" {
			json.Unmarshal([]byte(payloadJSON.String), &task.Payload)
		}

		if dependenciesStr.Valid && dependenciesStr.String != "" {
			task.DependsOn = strings.Split(dependenciesStr.String, ",")
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *SQLiteStore) LoadAllResults() ([]*models.TaskResult, error) {
	rows, err := s.db.Query(`
		SELECT task_id, status, result, error, retry_count, executed_at, duration_ms
		FROM task_results ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.TaskResult
	for rows.Next() {
		r := &models.TaskResult{}
		var executedAtNano int64
		err := rows.Scan(&r.TaskID, &r.Status, &r.Result, &r.Error,
			&r.RetryCount, &executedAtNano, &r.DurationMs)
		if err != nil {
			return nil, err
		}
		r.ExecutedAt = time.Unix(0, executedAtNano)
		results = append(results, r)
	}

	return results, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func nilTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t.UnixNano()
}

func (s *SQLiteStore) GetCache() *TaskStore {
	return s.cache
}

func (s *SQLiteStore) RecoverTasks() error {
	rows, err := s.db.Query(`
		SELECT id, status FROM tasks 
		WHERE status IN (?, ?, ?)`,
		models.TaskStatusPending, models.TaskStatusWaiting, models.TaskStatusRunning)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, status string
		rows.Scan(&id, &status)

		if status == string(models.TaskStatusRunning) {
			s.cache.UpdateTaskStatus(id, models.TaskStatusPending)
			s.db.Exec("UPDATE tasks SET status = ? WHERE id = ?", models.TaskStatusPending, id)
		}
	}

	fmt.Printf("Recovered %d pending/waiting tasks\n", len(s.cache.ListTasks("")))
	return nil
}
