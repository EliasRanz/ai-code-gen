package auth

import (
	"encoding/json"
	"time"
)

// AuthUserModel represents the database model for auth-specific user operations
type AuthUserModel struct {
	ID           string    `gorm:"primaryKey;column:id" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null;column:email" json:"email"`
	Username     string    `gorm:"uniqueIndex;column:username" json:"username"`
	Name         string    `gorm:"column:name" json:"name"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	Roles        string    `gorm:"column:roles" json:"roles"` // Store as JSON string
	Active       bool      `gorm:"column:active;default:true" json:"active"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name for the AuthUserModel
func (AuthUserModel) TableName() string {
	return "users"
}

// ToUser converts AuthUserModel to domain User
func (u *AuthUserModel) ToUser() User {
	var roles []string
	if u.Roles != "" && u.Roles != "[]" {
		_ = json.Unmarshal([]byte(u.Roles), &roles)
	}

	return User{
		ID:       UserID(u.ID),
		Email:    u.Email,
		Username: u.Username,
		Name:     u.Name,
		Password: u.PasswordHash, // Map PasswordHash to Password
		Roles:    roles,
		Active:   u.Active,
	}
}

// FromUser converts domain User to AuthUserModel
func (u *AuthUserModel) FromUser(domainUser User) error {
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
