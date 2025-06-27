package user

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/EliasRanz/ai-code-gen/internal/user"
	pb "github.com/EliasRanz/ai-code-gen/api/proto/user"
)

type mockDeleteProjectGRPCClient struct{}

func (m *mockDeleteProjectGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	if projectID == "notfound" {
		return &pb.DeleteProjectResponse{Success: false, Error: "Project not found"}, nil
	}
	return &pb.DeleteProjectResponse{Success: true, Error: ""}, nil
}

// Implement all required methods for UserGRPCClient interface
func (m *mockDeleteProjectGRPCClient) UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return nil, nil
}

func (m *mockDeleteProjectGRPCClient) CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, nil
}

func (m *mockDeleteProjectGRPCClient) GetUser(userID string) (*pb.GetUserResponse, error) {
	return nil, nil
}

func (m *mockDeleteProjectGRPCClient) ListUsers(page, limit int32, search string) (*pb.ListUsersResponse, error) {
	return nil, nil
}

func (m *mockDeleteProjectGRPCClient) DeleteUser(userID string) (*pb.DeleteUserResponse, error) {
	return nil, nil
}

func (m *mockDeleteProjectGRPCClient) CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	return nil, nil
}

func (m *mockDeleteProjectGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}

func (m *mockDeleteProjectGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}

func (m *mockDeleteProjectGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}

func (m *mockDeleteProjectGRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	return nil, nil
}

func TestDeleteProjectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockDeleteProjectGRPCClient{})
	r.DELETE("/projects/:id", user.DeleteProjectHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/projects/p1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Project deleted successfully")
}

func TestDeleteProjectHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockDeleteProjectGRPCClient{})
	r.DELETE("/projects/:id", user.DeleteProjectHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/projects/notfound", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Project not found")
}
