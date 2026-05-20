package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Task struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	DependsOn []string          `json:"depends_on,omitempty"`
	Payload   map[string]string `json:"payload"`
}

func main() {
	baseURL := "http://localhost:8080/api/v1"

	fmt.Println("=== 测试1: 任务依赖功能 ===")

	taskAID := addTask(baseURL, "task-a", nil, nil)
	fmt.Printf("创建任务 A: %s\n", taskAID)

	taskBID := addTask(baseURL, "task-b", nil, nil)
	fmt.Printf("创建任务 B: %s\n", taskBID)

	taskCID := addTask(baseURL, "task-c-depends-on-a-b", nil, []string{taskAID, taskBID})
	fmt.Printf("创建任务 C (依赖 A 和 B): %s\n", taskCID)

	fmt.Println("\n等待 3 秒让任务执行...")
	time.Sleep(3 * time.Second)

	fmt.Println("\n任务 C 状态:")
	printTaskStatus(baseURL, taskCID)

	fmt.Println("\n=== 测试2: 失败任务不触发依赖 ===")

	taskDID := addTask(baseURL, "task-d-fail", map[string]string{"action": "fail"}, nil)
	fmt.Printf("创建失败任务 D: %s\n", taskDID)

	taskEID := addTask(baseURL, "task-e-depends-on-d", nil, []string{taskDID})
	fmt.Printf("创建任务 E (依赖 D): %s\n", taskEID)

	fmt.Println("\n等待 5 秒让任务执行和重试...")
	time.Sleep(5 * time.Second)

	fmt.Println("\n任务 D 状态:")
	printTaskStatus(baseURL, taskDID)
	fmt.Println("\n任务 E 状态 (应该保持 waiting):")
	printTaskStatus(baseURL, taskEID)

	fmt.Println("\n=== 测试完成 ===")
}

func addTask(baseURL, name string, payload map[string]string, dependsOn []string) string {
	reqBody := map[string]interface{}{
		"name":      name,
		"payload":   payload,
		"max_retry": 1,
	}
	if dependsOn != nil {
		reqBody["depends_on"] = dependsOn
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(baseURL+"/tasks/onetime", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result struct {
		Task Task `json:"task"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Task.ID
}

func printTaskStatus(baseURL, taskID string) {
	resp, err := http.Get(baseURL + "/tasks/" + taskID)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Task Task `json:"task"`
	}
	json.Unmarshal(body, &result)

	fmt.Printf("  ID: %s\n", result.Task.ID)
	fmt.Printf("  名称: %s\n", result.Task.Name)
	fmt.Printf("  状态: %s\n", result.Task.Status)
	if len(result.Task.DependsOn) > 0 {
		fmt.Printf("  依赖: %v\n", result.Task.DependsOn)
	}
}
