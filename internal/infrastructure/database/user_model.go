package database

import (
	"encoding/json"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
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

// ToUser converts UserModel to domain auth.User
func (u *UserModel) ToUser() auth.User {
	var roles []string
	if u.Roles != "" && u.Roles != "[]" {
		_ = json.Unmarshal([]byte(u.Roles), &roles)
	}

	domainUser := auth.User{
		ID:       auth.UserID(u.ID),
		Email:    u.Email,
		Username: u.Username,
		Name:     u.Name,
		Password: u.PasswordHash, // Map PasswordHash to Password
		Roles:    roles,
		Active:   u.Active,
	}

	return domainUser
}

// FromUser converts domain auth.User to UserModel
func (u *UserModel) FromUser(domainUser auth.User) error {
	rolesJSON, err := json.Marshal(domainUser.Roles)
	if err != nil {
		return err
	}

	u.ID = string(domainUser.ID)
	u.Email = domainUser.Email
	u.Username = domainUser.Username
	u.Name = domainUser.Name
	u.PasswordHash = domainUser.Password // Map Password to PasswordHash
	u.Roles = string(rolesJSON)
	u.Active = domainUser.Active

	return nil
}

// ToDomainUser converts UserModel to domain user.User
func (u *UserModel) ToDomainUser() user.User {
	var roles []string
	if u.Roles != "" {
		json.Unmarshal([]byte(u.Roles), &roles)
	}

	return user.User{
		ID:           common.UserID(u.ID),
		Email:        u.Email,
		Username:     u.Username,
		Name:         u.Name,
		AvatarURL:    u.AvatarURL,
		PasswordHash: u.PasswordHash,
		Roles:        roles,
		Active:       u.Active,
		Status:       user.UserStatus(u.Status),
		Timestamps: common.Timestamps{
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
	}
}

// FromDomainUser converts domain user.User to UserModel
func (u *UserModel) FromDomainUser(domainUser user.User) error {
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
	u.Active = domainUser.Active
	u.Status = string(domainUser.Status)
	u.CreatedAt = domainUser.CreatedAt
	u.UpdatedAt = domainUser.UpdatedAt

	return nil
}
