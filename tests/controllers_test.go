package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pnas/api/v1"
	"pnas/internal/controllers"
	"pnas/internal/models"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// 初始化控制器
	healthController := controllers.NewHealthController(testLogger, healthRepo)
	userController := controllers.NewUserController(testLogger, userRepo, groupRepo)
	groupController := controllers.NewUserGroupController(testLogger, groupRepo)
	
	// 设置路由
	api := r.Group("/api/v1")
	{
		// 健康检查
		api.GET("/health/ping", healthController.Ping)
		
		// 用户管理
		users := api.Group("/users")
		{
			users.POST("/", userController.CreateUser)
			users.GET("/", userController.ListUsers)
			users.GET("/:id", userController.GetUser)
			users.PUT("/:id", userController.UpdateUser)
			users.DELETE("/:id", userController.DeleteUser)
		}
		
		// 用户组管理
		groups := api.Group("/user-groups")
		{
			groups.POST("/", groupController.CreateUserGroup)
			groups.GET("/", groupController.ListUserGroups)
			groups.GET("/:id", groupController.GetUserGroup)
			groups.GET("/:id/users", groupController.GetUserGroupWithUsers)
			groups.PUT("/:id", groupController.UpdateUserGroup)
			groups.DELETE("/:id", groupController.DeleteUserGroup)
		}
	}
	
	return r
}

func TestHealthController(t *testing.T) {
	router := setupRouter()
	cleanupDatabase()

	t.Run("健康检查接口", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/health/ping", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.PingResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "pong", response.Message)
	})
}

func TestUserController(t *testing.T) {
	router := setupRouter()
	cleanupDatabase()

	// 先创建测试用户组
	group := &models.UserGroup{
		Name:        "测试组",
		Description: "测试用户组",
		Status:      1,
	}
	err := groupRepo.Create(group)
	require.NoError(t, err)

	t.Run("创建用户", func(t *testing.T) {
		createReq := v1.CreateUserRequest{
			Username: "testuser",
			Password: "password123",
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}

		jsonData, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response v1.CommonResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "用户创建成功", response.Message)
	})

	t.Run("获取用户列表", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users?page=1&size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.UserListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.GreaterOrEqual(t, len(response.Data.Users), 1)
		assert.GreaterOrEqual(t, response.Data.Total, int64(1))
	})

	t.Run("获取用户详情", func(t *testing.T) {
		// 先创建一个用户
		user := &models.User{
			Username: "detailuser",
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}
		err := userRepo.Create(user)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users/"+strconv.Itoa(int(user.ID)), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "detailuser", response.Data.Username)
		assert.Equal(t, models.UserTypeNormal, response.Data.UserType)
	})

	t.Run("更新用户", func(t *testing.T) {
		// 先创建一个用户
		user := &models.User{
			Username: "updateuser",
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}
		err := userRepo.Create(user)
		require.NoError(t, err)

		updateReq := v1.UpdateUserRequest{
			Username: "updateduser",
			Status:   models.UserStatusLocked,
			UserType: models.UserTypeSecurity,
			GroupID:  group.ID,
		}

		jsonData, _ := json.Marshal(updateReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/users/"+strconv.Itoa(int(user.ID)), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.CommonResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "用户更新成功", response.Message)
	})

	t.Run("删除用户", func(t *testing.T) {
		// 先创建一个用户
		user := &models.User{
			Username: "deleteuser",
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}
		err := userRepo.Create(user)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/users/"+strconv.Itoa(int(user.ID)), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.CommonResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "用户删除成功", response.Message)
	})

	t.Run("创建用户-参数验证失败", func(t *testing.T) {
		createReq := v1.CreateUserRequest{
			Username: "", // 空用户名
			Password: "password123",
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}

		jsonData, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("获取不存在的用户", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/users/99999", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response v1.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Status)
	})
}

func TestUserGroupController(t *testing.T) {
	router := setupRouter()
	cleanupDatabase()

	t.Run("创建用户组", func(t *testing.T) {
		createReq := v1.CreateUserGroupRequest{
			Name:        "新用户组",
			Description: "新用户组描述",
		}

		jsonData, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/user-groups", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response v1.CommonResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "用户组创建成功", response.Message)
	})

	t.Run("获取用户组列表", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/user-groups?page=1&size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.UserGroupListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.GreaterOrEqual(t, len(response.Data.Groups), 1)
	})

	t.Run("获取用户组详情", func(t *testing.T) {
		// 先创建一个用户组
		group := &models.UserGroup{
			Name:        "详情测试组",
			Description: "详情测试组描述",
			Status:      1,
		}
		err := groupRepo.Create(group)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/user-groups/"+strconv.Itoa(int(group.ID)), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.UserGroupResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "详情测试组", response.Data.Name)
	})

	t.Run("获取用户组及其用户", func(t *testing.T) {
		// 先创建一个用户组
		group := &models.UserGroup{
			Name:        "用户测试组",
			Description: "用户测试组描述",
			Status:      1,
		}
		err := groupRepo.Create(group)
		require.NoError(t, err)

		// 创建用户
		user := &models.User{
			Username: "groupuser",
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}
		err = userRepo.Create(user)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/user-groups/"+strconv.Itoa(int(group.ID))+"/users", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.UserGroupWithUsersResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "用户测试组", response.Data.Name)
		assert.GreaterOrEqual(t, len(response.Data.Users), 1)
	})

	t.Run("更新用户组", func(t *testing.T) {
		// 先创建一个用户组
		group := &models.UserGroup{
			Name:        "更新测试组",
			Description: "更新测试组描述",
			Status:      1,
		}
		err := groupRepo.Create(group)
		require.NoError(t, err)

		updateReq := v1.UpdateUserGroupRequest{
			Name:        "更新后的组名",
			Description: "更新后的描述",
			Status:      0,
		}

		jsonData, _ := json.Marshal(updateReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/user-groups/"+strconv.Itoa(int(group.ID)), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.CommonResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "用户组更新成功", response.Message)
	})

	t.Run("删除用户组", func(t *testing.T) {
		// 先创建一个用户组
		group := &models.UserGroup{
			Name:        "删除测试组",
			Description: "删除测试组描述",
			Status:      1,
		}
		err := groupRepo.Create(group)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/user-groups/"+strconv.Itoa(int(group.ID)), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response v1.CommonResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "用户组删除成功", response.Message)
	})

	t.Run("创建用户组-参数验证失败", func(t *testing.T) {
		createReq := v1.CreateUserGroupRequest{
			Name:        "", // 空名称
			Description: "描述",
		}

		jsonData, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/user-groups", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("获取不存在的用户组", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/user-groups/99999", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response v1.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Status)
	})
}