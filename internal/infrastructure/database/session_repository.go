package database

import (
	"context"
	"fmt"

	"github.com/EliasRanz/ai-code-gen/internal/domain/auth"
	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"gorm.io/gorm"
)

// PostgreSQLSessionRepository implements the auth.SessionRepository interface using GORM
type PostgreSQLSessionRepository struct {
	db *gorm.DB
}

// NewPostgreSQLSessionRepository creates a new PostgreSQL session repository
func NewPostgreSQLSessionRepository(db *gorm.DB) (*PostgreSQLSessionRepository, error) {
	// Auto-migrate the schema
	if err := db.AutoMigrate(&auth.Session{}); err != nil {
		return nil, fmt.Errorf("failed to migrate session schema: %w", err)
	}
	return &PostgreSQLSessionRepository{db: db}, nil
}

// Create creates a new session
func (r *PostgreSQLSessionRepository) Create(ctx context.Context, session auth.Session) error {
	if err := r.db.WithContext(ctx).Create(&session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *PostgreSQLSessionRepository) GetByRefreshToken(ctx context.Context, token string) (auth.Session, error) {
	var session auth.Session
	if err := r.db.WithContext(ctx).First(&session, "refresh_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return auth.Session{}, fmt.Errorf("session not found")
		}
		return auth.Session{}, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

// Update updates a session
func (r *PostgreSQLSessionRepository) Update(ctx context.Context, session auth.Session) error {
	result := r.db.WithContext(ctx).Model(&auth.Session{}).Where("id = ?", session.ID).Updates(&session)
	if result.Error != nil {
		return fmt.Errorf("failed to update session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

// Delete deletes a session by its ID
func (r *PostgreSQLSessionRepository) Delete(ctx context.Context, sessionID common.SessionID) error {
	result := r.db.WithContext(ctx).Delete(&auth.Session{}, "id = ?", sessionID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

// DeleteByUserID deletes all sessions for a given user ID
func (r *PostgreSQLSessionRepository) DeleteByUserID(ctx context.Context, userID common.UserID) error {
	result := r.db.WithContext(ctx).Delete(&auth.Session{}, "user_id = ?", userID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete sessions by user id: %w", result.Error)
	}
	return nil
}

// GetByAccessToken retrieves a session by access token
func (r *PostgreSQLSessionRepository) GetByAccessToken(ctx context.Context, token string) (auth.Session, error) {
	var session auth.Session
	if err := r.db.WithContext(ctx).First(&session, "access_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return auth.Session{}, fmt.Errorf("session not found")
		}
		return auth.Session{}, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

// CleanExpired removes expired sessions from the database
func (r *PostgreSQLSessionRepository) CleanExpired(ctx context.Context) error {
	// This is a placeholder implementation.
	// In a real-world scenario, you would delete sessions where the expiry time is in the past.
	return nil
}
