package user

import (
	"encoding/json"
	"time"
)

// UserModel represents the GORM model for users
type UserModel struct {
	ID            string     `gorm:"primaryKey;column:id" json:"id"`
	Email         string     `gorm:"uniqueIndex;not null;column:email" json:"email"`
	Name          string     `gorm:"column:name" json:"name"`
	AvatarURL     string     `gorm:"column:avatar_url" json:"avatar_url"`
	Roles         string     `gorm:"column:roles;type:jsonb" json:"roles"` // Store as JSON
	IsActive      bool       `gorm:"column:is_active;default:true" json:"is_active"`
	EmailVerified bool       `gorm:"column:email_verified;default:false" json:"email_verified"`
	PasswordHash  string     `gorm:"column:password_hash" json:"-"`
	LastLoginAt   *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name for UserModel
func (UserModel) TableName() string {
	return "users"
}

// ToUser converts UserModel to User
func (u *UserModel) ToUser() *User {
	var roles []string
	if u.Roles != "" && u.Roles != "[]" {
		_ = json.Unmarshal([]byte(u.Roles), &roles)
	}
	
	return &User{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		AvatarURL:     u.AvatarURL,
		Roles:         roles,
		IsActive:      u.IsActive,
		EmailVerified: u.EmailVerified,
		PasswordHash:  u.PasswordHash,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// FromUser converts User to UserModel
func (u *UserModel) FromUser(user *User) error {
	rolesJSON, err := json.Marshal(user.Roles)
	if err != nil {
		return err
	}
	
	u.ID = user.ID
	u.Email = user.Email
	u.Name = user.Name
	u.AvatarURL = user.AvatarURL
	u.Roles = string(rolesJSON)
	u.IsActive = user.IsActive
	u.EmailVerified = user.EmailVerified
	u.PasswordHash = user.PasswordHash
	u.LastLoginAt = user.LastLoginAt
	u.CreatedAt = user.CreatedAt
	u.UpdatedAt = user.UpdatedAt
	
	return nil
}

// ProjectModel represents the GORM model for projects
type ProjectModel struct {
	ID          string     `gorm:"primaryKey;column:id" json:"id"`
	Name        string     `gorm:"column:name;not null" json:"name"`
	Description string     `gorm:"column:description" json:"description"`
	UserID      string     `gorm:"column:user_id;not null;index" json:"user_id"`
	Status      string     `gorm:"column:status;default:'draft'" json:"status"`
	Tags        string     `gorm:"column:tags;type:jsonb" json:"tags"` // Store as JSON
	Config      string     `gorm:"column:config;type:jsonb" json:"config"` // Store as JSON
	Metadata    string     `gorm:"column:metadata;type:jsonb" json:"metadata"` // Store as JSON
	IsPublic    bool       `gorm:"column:is_public;default:false" json:"is_public"`
	TemplateID  *string    `gorm:"column:template_id" json:"template_id,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name for ProjectModel
func (ProjectModel) TableName() string {
	return "projects"
}

// ToProject converts ProjectModel to Project
func (p *ProjectModel) ToProject() *Project {
	var tags []string
	var config, metadata map[string]interface{}
	
	if p.Tags != "" && p.Tags != "[]" {
		_ = json.Unmarshal([]byte(p.Tags), &tags)
	}
	if p.Config != "" && p.Config != "{}" {
		_ = json.Unmarshal([]byte(p.Config), &config)
	}
	if p.Metadata != "" && p.Metadata != "{}" {
		_ = json.Unmarshal([]byte(p.Metadata), &metadata)
	}
	
	return &Project{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		UserID:      p.UserID,
		Status:      p.Status,
		Tags:        tags,
		Config:      config,
		Metadata:    metadata,
		IsPublic:    p.IsPublic,
		TemplateID:  p.TemplateID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// FromProject converts Project to ProjectModel
func (p *ProjectModel) FromProject(project *Project) error {
	tagsJSON, err := json.Marshal(project.Tags)
	if err != nil {
		return err
	}
	configJSON, err := json.Marshal(project.Config)
	if err != nil {
		return err
	}
	metadataJSON, err := json.Marshal(project.Metadata)
	if err != nil {
		return err
	}
	
	p.ID = project.ID
	p.Name = project.Name
	p.Description = project.Description
	p.UserID = project.UserID
	p.Status = project.Status
	p.Tags = string(tagsJSON)
	p.Config = string(configJSON)
	p.Metadata = string(metadataJSON)
	p.IsPublic = project.IsPublic
	p.TemplateID = project.TemplateID
	p.CreatedAt = project.CreatedAt
	p.UpdatedAt = project.UpdatedAt
	
	return nil
}
