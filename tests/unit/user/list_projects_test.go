package user

import (
	"net/http"
	"net/http/httptest"
	"testing"
	pb "github.com/EliasRanz/ai-code-gen/api/proto/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockListProjectsGRPCClient struct{ mockGRPCClient }

func (m *mockListProjectsGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	if search == "error" {
		return &pb.ListProjectsResponse{Projects: nil, Total: 0, Error: "search error"}, nil
	}
	return &pb.ListProjectsResponse{
		Projects: []*pb.Project{
			{Id: "p1", Name: "A", Description: "desc", UserId: "u1", Status: pb.ProjectStatus_PROJECT_STATUS_ACTIVE, Tags: []string{"a"}, Config: "{}", CreatedAt: 1, UpdatedAt: 2},
			{Id: "p2", Name: "B", Description: "desc2", UserId: "u2", Status: pb.ProjectStatus_PROJECT_STATUS_DRAFT, Tags: []string{"b"}, Config: "{}", CreatedAt: 3, UpdatedAt: 4},
		},
		Total: 2,
		Error: "",
	}, nil
}
func (m *mockListProjectsGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}
func (m *mockListProjectsGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}
func (m *mockListProjectsGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}

func TestListProjectsHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockListProjectsGRPCClient{})
	r.GET("/projects", ListProjectsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/projects?page=1&limit=2", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "A")
	assert.Contains(t, w.Body.String(), "B")
	assert.Contains(t, w.Body.String(), "total")
}

func TestListProjectsHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockListProjectsGRPCClient{})
	r.GET("/projects", ListProjectsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/projects?search=error", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "search error")
}
