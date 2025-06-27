package user

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	pb "github.com/EliasRanz/ai-code-gen/api/proto/user"
)

type mockDeleteProjectGRPCClient struct{ mockGRPCClient }

func (m *mockDeleteProjectGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	if projectID == "notfound" {
		return &pb.DeleteProjectResponse{Success: false, Error: "Project not found"}, nil
	}
	return &pb.DeleteProjectResponse{Success: true, Error: ""}, nil
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

func TestDeleteProjectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockDeleteProjectGRPCClient{})
	r.DELETE("/projects/:id", DeleteProjectHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/projects/p1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Project deleted successfully")
}

func TestDeleteProjectHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockDeleteProjectGRPCClient{})
	r.DELETE("/projects/:id", DeleteProjectHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/projects/notfound", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Project not found")
}
