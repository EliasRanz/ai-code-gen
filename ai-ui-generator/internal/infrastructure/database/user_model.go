package database

import (
	"encoding/json"
	"time"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/user"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/common"
)

// UserModel represents the database model for users
type UserModel struct {
	ID           string    `gorm:"primaryKey;column:id" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null;column:email" json:"email"`
	Username     string    `gorm:"uniqueIndex;column:username" json:"username"`
	Name         string    `gorm:"column:name" json:"name"`
	AvatarURL    string    `gorm:"column:avatar_url" json:"avatar_url"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	Roles        string    `gorm:"column:roles" json:"roles"` // Store as JSON string
	Role         string    `gorm:"column:role" json:"role"`
	Active       bool      `gorm:"column:active;default:true" json:"active"`
	Status       string    `gorm:"column:status" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name for the UserModel
func (UserModel) TableName() string {
	return "users"
}

// ToUser converts UserModel to domain user.User
func (u *UserModel) ToUser() user.User {
	var roles []string
	if u.Roles != "" && u.Roles != "[]" {
		_ = json.Unmarshal([]byte(u.Roles), &roles)
	}
	
	domainUser := user.User{
		ID:           common.UserID(u.ID),
		Email:        u.Email,
		Username:     u.Username,
		Name:         u.Name,
		AvatarURL:    u.AvatarURL,
		PasswordHash: u.PasswordHash,
		Roles:        roles,
		Role:         user.Role(u.Role),
		Active:       u.Active,
		Status:       user.UserStatus(u.Status),
	}
	
	// Set timestamps using the common.Timestamps
	domainUser.Timestamps.CreatedAt = u.CreatedAt
	domainUser.Timestamps.UpdatedAt = u.UpdatedAt
	
	return domainUser
}

// FromUser converts domain user.User to UserModel
func (u *UserModel) FromUser(domainUser user.User) error {
	rolesJSON, err := json.Marshal(domainUser.Roles)
	if err != nil {
		return err
	}
	
	u.ID = string(domainUser.ID)
	u.Email = domainUser.Email
	u.Username = domainUser.Username
	u.Name = domainUser.Name
	u.AvatarURL = domainUser.AvatarURL
	u.PasswordHash = domainUser.PasswordHash
	u.Roles = string(rolesJSON)
	u.Role = string(domainUser.Role)
	u.Active = domainUser.Active
	u.Status = string(domainUser.Status)
	u.CreatedAt = domainUser.Timestamps.CreatedAt
	u.UpdatedAt = domainUser.Timestamps.UpdatedAt
	
	return nil
}
