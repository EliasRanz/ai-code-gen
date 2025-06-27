package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/EliasRanz/ai-code-gen/internal/user"
	pb "github.com/EliasRanz/ai-code-gen/api/proto/user"
)

type mockUpdateUserProfileGRPCClient struct{}

func (m *mockUpdateUserProfileGRPCClient) UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.Id == "notfound" {
		return &pb.UpdateUserResponse{User: nil, Error: "User not found"}, nil
	}
	return &pb.UpdateUserResponse{
		User: &pb.User{
			Id:        req.Id,
			Email:    "test@example.com",
			Name:     req.Name,
			AvatarUrl: req.AvatarUrl,
			Roles:    []string{"user"},
			CreatedAt: 111111,
			UpdatedAt: 222222,
		},
		Error: "",
	}, nil
}

func (m *mockUpdateUserProfileGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	return nil, nil
}

// Implement all required methods for UserGRPCClient interface
func (m *mockUpdateUserProfileGRPCClient) CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) GetUser(userID string) (*pb.GetUserResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) ListUsers(page, limit int32, search string) (*pb.ListUsersResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) DeleteUser(userID string) (*pb.DeleteUserResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	return nil, nil
}

func TestUpdateUserProfileHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockUpdateUserProfileGRPCClient{})
	r.PUT("/users/:id", user.UpdateUserHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Updated Name", "avatar_url": "http://avatar"}`
	req, _ := http.NewRequest("PUT", "/users/123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Updated Name")
}

func TestUpdateUserProfileHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockUpdateUserProfileGRPCClient{})
	r.PUT("/users/:id", user.UpdateUserHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Name"}`
	req, _ := http.NewRequest("PUT", "/users/notfound", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}
