package auth

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// GetSessionService handles session retrieval by access token
type GetSessionService struct {
	sessionRepo   SessionRepository
	tokenProvider TokenProvider
}

// NewGetSession creates a new instance of GetSessionService
func NewGetSession(sessionRepo SessionRepository, tokenProvider TokenProvider) *GetSessionService {
	return &GetSessionService{
		sessionRepo:   sessionRepo,
		tokenProvider: tokenProvider,
	}
}

// GetSessionRequest represents the input for session retrieval
type GetSessionRequest struct {
	AccessToken string `json:"access_token" validate:"required"`
}

// GetSessionResponse represents the output of session retrieval
type GetSessionResponse struct {
	SessionID    SessionID `json:"session_id"`
	UserID       UserID    `json:"user_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	Status       string    `json:"status"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
}

// Execute retrieves session information by access token
func (s *GetSessionService) Execute(ctx context.Context, req GetSessionRequest) (*GetSessionResponse, error) {
	// Validate request
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// Get session by access token
	session, err := s.getSessionByAccessToken(ctx, req.AccessToken)
	if err != nil {
		return nil, err
	}

	// Build and return response
	response := s.buildSessionResponse(session)
	return response, nil
}

// validateRequest validates the incoming request
func (s *GetSessionService) validateRequest(req GetSessionRequest) error {
	if strings.TrimSpace(req.AccessToken) == "" {
		return fmt.Errorf("access token is required")
	}
	return nil
}

// getSessionByAccessToken retrieves session by access token
func (s *GetSessionService) getSessionByAccessToken(ctx context.Context, accessToken string) (*Session, error) {
	// First validate the token to ensure it's valid
	_, err := s.tokenProvider.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %v", err)
	}

	// Get session from repository
	session, err := s.sessionRepo.GetByAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// Note: Return session regardless of expiry status
	// The status field will indicate if it's expired
	return &session, nil
}

// buildSessionResponse builds the session response
func (s *GetSessionService) buildSessionResponse(session *Session) *GetSessionResponse {
	// Use CreatedAt as fallback for LastAccessed if not set
	lastAccessed := session.LastAccessed
	if lastAccessed.IsZero() {
		lastAccessed = session.CreatedAt
	}

	return &GetSessionResponse{
		SessionID:    session.ID,
		UserID:       session.UserID,
		ExpiresAt:    session.ExpiresAt,
		CreatedAt:    session.CreatedAt,
		LastAccessed: lastAccessed,
		Status:       string(session.Status),
		IPAddress:    session.IPAddress,
		UserAgent:    session.UserAgent,
	}
}
