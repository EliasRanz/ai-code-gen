package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	pb "github.com/EliasRanz/ai-code-gen/api/proto/user"
)

type mockUpdateProjectGRPCClient struct{ mockGRPCClient }

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

func TestUpdateProjectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockUpdateProjectGRPCClient{})
	r.PUT("/projects/:id", UpdateProjectHandler)

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
	setMockGRPCClientForTest(&mockUpdateProjectGRPCClient{})
	r.PUT("/projects/:id", UpdateProjectHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Name"}`
	req, _ := http.NewRequest("PUT", "/projects/notfound", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Project not found")
}
