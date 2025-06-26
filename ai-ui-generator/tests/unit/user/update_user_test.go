package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	pb "github.com/ai-code-gen/ai-ui-generator/api/proto/user"
)

type mockGRPCClient struct{}

func (m *mockGRPCClient) UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.Id == "notfound" {
		return &pb.UpdateUserResponse{User: nil, Error: "User not found"}, nil
	}
	return &pb.UpdateUserResponse{
		User: &pb.User{
			Id:        req.Id,
			Email:    "test@example.com",
			Name:     req.Name,
			AvatarUrl: req.AvatarUrl,
			Roles:    req.Roles,
			CreatedAt: 111111,
			UpdatedAt: 222222,
		},
		Error: "",
	}, nil
}

func (m *mockGRPCClient) CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	return nil, nil
}
func (m *mockGRPCClient) ListUsers(page, limit int32, search string) (*pb.ListUsersResponse, error) {
	return nil, nil
}
func (m *mockGRPCClient) CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, nil
}
func (m *mockGRPCClient) GetUser(userID string) (*pb.GetUserResponse, error) {
	return nil, nil
}
func (m *mockGRPCClient) DeleteUser(userID string) (*pb.DeleteUserResponse, error) {
	return &pb.DeleteUserResponse{Success: true, Error: ""}, nil
}
func (m *mockGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}
func (m *mockGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}
func (m *mockGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}
func (m *mockGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}
func (m *mockGRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	return nil, nil
}

// Use interface for grpcClient
var _ interface{ UpdateUser(*pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) } = &mockGRPCClient{}

// Patch SetGRPCClient to accept any UserGRPCClient for testing
func setMockGRPCClientForTest(m UserGRPCClient) {
	grpcClient = m
}

func TestUpdateUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockGRPCClient{})
	r.PUT("/users/:id", UpdateUserHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Updated Name", "avatar_url": "http://avatar", "roles": ["user"]}`
	req, _ := http.NewRequest("PUT", "/users/123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Updated Name")
}

func TestUpdateUserHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockGRPCClient{})
	r.PUT("/users/:id", UpdateUserHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Name"}`
	req, _ := http.NewRequest("PUT", "/users/notfound", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}
