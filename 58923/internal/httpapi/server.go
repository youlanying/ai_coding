package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"task-scheduler/internal/service"
)

type Server struct {
	svc *service.TaskService
	mux *http.ServeMux
}

func NewServer(svc *service.TaskService) *Server {
	s := &Server{
		svc: svc,
		mux: http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/v1/tasks/delay", s.handleAddDelayTask)
	s.mux.HandleFunc("/api/v1/tasks/cron", s.handleAddCronTask)
	s.mux.HandleFunc("/api/v1/tasks/onetime", s.handleAddOneTimeTask)
	s.mux.HandleFunc("/api/v1/tasks/", s.handleTaskOperations)
	s.mux.HandleFunc("/api/v1/tasks", s.handleListTasks)
	s.mux.HandleFunc("/api/v1/workers", s.handleWorkerOperations)
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (s *Server) handleAddDelayTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req service.AddDelayTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	task, err := s.svc.AddDelayTask(&req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleAddCronTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req service.AddCronTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	task, err := s.svc.AddCronTask(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleAddOneTimeTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req service.AddOneTimeTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	task, err := s.svc.AddOneTimeTask(&req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleTaskOperations(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "task id required")
		return
	}

	taskID := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			task, exists := s.svc.GetTask(taskID)
			if !exists {
				writeError(w, http.StatusNotFound, "task not found")
				return
			}
			writeJSON(w, http.StatusOK, task)
		case http.MethodDelete:
			ok := s.svc.CancelTask(taskID)
			writeJSON(w, http.StatusOK, map[string]bool{"success": ok})
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(parts) >= 2 && parts[1] == "results" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if len(parts) >= 3 && parts[2] == "latest" {
			result, exists := s.svc.GetLatestResult(taskID)
			if !exists {
				writeError(w, http.StatusNotFound, "no results found")
				return
			}
			writeJSON(w, http.StatusOK, result)
		} else {
			results := s.svc.GetTaskResults(taskID)
			writeJSON(w, http.StatusOK, results)
		}
	}
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	status := r.URL.Query().Get("status")
	tasks := s.svc.ListTasks(status)
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleWorkerOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		stats := s.svc.GetWorkerPoolStats()
		writeJSON(w, http.StatusOK, stats)
	case http.MethodPut, http.MethodPost:
		var req service.SetWorkerCountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		s.svc.SetWorkerCount(req.Count)
		stats := s.svc.GetWorkerPoolStats()
		writeJSON(w, http.StatusOK, stats)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func writeError(w http.ResponseWriter, status int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   errMsg,
	})
}

func (s *Server) Serve(addr string) (*http.Server, error) {
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return httpServer, nil
}
