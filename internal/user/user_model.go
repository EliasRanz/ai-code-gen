package user

import (
	"encoding/json"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// UserModel represents the database model for users
type UserModel struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null;column:email" json:"email"`
	Username  string    `gorm:"uniqueIndex;column:username" json:"username"`
	Name      string    `gorm:"column:name" json:"name"`
	AvatarURL string    `gorm:"column:avatar_url" json:"avatar_url"`
	Roles     string    `gorm:"column:roles" json:"roles"` // Store as JSON string
	Role      string    `gorm:"column:role" json:"role"`
	Active    bool      `gorm:"column:active;default:true" json:"active"`
	Status    string    `gorm:"column:status" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name for the UserModel
func (UserModel) TableName() string {
	return "users"
}

// ToDomainUser converts UserModel to domain User
func (u *UserModel) ToDomainUser() User {
	var roles []string
	if u.Roles != "" {
		json.Unmarshal([]byte(u.Roles), &roles)
	}

	return User{
		ID:        utilities.UserID(u.ID),
		Email:     u.Email,
		Username:  u.Username,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		Roles:     roles,
		Active:    u.Active,
		Status:    UserStatus(u.Status),
		Timestamps: utilities.Timestamps{
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
	}
}

// FromDomainUser converts domain User to UserModel
func (u *UserModel) FromDomainUser(domainUser User) error {
	rolesJSON, err := json.Marshal(domainUser.Roles)
	if err != nil {
		return err
	}

	u.ID = string(domainUser.ID)
	u.Email = domainUser.Email
	u.Username = domainUser.Username
	u.Name = domainUser.Name
	u.AvatarURL = domainUser.AvatarURL
	u.Roles = string(rolesJSON)
	u.Active = domainUser.Active
	u.Status = string(domainUser.Status)
	u.CreatedAt = domainUser.CreatedAt
	u.UpdatedAt = domainUser.UpdatedAt

	return nil
}
