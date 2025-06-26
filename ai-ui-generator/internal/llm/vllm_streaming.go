package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// GenerateStream performs a streaming generation request
func (c *VLLMClient) GenerateStream(ctx context.Context, req *GenerationRequest) (<-chan *GenerationResponse, error) {
	log.Info().
		Str("model", req.Model).
		Str("prompt_preview", truncateString(req.Prompt, 100)).
		Int("max_tokens", req.MaxTokens).
		Msg("VLLM GenerateStream request")

	if c.baseURL == "" {
		// Fallback to stubbed streaming response if no baseURL configured
		return c.generateStubStreamResponse(ctx, req), nil
	}

	// Create VLLM API request
	vllmReq := &vllmRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	requestBody, err := json.Marshal(vllmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/completions", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Create response channel
	responseChan := make(chan *GenerationResponse, 10)

	// Start goroutine to process stream
	go c.processStreamResponse(ctx, resp.Body, responseChan)

	return responseChan, nil
}

// generateStubStreamResponse generates a stubbed streaming response for testing
func (c *VLLMClient) generateStubStreamResponse(ctx context.Context, req *GenerationRequest) <-chan *GenerationResponse {
	responseChan := make(chan *GenerationResponse, 10)

	go func() {
		defer close(responseChan)

		stubbedText := generateStubbedResponse(req.Prompt)
		responseID := generateResponseID()

		// Simulate streaming by sending chunks
		words := strings.Fields(stubbedText)
		for i, word := range words {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Simulate processing time
			time.Sleep(50 * time.Millisecond)

			text := word
			if i < len(words)-1 {
				text += " "
			}

			finishReason := ""
			if i == len(words)-1 {
				finishReason = "stop"
			}

			response := &GenerationResponse{
				ID:      responseID,
				Model:   req.Model,
				Content: text,
				Usage: &TokenUsage{
					PromptTokens:     estimateTokens(req.Prompt),
					CompletionTokens: estimateTokens(text),
				},
				FinishReason: finishReason,
			}

			if finishReason != "" {
				response.Usage.TotalTokens = response.Usage.PromptTokens + response.Usage.CompletionTokens
			}

			select {
			case responseChan <- response:
			case <-ctx.Done():
				return
			}
		}
	}()

	return responseChan
}

// processStreamResponse processes a streaming response from VLLM
func (c *VLLMClient) processStreamResponse(ctx context.Context, body io.ReadCloser, responseChan chan<- *GenerationResponse) {
	defer close(responseChan)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	responseID := generateResponseID()

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove "data: " prefix
		jsonData := strings.TrimPrefix(line, "data: ")
		if jsonData == "[DONE]" {
			break
		}

		var vllmResp vllmResponse
		if err := json.Unmarshal([]byte(jsonData), &vllmResp); err != nil {
			log.Error().Err(err).Str("data", jsonData).Msg("Failed to unmarshal VLLM stream response")
			continue
		}

		if len(vllmResp.Choices) == 0 {
			continue
		}

		choice := vllmResp.Choices[0]
		content := ""
		if choice.Delta != nil && choice.Delta.Content != nil {
			content = *choice.Delta.Content
		} else if choice.Text != "" {
			content = choice.Text
		}

		response := &GenerationResponse{
			ID:      responseID,
			Model:   vllmResp.Model,
			Content: content,
		}

		if choice.FinishReason != nil {
			response.FinishReason = *choice.FinishReason
		}

		if vllmResp.Usage != nil {
			response.Usage = &TokenUsage{
				PromptTokens:     vllmResp.Usage.PromptTokens,
				CompletionTokens: vllmResp.Usage.CompletionTokens,
				TotalTokens:      vllmResp.Usage.TotalTokens,
			}
		}

		select {
		case responseChan <- response:
		case <-ctx.Done():
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Msg("Error reading VLLM stream response")
	}
}
