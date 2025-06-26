package user

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	pb "github.com/EliasRanz/ai-code-gen/ai-ui-generator/api/proto/user"
)

type mockGetUserGRPCClient struct{ mockGRPCClient }

func (m *mockGetUserGRPCClient) GetUser(userID string) (*pb.GetUserResponse, error) {
	if userID == "notfound" {
		return &pb.GetUserResponse{User: nil, Error: "User not found"}, nil
	}
	return &pb.GetUserResponse{
		User: &pb.User{
			Id:        userID,
			Email:    "test@example.com",
			Name:     "Test User",
			AvatarUrl: "http://avatar",
			Roles:    []string{"user"},
			CreatedAt: 111111,
			UpdatedAt: 222222,
		},
		Error: "",
	}, nil
}

func (m *mockGetUserGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}

func (m *mockGetUserGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}

func (m *mockGetUserGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}

func (m *mockGetUserGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}

func (m *mockGetUserGRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	return nil, nil
}

func TestGetUserProfileHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockGetUserGRPCClient{})
	r.GET("/users/:id/profile", GetUserProfileHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/123/profile", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test@example.com")
	assert.Contains(t, w.Body.String(), "Test User")
}

func TestGetUserProfileHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockGetUserGRPCClient{})
	r.GET("/users/:id/profile", GetUserProfileHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/notfound/profile", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}
