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

type mockListUserProjectsGRPCClient struct{}

func (m *mockListUserProjectsGRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	if userID == "error" {
		return &pb.ListUserProjectsResponse{Projects: nil, Total: 0, Error: "user error"}, nil
	}
	return &pb.ListUserProjectsResponse{
		Projects: []*pb.Project{
			{Id: "p1", Name: "A", Description: "desc", UserId: userID, Status: pb.ProjectStatus_PROJECT_STATUS_ACTIVE, Tags: []string{"a"}, Config: "{}", CreatedAt: 1, UpdatedAt: 2},
			{Id: "p2", Name: "B", Description: "desc2", UserId: userID, Status: pb.ProjectStatus_PROJECT_STATUS_DRAFT, Tags: []string{"b"}, Config: "{}", CreatedAt: 3, UpdatedAt: 4},
		},
		Total: 2,
		Error: "",
	}, nil
}
func (m *mockListUserProjectsGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}
func (m *mockListUserProjectsGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}
func (m *mockListUserProjectsGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}
func (m *mockListUserProjectsGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}

// Implement all required methods for UserGRPCClient interface
func (m *mockListUserProjectsGRPCClient) UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return nil, nil
}

func (m *mockListUserProjectsGRPCClient) CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, nil
}

func (m *mockListUserProjectsGRPCClient) GetUser(userID string) (*pb.GetUserResponse, error) {
	return nil, nil
}

func (m *mockListUserProjectsGRPCClient) ListUsers(page, limit int32, search string) (*pb.ListUsersResponse, error) {
	return nil, nil
}

func (m *mockListUserProjectsGRPCClient) DeleteUser(userID string) (*pb.DeleteUserResponse, error) {
	return nil, nil
}

func (m *mockListUserProjectsGRPCClient) CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	return nil, nil
}

func TestListUserProjectsHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockListUserProjectsGRPCClient{})
	r.GET("/users/:user_id/projects", user.ListUserProjectsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/u1/projects?page=1&limit=2", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "A")
	assert.Contains(t, w.Body.String(), "B")
	assert.Contains(t, w.Body.String(), "total")
}

func TestListUserProjectsHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockListUserProjectsGRPCClient{})
	r.GET("/users/:user_id/projects", user.ListUserProjectsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/error/projects", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "user error")
}
