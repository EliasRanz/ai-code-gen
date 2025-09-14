package auth

import (
	"context"
	"fmt"
	"os"

	"gorm.io/gorm"
)

// PostgreSQLSessionRepository implements the SessionRepository interface using GORM
type PostgreSQLSessionRepository struct {
	db *gorm.DB
}

// NewPostgreSQLSessionRepository creates a new PostgreSQL session repository
func NewPostgreSQLSessionRepository(db *gorm.DB) (*PostgreSQLSessionRepository, error) {
	// Skip auto-migration in CI environment (database is pre-initialized)
	if os.Getenv("ENVIRONMENT") != "ci" {
		// Auto-migrate the schema
		if err := db.AutoMigrate(&Session{}); err != nil {
			return nil, fmt.Errorf("failed to migrate session schema: %w", err)
		}
	}
	return &PostgreSQLSessionRepository{db: db}, nil
}

// Create creates a new session
func (r *PostgreSQLSessionRepository) Create(ctx context.Context, session Session) error {
	if err := r.db.WithContext(ctx).Create(&session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *PostgreSQLSessionRepository) GetByRefreshToken(ctx context.Context, token string) (Session, error) {
	var session Session
	if err := r.db.WithContext(ctx).First(&session, "refresh_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Session{}, fmt.Errorf("session not found")
		}
		return Session{}, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

// Update updates a session
func (r *PostgreSQLSessionRepository) Update(ctx context.Context, session Session) error {
	result := r.db.WithContext(ctx).Model(&Session{}).Where("id = ?", session.ID).Updates(&session)
	if result.Error != nil {
		return fmt.Errorf("failed to update session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

// Delete deletes a session by its ID
func (r *PostgreSQLSessionRepository) Delete(ctx context.Context, sessionID SessionID) error {
	result := r.db.WithContext(ctx).Delete(&Session{}, "id = ?", sessionID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

// DeleteByUserID deletes all sessions for a given user ID
func (r *PostgreSQLSessionRepository) DeleteByUserID(ctx context.Context, userID UserID) error {
	result := r.db.WithContext(ctx).Delete(&Session{}, "user_id = ?", userID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete sessions by user id: %w", result.Error)
	}
	return nil
}

// GetByAccessToken retrieves a session by access token
func (r *PostgreSQLSessionRepository) GetByAccessToken(ctx context.Context, token string) (Session, error) {
	var session Session
	if err := r.db.WithContext(ctx).First(&session, "access_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Session{}, fmt.Errorf("session not found")
		}
		return Session{}, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

// CleanExpired removes expired sessions from the database
func (r *PostgreSQLSessionRepository) CleanExpired(ctx context.Context) error {
	// This is a placeholder implementation.
	// In a real-world scenario, you would delete sessions where the expiry time is in the past.
	return nil
}
