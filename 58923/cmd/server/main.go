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
	taskStore := store.NewTaskStore()
	workerPool := worker.NewWorkerPool(taskStore, worker.DefaultExecutor, 5)
	taskScheduler := scheduler.NewScheduler(taskStore, workerPool)
	taskService := service.NewTaskService(taskStore, taskScheduler, workerPool)

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
	fmt.Println("")
	fmt.Println("Example HTTP calls:")
	fmt.Println("  # Add delay task (2 seconds)")
	fmt.Println("  curl -X POST http://localhost:8080/api/v1/tasks/delay -H 'Content-Type: application/json' -d '{\"name\":\"test\",\"payload\":{\"key\":\"value\"},\"delay_ms\":2000,\"max_retry\":3}'")
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
