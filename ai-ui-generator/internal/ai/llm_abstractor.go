package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// LLMClient defines the interface for LLM communication
type LLMClient interface {
	Generate(prompt string) (string, error)
	StreamGenerate(prompt string, responseChannel chan string) error
}

// OpenAICompatibleClient implements LLMClient for OpenAI-compatible APIs
type OpenAICompatibleClient struct {
	endpoint    string
	apiKey      string
	modelName   string
	maxTokens   int
	temperature float64
}

// NewOpenAICompatibleClient creates a new OpenAI-compatible client
func NewOpenAICompatibleClient(endpoint, apiKey, modelName string, maxTokens int, temperature float64) *OpenAICompatibleClient {
	return &OpenAICompatibleClient{
		endpoint:    endpoint,
		apiKey:      apiKey,
		modelName:   modelName,
		maxTokens:   maxTokens,
		temperature: temperature,
	}
}

// Generate generates text using the LLM
func (c *OpenAICompatibleClient) Generate(prompt string) (string, error) {
	payload := map[string]interface{}{
		"model":       c.modelName,
		"prompt":      prompt,
		"max_tokens":  c.maxTokens,
		"temperature": c.temperature,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm error: %s", string(b))
	}
	var result struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}
	return result.Choices[0].Text, nil
}

// StreamGenerate generates text with streaming
func (c *OpenAICompatibleClient) StreamGenerate(prompt string, responseChannel chan string) error {
	payload := map[string]interface{}{
		"model":       c.modelName,
		"prompt":      prompt,
		"max_tokens":  c.maxTokens,
		"temperature": c.temperature,
		"stream":      true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llm error: %s", string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || len(line) < 6 || line[:6] != "data: " {
			continue
		}
		data := line[6:]
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Text string `json:"text"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err == nil {
			for _, choice := range chunk.Choices {
				responseChannel <- choice.Text
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// MockLLMClient implements LLMClient for testing
type MockLLMClient struct{}

// NewMockLLMClient creates a new mock LLM client
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{}
}

// Generate returns mock generated text
func (c *MockLLMClient) Generate(prompt string) (string, error) {
	return "// Mock generated code\nfunction MockComponent() {\n  return <div>Hello World</div>;\n}", nil
}

// StreamGenerate returns mock streaming text
func (c *MockLLMClient) StreamGenerate(prompt string, responseChannel chan string) error {
	mockResponse := "// Mock streaming response\nfunction Component() {\n  return <div>Streaming</div>;\n}"
	responseChannel <- mockResponse
	close(responseChannel)
	return nil
}
