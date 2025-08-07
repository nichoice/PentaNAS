package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pnas/api/v1"
	"pnas/internal/models"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserManagementIntegration(t *testing.T) {
	router := setupRouter()
	cleanupDatabase()

	var createdGroupID uint
	var createdUserID uint

	t.Run("完整的用户管理流程", func(t *testing.T) {
		// 1. 创建用户组
		t.Run("创建用户组", func(t *testing.T) {
			createGroupReq := v1.CreateUserGroupRequest{
				Name:        "集成测试组",
				Description: "集成测试用户组",
			}

			jsonData, _ := json.Marshal(createGroupReq)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/user-groups", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			// 获取创建的用户组ID
			groups, _, err := groupRepo.List(0, 10)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(groups), 1)
			
			for _, group := range groups {
				if group.Name == "集成测试组" {
					createdGroupID = group.ID
					break
				}
			}
			require.NotZero(t, createdGroupID)
		})

		// 2. 创建用户
		t.Run("创建用户", func(t *testing.T) {
			createUserReq := v1.CreateUserRequest{
				Username: "integrationuser",
				Password: "password123",
				UserType: models.UserTypeNormal,
				GroupID:  createdGroupID,
			}

			jsonData, _ := json.Marshal(createUserReq)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			// 获取创建的用户ID
			user, err := userRepo.GetByUsername("integrationuser")
			require.NoError(t, err)
			require.NotNil(t, user)
			createdUserID = user.ID
		})

		// 3. 验证用户组包含用户
		t.Run("验证用户组包含用户", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/user-groups/"+strconv.Itoa(int(createdGroupID))+"/users", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response v1.UserGroupWithUsersResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "success", response.Status)
			assert.Equal(t, "集成测试组", response.Data.Name)
			assert.GreaterOrEqual(t, len(response.Data.Users), 1)

			// 验证用户在组中
			found := false
			for _, user := range response.Data.Users {
				if user.Username == "integrationuser" {
					found = true
					assert.Equal(t, models.UserTypeNormal, user.UserType)
					assert.Equal(t, models.UserStatusActive, user.Status)
					break
				}
			}
			assert.True(t, found, "用户应该在用户组中")
		})

		// 4. 更新用户信息
		t.Run("更新用户信息", func(t *testing.T) {
			updateUserReq := v1.UpdateUserRequest{
				Username: "updatedintegrationuser",
				Status:   models.UserStatusLocked,
				UserType: models.UserTypeSecurity,
				GroupID:  createdGroupID,
			}

			jsonData, _ := json.Marshal(updateUserReq)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/api/v1/users/"+strconv.Itoa(int(createdUserID)), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// 验证更新
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/users/"+strconv.Itoa(int(createdUserID)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response v1.UserResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "updatedintegrationuser", response.Data.Username)
			assert.Equal(t, models.UserStatusLocked, response.Data.Status)
			assert.Equal(t, models.UserTypeSecurity, response.Data.UserType)
		})

		// 5. 更新用户组信息
		t.Run("更新用户组信息", func(t *testing.T) {
			updateGroupReq := v1.UpdateUserGroupRequest{
				Name:        "更新后的集成测试组",
				Description: "更新后的集成测试用户组描述",
				Status:      1,
			}

			jsonData, _ := json.Marshal(updateGroupReq)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/api/v1/user-groups/"+strconv.Itoa(int(createdGroupID)), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// 验证更新
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/user-groups/"+strconv.Itoa(int(createdGroupID)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response v1.UserGroupResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "更新后的集成测试组", response.Data.Name)
			assert.Equal(t, "更新后的集成测试用户组描述", response.Data.Description)
		})

		// 6. 测试分页和筛选
		t.Run("测试用户列表分页和筛选", func(t *testing.T) {
			// 创建更多用户
			for i := 1; i <= 5; i++ {
				user := &models.User{
					Username: "testuser" + strconv.Itoa(i),
					Password: "password123",
					Status:   models.UserStatusActive,
					UserType: models.UserTypeNormal,
					GroupID:  createdGroupID,
				}
				err := userRepo.Create(user)
				require.NoError(t, err)
			}

			// 测试分页
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/users?page=1&size=3", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response v1.UserListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "success", response.Status)
			assert.LessOrEqual(t, len(response.Data.Users), 3)
			assert.GreaterOrEqual(t, response.Data.Total, int64(6))

			// 测试按用户类型筛选
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/users?user_type="+strconv.Itoa(int(models.UserTypeNormal)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "success", response.Status)
			
			// 验证所有返回的用户都是普通用户类型
			for _, user := range response.Data.Users {
				assert.Equal(t, models.UserTypeNormal, user.UserType)
			}
		})

		// 7. 删除用户
		t.Run("删除用户", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/api/v1/users/"+strconv.Itoa(int(createdUserID)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// 验证删除
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/users/"+strconv.Itoa(int(createdUserID)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		// 8. 删除用户组
		t.Run("删除用户组", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/api/v1/user-groups/"+strconv.Itoa(int(createdGroupID)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// 验证删除
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/user-groups/"+strconv.Itoa(int(createdGroupID)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}

func TestThreeRoleManagementIntegration(t *testing.T) {
	router := setupRouter()
	cleanupDatabase()

	t.Run("三员管理集成测试", func(t *testing.T) {
		// 创建三员管理的用户组
		groups := []struct {
			name        string
			description string
			userType    models.UserType
		}{
			{"系统管理员组", "负责系统配置、用户管理、系统维护", models.UserTypeSystem},
			{"安全管理员组", "负责安全策略制定、权限管理、安全审核", models.UserTypeSecurity},
			{"审计管理员组", "负责系统审计、日志分析、合规检查", models.UserTypeAudit},
		}

		var groupIDs []uint

		// 创建用户组
		for _, g := range groups {
			createGroupReq := v1.CreateUserGroupRequest{
				Name:        g.name,
				Description: g.description,
			}

			jsonData, _ := json.Marshal(createGroupReq)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/user-groups", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			// 获取创建的用户组ID
			group, err := groupRepo.GetByName(g.name)
			require.NoError(t, err)
			groupIDs = append(groupIDs, group.ID)
		}

		// 创建三员管理用户
		users := []struct {
			username string
			userType models.UserType
			groupID  uint
		}{
			{"sysadmin_test", models.UserTypeSystem, groupIDs[0]},
			{"secadmin_test", models.UserTypeSecurity, groupIDs[1]},
			{"auditadmin_test", models.UserTypeAudit, groupIDs[2]},
		}

		for _, u := range users {
			createUserReq := v1.CreateUserRequest{
				Username: u.username,
				Password: "password123",
				UserType: u.userType,
				GroupID:  u.groupID,
			}

			jsonData, _ := json.Marshal(createUserReq)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
		}

		// 验证三员管理用户创建成功
		t.Run("验证三员管理用户", func(t *testing.T) {
			// 验证系统管理员
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/users?user_type="+strconv.Itoa(int(models.UserTypeSystem)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response v1.UserListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(response.Data.Users), 1)

			// 验证安全管理员
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/users?user_type="+strconv.Itoa(int(models.UserTypeSecurity)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(response.Data.Users), 1)

			// 验证审计管理员
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/users?user_type="+strconv.Itoa(int(models.UserTypeAudit)), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(response.Data.Users), 1)
		})

		// 验证每个用户组都有对应的用户
		t.Run("验证用户组包含对应用户", func(t *testing.T) {
			for i, groupID := range groupIDs {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/user-groups/"+strconv.Itoa(int(groupID))+"/users", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response v1.UserGroupWithUsersResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, groups[i].name, response.Data.Name)
				assert.GreaterOrEqual(t, len(response.Data.Users), 1)

				// 验证用户类型匹配
				for _, user := range response.Data.Users {
					if user.Username == users[i].username {
						assert.Equal(t, groups[i].userType, user.UserType)
					}
				}
			}
		})
	})
}
