package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"task-scheduler/internal/grpcapi"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := grpcapi.NewTaskServiceClient(conn)

	fmt.Println("=== Testing gRPC API ===")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("\n1. Adding a delay task (1 second)...")
	delayResp, err := client.AddDelayTask(ctx, &grpcapi.AddDelayTaskRequest{
		Name:     "grpc-test-delay",
		Payload:  map[string]string{"source": "grpc"},
		DelayMs:  1000,
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatalf("AddDelayTask failed: %v", err)
	}
	fmt.Printf("   Created task: %s\n", delayResp.Task.Id)

	fmt.Println("\n2. Adding a one-time task...")
	oneTimeResp, err := client.AddOneTimeTask(ctx, &grpcapi.AddOneTimeTaskRequest{
		Name:     "grpc-test-onetime",
		Payload:  map[string]string{"key": "grpc-value"},
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatalf("AddOneTimeTask failed: %v", err)
	}
	fmt.Printf("   Created task: %s\n", oneTimeResp.Task.Id)

	fmt.Println("\n3. Waiting 2 seconds for tasks to execute...")
	time.Sleep(2 * time.Second)

	fmt.Println("\n4. Getting task status...")
	taskResp, err := client.GetTask(ctx, &grpcapi.GetTaskRequest{Id: delayResp.Task.Id})
	if err != nil {
		log.Fatalf("GetTask failed: %v", err)
	}
	fmt.Printf("   Delay task status: %s, result: %s\n", taskResp.Task.Status, taskResp.Task.Result)

	taskResp2, err := client.GetTask(ctx, &grpcapi.GetTaskRequest{Id: oneTimeResp.Task.Id})
	if err != nil {
		log.Fatalf("GetTask failed: %v", err)
	}
	fmt.Printf("   OneTime task status: %s, result: %s\n", taskResp2.Task.Status, taskResp2.Task.Result)

	fmt.Println("\n5. Getting worker pool stats...")
	statsResp, err := client.GetWorkerPoolStats(ctx, &grpcapi.GetWorkerPoolStatsRequest{})
	if err != nil {
		log.Fatalf("GetWorkerPoolStats failed: %v", err)
	}
	fmt.Printf("   Workers: total=%d, active=%d, idle=%d\n",
		statsResp.Stats.TotalWorkers, statsResp.Stats.ActiveWorkers, statsResp.Stats.IdleWorkers)

	fmt.Println("\n6. Scaling workers to 8...")
	scaleResp, err := client.SetWorkerCount(ctx, &grpcapi.SetWorkerCountRequest{Count: 8})
	if err != nil {
		log.Fatalf("SetWorkerCount failed: %v", err)
	}
	fmt.Printf("   Workers after scaling: total=%d\n", scaleResp.Stats.TotalWorkers)

	fmt.Println("\n7. Listing all tasks...")
	listResp, err := client.ListTasks(ctx, &grpcapi.ListTasksRequest{})
	if err != nil {
		log.Fatalf("ListTasks failed: %v", err)
	}
	for _, t := range listResp.Tasks {
		fmt.Printf("   - %s [%s] %s\n", t.Id[:8]+"...", t.Status, t.Name)
	}

	fmt.Println("\n=== gRPC API Test Complete ===")
}
