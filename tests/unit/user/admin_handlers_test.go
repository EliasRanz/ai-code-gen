package user_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/EliasRanz/ai-code-gen/internal/user"
)

func setupAdminRouter(handler *user.Handler) *gin.Engine {
	r := gin.Default()
	r.GET("/admin/users", func(c *gin.Context) {
		c.Set("roles", []string{"admin"})
		handler.AdminListUsersHandler(c)
	})
	r.GET("/admin/projects", func(c *gin.Context) {
		c.Set("roles", []string{"admin"})
		handler.AdminListProjectsHandler(c)
	})
	return r
}

func setupNonAdminRouter(handler *user.Handler) *gin.Engine {
	r := gin.Default()
	r.GET("/admin/users", func(c *gin.Context) {
		c.Set("roles", []string{"user"})
		handler.AdminListUsersHandler(c)
	})
	r.GET("/admin/projects", func(c *gin.Context) {
		c.Set("roles", []string{"user"})
		handler.AdminListProjectsHandler(c)
	})
	return r
}

func TestAdminListUsersHandler_AdminAccess(t *testing.T) {
	repo := &mockRepo{}
	projectRepo := &mockProjectRepo{}
	svc := user.NewServiceWithProjects(repo, projectRepo)
	h := user.NewHandler(svc)
	r := setupAdminRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users")
}

func TestAdminListUsersHandler_Forbidden(t *testing.T) {
	repo := &mockRepo{}
	projectRepo := &mockProjectRepo{}
	svc := user.NewServiceWithProjects(repo, projectRepo)
	h := user.NewHandler(svc)
	r := setupNonAdminRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin access required")
}

func TestAdminListProjectsHandler_AdminAccess(t *testing.T) {
	repo := &mockRepo{}
	projectRepo := &mockProjectRepo{}
	svc := user.NewServiceWithProjects(repo, projectRepo)
	h := user.NewHandler(svc)
	r := setupAdminRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/projects", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "projects")
}

func TestAdminListProjectsHandler_Forbidden(t *testing.T) {
	repo := &mockRepo{}
	projectRepo := &mockProjectRepo{}
	svc := user.NewServiceWithProjects(repo, projectRepo)
	h := user.NewHandler(svc)
	r := setupNonAdminRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/projects", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin access required")
}

// mockService implements minimal Service interface for admin tests

type mockService struct{}

func (m *mockService) ListUsers(limit, offset int) ([]*user.User, error) {
	return []*user.User{{ID: "1", Email: "admin@example.com", Name: "Admin", Roles: []string{"admin"}}}, nil
}
func (m *mockService) ListProjects(limit, offset int) ([]*user.Project, error) {
	return []*user.Project{{ID: "1", Name: "Test Project", UserID: "1"}}, nil
}

// mockRepo implements Repository for user listing
// mockProjectRepo implements ProjectRepository for project listing

type mockRepo struct{}
func (m *mockRepo) List(limit, offset int) ([]*user.User, error) {
	return []*user.User{{ID: "1", Email: "admin@example.com", Name: "Admin", Roles: []string{"admin"}}}, nil
}
func (m *mockRepo) GetByEmail(email string) (*user.User, error) { return nil, nil }
// Unused methods for interface compliance
func (m *mockRepo) GetByID(id string) (*user.User, error) { return nil, nil }
func (m *mockRepo) Update(id string, updates map[string]interface{}) (*user.User, error) { return nil, nil }
func (m *mockRepo) Delete(id string) error { return nil }
func (m *mockRepo) Create(user *user.User) error { return nil }


type mockProjectRepo struct{}
func (m *mockProjectRepo) List(limit, offset int) ([]*user.Project, error) {
	return []*user.Project{{ID: "1", Name: "Test Project", UserID: "1"}}, nil
}
// Unused methods for interface compliance
func (m *mockProjectRepo) GetByID(id string) (*user.Project, error) { return nil, nil }
func (m *mockProjectRepo) Update(id string, updates map[string]interface{}) (*user.Project, error) { return nil, nil }
func (m *mockProjectRepo) Delete(id string) error { return nil }
func (m *mockProjectRepo) Create(project *user.Project) error { return nil }
func (m *mockProjectRepo) ListByUserID(userID string, limit, offset int) ([]*user.Project, error) { return nil, nil }
