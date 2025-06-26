package user

import (
	"fmt"
	"gorm.io/gorm"
)

// Repository defines the interface for user data access
type Repository interface {
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(id string, updates map[string]interface{}) (*User, error)
	Delete(id string) error
	List(limit, offset int) ([]*User, error)
}

// GormRepository implements Repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM user repository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// GetByID retrieves a user by ID
func (r *GormRepository) GetByID(id string) (*User, error) {
	var userModel UserModel
	if err := r.db.First(&userModel, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return userModel.ToUser(), nil
}

// GetByEmail retrieves a user by email
func (r *GormRepository) GetByEmail(email string) (*User, error) {
	var userModel UserModel
	if err := r.db.First(&userModel, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return userModel.ToUser(), nil
}

// Create creates a new user
func (r *GormRepository) Create(user *User) error {
	userModel := &UserModel{}
	if err := userModel.FromUser(user); err != nil {
		return fmt.Errorf("failed to convert user: %w", err)
	}
	
	if err := r.db.Create(userModel).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	// Update the original user with generated values
	*user = *userModel.ToUser()
	return nil
}

// Update updates a user
func (r *GormRepository) Update(id string, updates map[string]interface{}) (*User, error) {
	if len(updates) == 0 {
		return r.GetByID(id) // No updates to perform
	}

	// Perform the update
	result := r.db.Model(&UserModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update user: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user not found")
	}
	
	// Return the updated user
	return r.GetByID(id)
}

// Delete deletes a user
func (r *GormRepository) Delete(id string) error {
	result := r.db.Delete(&UserModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// List lists users with pagination
func (r *GormRepository) List(limit, offset int) ([]*User, error) {
	var userModels []UserModel
	if err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&userModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	
	users := make([]*User, len(userModels))
	for i, userModel := range userModels {
		users[i] = userModel.ToUser()
	}
	
	return users, nil
}