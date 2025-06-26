// Package database provides database infrastructure implementations
package database

import (
	"context"
	"fmt"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/user"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/infrastructure/config"
)

// PostgreSQLUserRepository implements the user.Repository interface using GORM
type PostgreSQLUserRepository struct {
	db *gorm.DB
}

// NewPostgreSQLUserRepository creates a new PostgreSQL user repository
func NewPostgreSQLUserRepository(cfg config.DatabaseConfig) (*PostgreSQLUserRepository, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&UserModel{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &PostgreSQLUserRepository{db: db}, nil
}

// Create creates a new user
func (r *PostgreSQLUserRepository) Create(ctx context.Context, u user.User) error {
	userModel := &UserModel{}
	if err := userModel.FromUser(u); err != nil {
		return fmt.Errorf("failed to convert user: %w", err)
	}

	if userModel.ID == "" {
		userModel.ID = generateUserID()
	}

	if err := r.db.WithContext(ctx).Create(userModel).Error; err != nil {
		if isUniqueViolation(err) {
			return common.NewConflictError("user already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *PostgreSQLUserRepository) GetByID(ctx context.Context, id common.UserID) (user.User, error) {
	var userModel UserModel
	if err := r.db.WithContext(ctx).First(&userModel, "id = ?", string(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return user.User{}, common.NewNotFoundError("user not found")
		}
		return user.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userModel.ToUser(), nil
}

// GetByEmail retrieves a user by email
func (r *PostgreSQLUserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	var userModel UserModel
	if err := r.db.WithContext(ctx).First(&userModel, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return user.User{}, common.NewNotFoundError("user not found")
		}
		return user.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userModel.ToUser(), nil
}

// Update updates an existing user
func (r *PostgreSQLUserRepository) Update(ctx context.Context, u user.User) error {
	userModel := &UserModel{}
	if err := userModel.FromUser(u); err != nil {
		return fmt.Errorf("failed to convert user: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", string(u.ID)).Updates(userModel)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return common.NewConflictError("user already exists")
		}
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return common.NewNotFoundError("user not found")
	}
	return nil
}

// Delete deletes a user by ID
func (r *PostgreSQLUserRepository) Delete(ctx context.Context, id common.UserID) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", string(id))
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return common.NewNotFoundError("user not found")
	}
	return nil
}

// List retrieves users with pagination and search
func (r *PostgreSQLUserRepository) List(ctx context.Context, params common.PaginationParams, search string) ([]user.User, error) {
	var userModels []UserModel
	query := r.db.WithContext(ctx).Select("id, email, username, name, avatar_url, roles, role, active, status, created_at, updated_at")
	
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ? OR name ILIKE ?", searchPattern, searchPattern, searchPattern)
	}
	
	if err := query.Order("created_at DESC").Limit(int(params.Limit)).Offset(int(params.Offset())).Find(&userModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]user.User, len(userModels))
	for i, userModel := range userModels {
		domainUser := userModel.ToUser()
		domainUser.PasswordHash = "" // Clear password for security
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
