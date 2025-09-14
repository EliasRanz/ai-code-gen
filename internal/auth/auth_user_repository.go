// Package auth provides authentication domain and infrastructure
package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AuthUserRepository implements the auth.UserRepository interface using GORM
// This is specifically for auth-related user operations with simplified types
type AuthUserRepository struct {
	db *gorm.DB
}

// NewAuthUserRepository creates a new auth-specific user repository
func NewAuthUserRepository(db *gorm.DB) (*AuthUserRepository, error) {
	// Skip auto-migration in CI environment (database is pre-initialized)
	if os.Getenv("ENVIRONMENT") != "ci" {
		// Auto-migrate the schema (using AuthUserModel for auth operations)
		if err := db.AutoMigrate(&AuthUserModel{}); err != nil {
			return nil, fmt.Errorf("failed to migrate schema: %w", err)
		}
	}

	return &AuthUserRepository{db: db}, nil
}

// Create creates a new user
func (r *AuthUserRepository) Create(ctx context.Context, u User) error {
	userModel := &AuthUserModel{}
	if err := userModel.FromUser(u); err != nil {
		return fmt.Errorf("failed to convert user: %w", err)
	}

	if userModel.ID == "" {
		userModel.ID = generateUserID()
	}

	if err := r.db.WithContext(ctx).Create(userModel).Error; err != nil {
		if isUniqueViolation(err) {
			return NewConflictError("user already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *AuthUserRepository) GetByID(ctx context.Context, id UserID) (User, error) {
	var userModel AuthUserModel
	if err := r.db.WithContext(ctx).First(&userModel, "id = ?", string(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return User{}, NewNotFoundError("user not found")
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userModel.ToUser(), nil
}

// GetByEmail retrieves a user by email
func (r *AuthUserRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	var userModel AuthUserModel
	if err := r.db.WithContext(ctx).First(&userModel, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return User{}, NewNotFoundError("user not found")
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userModel.ToUser(), nil
}

// Update updates an existing user
func (r *AuthUserRepository) Update(ctx context.Context, u User) error {
	userModel := &AuthUserModel{}
	if err := userModel.FromUser(u); err != nil {
		return fmt.Errorf("failed to convert user: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&AuthUserModel{}).Where("id = ?", string(u.ID)).Updates(userModel)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return NewConflictError("user already exists")
		}
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return NewNotFoundError("user not found")
	}
	return nil
}

// Delete deletes a user by ID
func (r *AuthUserRepository) Delete(ctx context.Context, id UserID) error {
	result := r.db.WithContext(ctx).Delete(&AuthUserModel{}, "id = ?", string(id))
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return NewNotFoundError("user not found")
	}
	return nil
}

// List retrieves users with pagination and search
func (r *AuthUserRepository) List(ctx context.Context, params PaginationParams, search string) ([]User, error) {
	var userModels []AuthUserModel
	query := r.db.WithContext(ctx).Select("id, email, username, name, password_hash, roles, active, created_at, updated_at")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ? OR name ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Order("created_at DESC").Limit(int(params.Limit)).Offset(int(params.Offset())).Find(&userModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]User, len(userModels))
	for i, userModel := range userModels {
		domainUser := userModel.ToUser()
		domainUser.Password = "" // Clear password for security
		users[i] = domainUser
	}

	return users, nil
}

// Count returns the total number of users matching the search criteria
func (r *AuthUserRepository) Count(ctx context.Context, search string) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&AuthUserModel{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ? OR name ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return int(count), nil
}

// Close closes the database connection
func (r *AuthUserRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Utility functions for auth database operations
func generateUserID() string {
	// Simple ID generation - in production use UUID or similar
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique") ||
		strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
