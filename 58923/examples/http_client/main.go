package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type Task struct {
	ID     string            `json:"id"`
	Type   string            `json:"type"`
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Result string            `json:"result,omitempty"`
	Error  string            `json:"error,omitempty"`
}

func main() {
	baseURL := "http://localhost:8080/api/v1"

	fmt.Println("=== Testing HTTP API ===")

	fmt.Println("\n1. Adding a delay task (1 second)...")
	taskID := addDelayTask(baseURL, "test-delay", map[string]string{"action": "test"}, 1000, 3)
	fmt.Printf("   Created task: %s\n", taskID)

	fmt.Println("\n2. Adding a one-time task...")
	oneTimeID := addOneTimeTask(baseURL, "test-onetime", map[string]string{"key": "value"}, 3)
	fmt.Printf("   Created task: %s\n", oneTimeID)

	fmt.Println("\n3. Waiting 2 seconds for tasks to execute...")
	time.Sleep(2 * time.Second)

	fmt.Println("\n4. Checking task results...")
	checkTask(baseURL, taskID)
	checkTask(baseURL, oneTimeID)

	fmt.Println("\n5. Getting worker pool stats...")
	stats := getWorkerStats(baseURL)
	fmt.Printf("   Workers: total=%d, active=%d, idle=%d, pending=%d\n",
		stats["total_workers"], stats["active_workers"], stats["idle_workers"], stats["pending_tasks"])

	fmt.Println("\n6. Scaling workers to 10...")
	stats = setWorkerCount(baseURL, 10)
	fmt.Printf("   Workers: total=%d, active=%d, idle=%d\n",
		stats["total_workers"], stats["active_workers"], stats["idle_workers"])

	fmt.Println("\n7. Listing all tasks...")
	listTasks(baseURL)

	fmt.Println("\n=== HTTP API Test Complete ===")
}

func addDelayTask(baseURL, name string, payload map[string]string, delayMs int64, maxRetry int) string {
	body := map[string]interface{}{
		"name":      name,
		"payload":   payload,
		"delay_ms":  delayMs,
		"max_retry": maxRetry,
	}
	var task Task
	postJSON(baseURL+"/tasks/delay", body, &task)
	return task.ID
}

func addOneTimeTask(baseURL, name string, payload map[string]string, maxRetry int) string {
	body := map[string]interface{}{
		"name":      name,
		"payload":   payload,
		"max_retry": maxRetry,
	}
	var task Task
	postJSON(baseURL+"/tasks/onetime", body, &task)
	return task.ID
}

func checkTask(baseURL, taskID string) {
	var task Task
	getJSON(baseURL+"/tasks/"+taskID, &task)
	fmt.Printf("   Task %s: status=%s, result=%s, error=%s\n",
		taskID[:8]+"...", task.Status, task.Result, task.Error)

	var results []map[string]interface{}
	getJSON(baseURL+"/tasks/"+taskID+"/results", &results)
	fmt.Printf("   Execution history: %d records\n", len(results))
}

func getWorkerStats(baseURL string) map[string]interface{} {
	var stats map[string]interface{}
	getJSON(baseURL+"/workers", &stats)
	return stats
}

func setWorkerCount(baseURL string, count int) map[string]interface{} {
	body := map[string]int{"count": count}
	var stats map[string]interface{}
	putJSON(baseURL+"/workers", body, &stats)
	return stats
}

func listTasks(baseURL string) {
	var tasks []Task
	getJSON(baseURL+"/tasks", &tasks)
	for _, t := range tasks {
		fmt.Printf("   - %s [%s] %s\n", t.ID[:8]+"...", t.Status, t.Name)
	}
}

func postJSON(url string, body interface{}, result interface{}) {
	b, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	parseResponse(resp, result)
}

func putJSON(url string, body interface{}, result interface{}) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	parseResponse(resp, result)
}

func getJSON(url string, result interface{}) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	parseResponse(resp, result)
}

func parseResponse(resp *http.Response, result interface{}) {
	body, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	json.Unmarshal(body, &apiResp)
	if !apiResp.Success {
		panic(fmt.Sprintf("API error: %s", apiResp.Error))
	}
	if result != nil && len(apiResp.Data) > 0 {
		json.Unmarshal(apiResp.Data, result)
	}
}
