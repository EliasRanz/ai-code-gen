package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	pb "github.com/EliasRanz/ai-code-gen/ai-ui-generator/api/proto/user"
)

type mockCreateProjectGRPCClient struct{ mockGRPCClient }

func (m *mockCreateProjectGRPCClient) CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	if req.Name == "fail" {
		return &pb.CreateProjectResponse{Project: nil, Error: "creation error"}, nil
	}
	return &pb.CreateProjectResponse{
		Project: &pb.Project{
			Id:          "p1",
			Name:        req.Name,
			Description: req.Description,
			UserId:      req.UserId,
			Status:      pb.ProjectStatus_PROJECT_STATUS_ACTIVE,
			Tags:        req.Tags,
			Config:      req.Config,
			CreatedAt:   111111,
			UpdatedAt:   222222,
		},
		Error: "",
	}, nil
}

func (m *mockCreateProjectGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}

func (m *mockCreateProjectGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}

func (m *mockCreateProjectGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}

func (m *mockCreateProjectGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}

func TestCreateProjectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockCreateProjectGRPCClient{})
	r.POST("/projects", CreateProjectHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Test Project", "user_id": "u1", "description": "desc", "tags": ["a"], "config": "{}"}`
	req, _ := http.NewRequest("POST", "/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Test Project")
	assert.Contains(t, w.Body.String(), "u1")
}

func TestCreateProjectHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockCreateProjectGRPCClient{})
	r.POST("/projects", CreateProjectHandler)

	w := httptest.NewRecorder()
	body := `{"name": "fail", "user_id": "u1"}`
	req, _ := http.NewRequest("POST", "/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "creation error")
}
