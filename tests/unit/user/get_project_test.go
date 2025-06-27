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

type mockGetProjectGRPCClient struct{}

func (m *mockGetProjectGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	if projectID == "notfound" {
		return &pb.GetProjectResponse{Project: nil, Error: "Project not found"}, nil
	}
	return &pb.GetProjectResponse{
		Project: &pb.Project{
			Id:          projectID,
			Name:        "Test Project",
			Description: "desc",
			UserId:      "u1",
			Status:      pb.ProjectStatus_PROJECT_STATUS_ACTIVE,
			Tags:        []string{"a"},
			Config:      "{}",
			CreatedAt:   111111,
			UpdatedAt:   222222,
		},
		Error: "",
	}, nil
}

// Implement all required methods for UserGRPCClient interface
func (m *mockGetProjectGRPCClient) UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return nil, nil
}

func (m *mockGetProjectGRPCClient) CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, nil
}

func (m *mockGetProjectGRPCClient) GetUser(userID string) (*pb.GetUserResponse, error) {
	return nil, nil
}

func (m *mockGetProjectGRPCClient) ListUsers(page, limit int32, search string) (*pb.ListUsersResponse, error) {
	return nil, nil
}

func (m *mockGetProjectGRPCClient) DeleteUser(userID string) (*pb.DeleteUserResponse, error) {
	return nil, nil
}

func (m *mockGetProjectGRPCClient) CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	return nil, nil
}

func (m *mockGetProjectGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}

func (m *mockGetProjectGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}

func (m *mockGetProjectGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}

func (m *mockGetProjectGRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	return nil, nil
}

func TestGetProjectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockGetProjectGRPCClient{})
	r.GET("/projects/:id", user.GetProjectHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/projects/p1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Project")
	assert.Contains(t, w.Body.String(), "u1")
}

func TestGetProjectHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockGetProjectGRPCClient{})
	r.GET("/projects/:id", user.GetProjectHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/projects/notfound", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Project not found")
}
