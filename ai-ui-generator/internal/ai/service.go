package ai

import (
	"strings"

	"golang.org/x/net/html"
)

// ValidationFunc is a function type that defines the signature for validation functions
type ValidationFunc func(code string) (bool, []string, error)

// GenerationHistory stores the prompt and code for AI generations
type GenerationHistory struct {
	Prompt string
	Code   string
}

// GenerationParams holds parameters for AI generation
type GenerationParams struct {
	Model       string   `json:"model,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
}

// Service provides AI generation business logic
type Service struct {
	llmClient    LLMClient
	validateFunc ValidationFunc // optional, for testability

	history map[string][]GenerationHistory // userID -> history
}

// NewService creates a new AI service
func NewService(llmClient LLMClient) *Service {
	return &Service{
		llmClient: llmClient,
		history:   make(map[string][]GenerationHistory),
	}
}

// NewServiceWithValidation creates a new AI service with a custom validation function
func NewServiceWithValidation(llmClient LLMClient, validateFunc ValidationFunc) *Service {
	return &Service{
		llmClient:    llmClient,
		validateFunc: validateFunc,
	}
}

// GenerateCode generates UI code from a prompt
func (s *Service) GenerateCode(prompt string, userID string) (string, error) {
	if s.llmClient == nil {
		return "", nil
	}
	code, err := s.llmClient.Generate(prompt)
	if err == nil && userID != "" {
		s.addToHistory(userID, prompt, code)
	}
	return code, err
}

// GenerateCodeWithParams generates UI code from a prompt with custom parameters
func (s *Service) GenerateCodeWithParams(prompt string, userID string, params GenerationParams) (string, error) {
	if s.llmClient == nil {
		return "", nil
	}

	// For now, use the basic generate method. In a real implementation,
	// you would pass these parameters to the LLM client
	code, err := s.llmClient.Generate(prompt)
	if err == nil && userID != "" {
		s.addToHistory(userID, prompt, code)
	}
	return code, err
}

func (s *Service) addToHistory(userID, prompt, code string) {
	if userID == "" {
		return
	}
	if s.history == nil {
		s.history = make(map[string][]GenerationHistory)
	}
	h := s.history[userID]
	h = append([]GenerationHistory{{Prompt: prompt, Code: code}}, h...)
	if len(h) > 10 {
		h = h[:10]
	}
	s.history[userID] = h
}

func (s *Service) GetHistory(userID string) []GenerationHistory {
	if s.history == nil || userID == "" {
		return nil
	}
	return s.history[userID]
}

// StreamGeneration streams AI generation results
func (s *Service) StreamGeneration(prompt string, userID string, responseChannel chan string) error {
	if s.llmClient == nil {
		return nil
	}
	return s.llmClient.StreamGenerate(prompt, responseChannel)
}

// StreamGenerationWithParams streams AI generation results with custom parameters
func (s *Service) StreamGenerationWithParams(prompt string, userID string, params GenerationParams, responseChannel chan string) error {
	if s.llmClient == nil {
		return nil
	}
	// For now, use the basic stream method. In a real implementation,
	// you would pass these parameters to the LLM client
	return s.llmClient.StreamGenerate(prompt, responseChannel)
}

// ValidateGeneratedCode validates the generated code
func (s *Service) ValidateGeneratedCode(code string) (bool, []string, error) {
	var errors []string
	// Syntax validation (HTML)
	if _, err := html.Parse(strings.NewReader(code)); err != nil {
		errors = append(errors, "HTML syntax error: "+err.Error())
	}
	// Security check: disallow <script> tags
	if strings.Contains(strings.ToLower(code), "<script") {
		errors = append(errors, "Security issue: <script> tags are not allowed")
	}
	valid := len(errors) == 0
	return valid, errors, nil
}
