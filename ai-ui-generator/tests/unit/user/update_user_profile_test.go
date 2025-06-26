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

type mockUpdateUserProfileGRPCClient struct{ mockGRPCClient }

func (m *mockUpdateUserProfileGRPCClient) UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.Id == "notfound" {
		return &pb.UpdateUserResponse{User: nil, Error: "User not found"}, nil
	}
	return &pb.UpdateUserResponse{
		User: &pb.User{
			Id:        req.Id,
			Email:    "test@example.com",
			Name:     req.Name,
			AvatarUrl: req.AvatarUrl,
			Roles:    []string{"user"},
			CreatedAt: 111111,
			UpdatedAt: 222222,
		},
		Error: "",
	}, nil
}

func (m *mockUpdateUserProfileGRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	return nil, nil
}

func (m *mockUpdateUserProfileGRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	return nil, nil
}

func TestUpdateUserProfileHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockUpdateUserProfileGRPCClient{})
	r.PUT("/users/:id/profile", UpdateUserProfileHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Updated Name", "avatar_url": "http://avatar"}`
	req, _ := http.NewRequest("PUT", "/users/123/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Updated Name")
}

func TestUpdateUserProfileHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	setMockGRPCClientForTest(&mockUpdateUserProfileGRPCClient{})
	r.PUT("/users/:id/profile", UpdateUserProfileHandler)

	w := httptest.NewRecorder()
	body := `{"name": "Name"}`
	req, _ := http.NewRequest("PUT", "/users/notfound/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}
