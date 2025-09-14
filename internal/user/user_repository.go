// Package user provides user domain and infrastructure
package user

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// PostgreSQLUserRepository implements the user.Repository interface using GORM
type PostgreSQLUserRepository struct {
	db *gorm.DB
}

// NewPostgreSQLUserRepository creates a new PostgreSQL user repository
func NewPostgreSQLUserRepository(db *gorm.DB) (*PostgreSQLUserRepository, error) {
	// Skip auto-migration in CI environment (database is pre-initialized)
	if os.Getenv("ENVIRONMENT") != "ci" {
		// Auto-migrate the schema
		if err := db.AutoMigrate(&UserModel{}); err != nil {
			return nil, fmt.Errorf("failed to migrate schema: %w", err)
		}
	}

	return &PostgreSQLUserRepository{db: db}, nil
}

// Create creates a new user
func (r *PostgreSQLUserRepository) Create(ctx context.Context, u User) error {
	userModel := &UserModel{}
	if err := userModel.FromDomainUser(u); err != nil {
		return fmt.Errorf("failed to convert user: %w", err)
	}

	if userModel.ID == "" {
		userModel.ID = generateUserID()
	}

	if err := r.db.WithContext(ctx).Create(userModel).Error; err != nil {
		if isUniqueViolation(err) {
			return utilities.NewConflictError("user already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *PostgreSQLUserRepository) GetByID(ctx context.Context, id utilities.UserID) (User, error) {
	var userModel UserModel
	if err := r.db.WithContext(ctx).First(&userModel, "id = ?", string(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return User{}, utilities.NewNotFoundError("user not found")
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userModel.ToDomainUser(), nil
}

// GetByEmail retrieves a user by email
func (r *PostgreSQLUserRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	var userModel UserModel
	if err := r.db.WithContext(ctx).First(&userModel, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return User{}, utilities.NewNotFoundError("user not found")
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userModel.ToDomainUser(), nil
}

// Update updates an existing user
func (r *PostgreSQLUserRepository) Update(ctx context.Context, u User) error {
	userModel := &UserModel{}
	if err := userModel.FromDomainUser(u); err != nil {
		return fmt.Errorf("failed to convert user: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", string(u.ID)).Updates(userModel)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return utilities.NewConflictError("user already exists")
		}
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utilities.NewNotFoundError("user not found")
	}
	return nil
}

// Delete deletes a user by ID
func (r *PostgreSQLUserRepository) Delete(ctx context.Context, id utilities.UserID) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", string(id))
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utilities.NewNotFoundError("user not found")
	}
	return nil
}

// List retrieves users with pagination and search
func (r *PostgreSQLUserRepository) List(ctx context.Context, params utilities.PaginationParams, search string) ([]User, error) {
	var userModels []UserModel
	query := r.db.WithContext(ctx).Select("id, email, username, name, avatar_url, roles, role, active, status, created_at, updated_at")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ? OR name ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Order("created_at DESC").Limit(int(params.Limit)).Offset(int(params.Offset())).Find(&userModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]User, len(userModels))
	for i, userModel := range userModels {
		domainUser := userModel.ToDomainUser()
		users[i] = domainUser
	}

	return users, nil
}

// Count returns the total number of users matching the search criteria
func (r *PostgreSQLUserRepository) Count(ctx context.Context, search string) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&UserModel{})

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
func (r *PostgreSQLUserRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Utility functions for database operations
func generateUserID() string {
	// Simple ID generation - in production use UUID or similar
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique") ||
		strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
