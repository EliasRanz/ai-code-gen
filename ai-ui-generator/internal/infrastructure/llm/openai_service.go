package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ai-code-gen/ai-ui-generator/internal/domain/ai"
)

// OpenAIService implements LLMService using OpenAI API
type OpenAIService struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOpenAIService creates a new OpenAI service
func NewOpenAIService(apiKey string) *OpenAIService {
	return &OpenAIService{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   "gpt-3.5-turbo",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// OpenAIRequest represents the request format for OpenAI API
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response format from OpenAI API
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// Choice represents a completion choice
type Choice struct {
	Index   int         `json:"index"`
	Message Message     `json:"message"`
	Delta   MessageDelta `json:"delta,omitempty"`
}

// MessageDelta represents a partial message in streaming
type MessageDelta struct {
	Content string `json:"content,omitempty"`
}

// Generate implements non-streaming code generation
func (s *OpenAIService) Generate(ctx context.Context, req ai.GenerationRequest) (ai.GenerationResult, error) {
	openAIReq := OpenAIRequest{
		Model: s.model,
		Messages: []Message{
			{Role: "user", Content: req.Prompt},
		},
		Stream: false,
	}

	resp, err := s.makeRequest(ctx, openAIReq)
	if err != nil {
		return ai.GenerationResult{}, err
	}
	defer resp.Body.Close()

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return ai.GenerationResult{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return ai.GenerationResult{}, fmt.Errorf("no choices in response")
	}

	return ai.GenerationResult{
		Code:       openAIResp.Choices[0].Message.Content,
		Model:      openAIResp.Model,
		UsedTokens: 0, // OpenAI doesn't return token count in this response format
	}, nil
}

// GenerateStream implements streaming code generation
func (s *OpenAIService) GenerateStream(ctx context.Context, req ai.GenerationRequest, ch chan<- ai.StreamChunk) error {
	defer close(ch)

	openAIReq := OpenAIRequest{
		Model: s.model,
		Messages: []Message{
			{Role: "user", Content: req.Prompt},
		},
		Stream: true,
	}

	resp, err := s.makeRequest(ctx, openAIReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return s.processStreamResponse(ctx, resp.Body, ch)
}

// Stream implements legacy streaming interface
func (s *OpenAIService) Stream(ctx context.Context, req ai.GenerationRequest, ch chan<- string) error {
	defer close(ch)

	// Convert to new streaming format
	streamCh := make(chan ai.StreamChunk, 10)
	done := make(chan error, 1)

	go func() {
		done <- s.GenerateStream(ctx, req, streamCh)
	}()

	for {
		select {
		case chunk, ok := <-streamCh:
			if !ok {
				return <-done
			}
			if chunk.Content != "" {
				ch <- chunk.Content
			}
		case err := <-done:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Validate implements code validation (placeholder implementation)
func (s *OpenAIService) Validate(ctx context.Context, code string) (ai.ValidationResult, error) {
	// Simple validation - check if code is not empty and has basic structure
	isValid := strings.TrimSpace(code) != "" && (strings.Contains(code, "func") || strings.Contains(code, "class") || strings.Contains(code, "def"))
	
	var issues []string
	if !isValid {
		issues = append(issues, "Code appears to be empty or malformed")
	}

	return ai.ValidationResult{
		Valid:  isValid,
		Errors: issues,
	}, nil
}

// makeRequest creates and sends an HTTP request to OpenAI API
func (s *OpenAIService) makeRequest(ctx context.Context, reqBody OpenAIRequest) (*http.Response, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// processStreamResponse processes Server-Sent Events from OpenAI streaming API
func (s *OpenAIService) processStreamResponse(ctx context.Context, body io.Reader, ch chan<- ai.StreamChunk) error {
	decoder := json.NewDecoder(body)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var line map[string]interface{}
			if err := decoder.Decode(&line); err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("failed to decode stream line: %w", err)
			}

			// Parse streaming response
			if choices, ok := line["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok && content != "" {
							chunk := ai.StreamChunk{
								Content:    content,
								Model:      s.model, // Use configured model name
								IsComplete: false,
							}
							ch <- chunk
						}
					}
				}
			}

			// Check for completion
			if finishReason, ok := line["finish_reason"]; ok && finishReason != nil {
				chunk := ai.StreamChunk{
					Content:    "",
					Model:      s.model,
					IsComplete: true,
				}
				ch <- chunk
				return nil
			}
		}
	}
}
