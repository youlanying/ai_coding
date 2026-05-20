package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"task-scheduler/internal/grpcapi"
	"task-scheduler/internal/httpapi"
	"task-scheduler/internal/scheduler"
	"task-scheduler/internal/service"
	"task-scheduler/internal/store"
	"task-scheduler/internal/worker"
)

func main() {
	dbPath := "./tasks.db"
	sqliteStore, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to open SQLite store: %v", err)
	}
	defer sqliteStore.Close()

	log.Printf("Using SQLite store at %s", dbPath)

	if err := sqliteStore.RecoverTasks(); err != nil {
		log.Printf("Warning: failed to recover tasks: %v", err)
	}

	workerPool := worker.NewWorkerPool(sqliteStore, worker.DefaultExecutor, 5)
	taskScheduler := scheduler.NewScheduler(sqliteStore, workerPool)

	workerPool.OnTaskComplete = func(taskID string) {
		taskScheduler.NotifyTaskComplete(taskID)
	}

	taskScheduler.RecoverCronJobs()

	taskService := service.NewTaskService(sqliteStore, taskScheduler, workerPool)

	httpAddr := ":8080"
	grpcAddr := ":50051"

	httpServer := httpapi.NewServer(taskService)
	if _, err := httpServer.Serve(httpAddr); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
	log.Printf("HTTP server listening on %s", httpAddr)

	if _, err := grpcapi.Serve(taskService, grpcAddr); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
	log.Printf("gRPC server listening on %s", grpcAddr)

	fmt.Println("=== Task Scheduler Service Started ===")
	fmt.Println("HTTP API: http://localhost:8080/api/v1/")
	fmt.Println("gRPC API: localhost:50051")
	fmt.Println("SQLite: tasks.db")
	fmt.Println("")
	fmt.Println("Example HTTP calls:")
	fmt.Println("  # Add delay task (2 seconds)")
	fmt.Println(`  curl -X POST http://localhost:8080/api/v1/tasks/delay -H 'Content-Type: application/json' -d '{"name":"test","payload":{"key":"value"},"delay_ms":2000,"max_retry":3}'`)
	fmt.Println("")
	fmt.Println("  # Add task with dependencies")
	fmt.Println(`  curl -X POST http://localhost:8080/api/v1/tasks/onetime -H 'Content-Type: application/json' -d '{"name":"dependent-task","depends_on":["task-id-1","task-id-2"],"max_retry":3}'`)
	fmt.Println("")
	fmt.Println("  # Add cron task (every 5 seconds)")
	fmt.Println(`  curl -X POST http://localhost:8080/api/v1/tasks/cron -H 'Content-Type: application/json' -d '{"name":"cron-test","payload":{"foo":"bar"},"cron_expr":"*/5 * * * * *","max_retry":3}'`)
	fmt.Println("")
	fmt.Println("  # List all tasks")
	fmt.Println("  curl http://localhost:8080/api/v1/tasks")
	fmt.Println("")
	fmt.Println("  # Get worker pool stats")
	fmt.Println("  curl http://localhost:8080/api/v1/workers")
	fmt.Println("")
	fmt.Println("  # Adjust worker count to 10")
	fmt.Println(`  curl -X PUT http://localhost:8080/api/v1/workers -H 'Content-Type: application/json' -d '{"count":10}'`)
	fmt.Println("")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	taskService.Shutdown()
	log.Println("Service stopped")
}
