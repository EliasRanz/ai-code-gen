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

type mockUpdateProjectGRPCClient struct{}

func (m *mockUpdateProjectGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	if req.Id == "notfound" {
		return &pb.UpdateProjectResponse{Project: nil, Error: "Project not found"}, nil
	}
	return &pb.UpdateProjectResponse{
		Project: &pb.Project{
			Id:          req.Id,
			Name:        req.Name,
			Description: req.Description,
			UserId:      "u1",
			Status:      req.Status,
			Tags:        req.Tags,
			Config:      req.Config,
			CreatedAt:   111111,
			UpdatedAt:   222222,
		},
		Error: "",
	}, nil
}
func (m *mockUpdateProjectGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}

// Implement all required methods for UserGRPCClient interface
func (m *mockUpdateProjectGRPCClient) UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return nil, nil
}

func (m *mockUpdateProjectGRPCClient) CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, nil
}

func (m *mockUpdateProjectGRPCClient) GetUser(userID string) (*pb.GetUserResponse, error) {
	return nil, nil
}

func (m *mockUpdateProjectGRPCClient) ListUsers(page, limit int32, search string) (*pb.ListUsersResponse, error) {
	return nil, nil
}

func (m *mockUpdateProjectGRPCClient) DeleteUser(userID string) (*pb.DeleteUserResponse, error) {
	return nil, nil
}

func (m *mockUpdateProjectGRPCClient) CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	return nil, nil
}

func (m *mockUpdateProjectGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}

func (m *mockUpdateProjectGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}

func (m *mockUpdateProjectGRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	return nil, nil
}

func TestUpdateProjectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockUpdateProjectGRPCClient{})
	r.PUT("/projects/:id", user.UpdateProjectHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Updated Project", "description": "desc", "status": "completed", "tags": ["a"], "config": "{}"}`
	req, _ := http.NewRequest("PUT", "/projects/p1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Updated Project")
}

func TestUpdateProjectHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	user.SetGRPCClient(&mockUpdateProjectGRPCClient{})
	r.PUT("/projects/:id", user.UpdateProjectHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Name"}`
	req, _ := http.NewRequest("PUT", "/projects/notfound", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Project not found")
}
