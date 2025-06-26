package user

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	pb "github.com/EliasRanz/ai-code-gen/ai-ui-generator/api/proto/user"
)

type mockGetProjectGRPCClient struct{ mockGRPCClient }

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
	setMockGRPCClientForTest(&mockGetProjectGRPCClient{})
	r.GET("/projects/:id", GetProjectHandler)

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
	setMockGRPCClientForTest(&mockGetProjectGRPCClient{})
	r.GET("/projects/:id", GetProjectHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/projects/notfound", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Project not found")
}
