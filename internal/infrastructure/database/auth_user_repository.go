// Package database provides database infrastructure implementations
package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
)

// AuthUserRepository implements the auth.UserRepository interface using GORM
// This is specifically for auth-related user operations with simplified types
type AuthUserRepository struct {
	db *gorm.DB
}

// NewAuthUserRepository creates a new auth-specific user repository
func NewAuthUserRepository(db *gorm.DB) (*AuthUserRepository, error) {
	// Auto-migrate the schema (using the same UserModel as the domain repo)
	if err := db.AutoMigrate(&UserModel{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &AuthUserRepository{db: db}, nil
}

// Create creates a new user
func (r *AuthUserRepository) Create(ctx context.Context, u auth.User) error {
	userModel := &UserModel{}
	if err := userModel.FromUser(u); err != nil {
		return fmt.Errorf("failed to convert user: %w", err)
	}

	if userModel.ID == "" {
		userModel.ID = generateUserID()
	}

	if err := r.db.WithContext(ctx).Create(userModel).Error; err != nil {
		if isUniqueViolation(err) {
			return auth.NewConflictError("user already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *AuthUserRepository) GetByID(ctx context.Context, id auth.UserID) (auth.User, error) {
	var userModel UserModel
	if err := r.db.WithContext(ctx).First(&userModel, "id = ?", string(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return auth.User{}, auth.NewNotFoundError("user not found")
		}
		return auth.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userModel.ToUser(), nil
}

// GetByEmail retrieves a user by email
func (r *AuthUserRepository) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	var userModel UserModel
	if err := r.db.WithContext(ctx).First(&userModel, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return auth.User{}, auth.NewNotFoundError("user not found")
		}
		return auth.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userModel.ToUser(), nil
}

// Update updates an existing user
func (r *AuthUserRepository) Update(ctx context.Context, u auth.User) error {
	userModel := &UserModel{}
	if err := userModel.FromUser(u); err != nil {
		return fmt.Errorf("failed to convert user: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", string(u.ID)).Updates(userModel)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return auth.NewConflictError("user already exists")
		}
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return auth.NewNotFoundError("user not found")
	}
	return nil
}

// Delete deletes a user by ID
func (r *AuthUserRepository) Delete(ctx context.Context, id auth.UserID) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", string(id))
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return auth.NewNotFoundError("user not found")
	}
	return nil
}

// List retrieves users with pagination and search
func (r *AuthUserRepository) List(ctx context.Context, params auth.PaginationParams, search string) ([]auth.User, error) {
	var userModels []UserModel
	query := r.db.WithContext(ctx).Select("id, email, username, name, avatar_url, roles, role, active, status, created_at, updated_at")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ? OR name ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Order("created_at DESC").Limit(int(params.Limit)).Offset(int(params.Offset())).Find(&userModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]auth.User, len(userModels))
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
func (r *AuthUserRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
