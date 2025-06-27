package user

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	pb "github.com/EliasRanz/ai-code-gen/api/proto/user"
)

type mockDeleteUserGRPCClient struct{ mockGRPCClient }

func (m *mockDeleteUserGRPCClient) DeleteUser(userID string) (*pb.DeleteUserResponse, error) {
	if userID == "notfound" {
		return &pb.DeleteUserResponse{Success: false, Error: "User not found"}, nil
	}
	return &pb.DeleteUserResponse{Success: true, Error: ""}, nil
}

func (m *mockDeleteUserGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}

func (m *mockDeleteUserGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}

func (m *mockDeleteUserGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}

func (m *mockDeleteUserGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}

func (m *mockDeleteUserGRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	return nil, nil
}

func TestDeleteUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockDeleteUserGRPCClient{})
	r.DELETE("/users/:id", DeleteUserHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/123", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User deleted successfully")
}

func TestDeleteUserHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockDeleteUserGRPCClient{})
	r.DELETE("/users/:id", DeleteUserHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/notfound", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}
