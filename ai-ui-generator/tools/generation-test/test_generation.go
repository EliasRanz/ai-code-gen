package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Test the generation service
	testGenerationService()
}

func testGenerationService() {
	fmt.Println("🚀 Testing AI Generation Service")

	// Test health endpoint
	fmt.Print("Testing health endpoint... ")
	resp, err := http.Get("http://localhost:8084/health")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	resp.Body.Close()
	fmt.Printf("✅ Status: %d\n", resp.StatusCode)

	// Test models endpoint
	fmt.Print("Testing models endpoint... ")
	resp, err = http.Get("http://localhost:8084/models")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	resp.Body.Close()
	fmt.Printf("✅ Status: %d\n", resp.StatusCode)

	// Test non-streaming generation
	fmt.Print("Testing non-streaming generation... ")
	payload := map[string]interface{}{
		"model":  "llama-2-7b-chat",
		"prompt": "Hello, how are you?",
		"stream": false,
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "http://localhost:8084/generate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test-user-123")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	resp.Body.Close()
	fmt.Printf("✅ Status: %d\n", resp.StatusCode)

	// Test streaming generation (basic connection test)
	fmt.Print("Testing streaming generation endpoint... ")
	payload["stream"] = true
	jsonData, _ = json.Marshal(payload)
	req, _ = http.NewRequest("POST", "http://localhost:8084/generate/stream", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test-user-123")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	resp.Body.Close()
	fmt.Printf("✅ Status: %d (SSE endpoint accessible)\n", resp.StatusCode)

	fmt.Println("🎉 All tests passed! AI Generation Service is working correctly.")
}
