package tests

import (
	"pnas/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserType(t *testing.T) {
	tests := []struct {
		name     string
		userType models.UserType
		expected int
	}{
		{"系统管理员", models.UserTypeSystem, 1},
		{"安全管理员", models.UserTypeSecurity, 2},
		{"审计管理员", models.UserTypeAudit, 3},
		{"普通用户", models.UserTypeNormal, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, int(tt.userType))
		})
	}
}

func TestUserStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   models.UserStatus
		expected int
	}{
		{"禁用", models.UserStatusDisabled, 0},
		{"正常", models.UserStatusActive, 1},
		{"锁定", models.UserStatusLocked, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, int(tt.status))
		})
	}
}

func TestUserModel(t *testing.T) {
	user := &models.User{
		Username: "testuser",
		Password: "password123",
		Status:   models.UserStatusActive,
		UserType: models.UserTypeNormal,
		GroupID:  1,
	}

	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "password123", user.Password)
	assert.Equal(t, models.UserStatusActive, user.Status)
	assert.Equal(t, models.UserTypeNormal, user.UserType)
	assert.Equal(t, uint(1), user.GroupID)
	assert.Equal(t, "users", user.TableName())
}

func TestUserGroupModel(t *testing.T) {
	group := &models.UserGroup{
		Name:        "测试组",
		Description: "测试用户组",
		Status:      1,
	}

	assert.Equal(t, "测试组", group.Name)
	assert.Equal(t, "测试用户组", group.Description)
	assert.Equal(t, 1, group.Status)
	assert.Equal(t, "user_groups", group.TableName())
}

func TestHealthCheckModel(t *testing.T) {
	healthCheck := &models.HealthCheck{
		ClientIP:  "127.0.0.1",
		UserAgent: "test-agent",
		Status:    "success",
	}

	assert.Equal(t, "127.0.0.1", healthCheck.ClientIP)
	assert.Equal(t, "test-agent", healthCheck.UserAgent)
	assert.Equal(t, "success", healthCheck.Status)
	assert.Equal(t, "health_checks", healthCheck.TableName())
}