package tests

import (
	"pnas/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository(t *testing.T) {
	// 清理数据库
	cleanupDatabase()

	// 创建测试用户组
	group := &models.UserGroup{
		Name:        "测试组",
		Description: "测试用户组",
		Status:      1,
	}
	err := groupRepo.Create(group)
	require.NoError(t, err)

	t.Run("创建用户", func(t *testing.T) {
		user := &models.User{
			Username: "testuser",
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}

		err := userRepo.Create(user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("根据用户名查询用户", func(t *testing.T) {
		user, err := userRepo.GetByUsername("testuser")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, models.UserTypeNormal, user.UserType)
	})

	t.Run("根据ID查询用户", func(t *testing.T) {
		// 先创建一个用户
		user := &models.User{
			Username: "testuser2",
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeSystem,
			GroupID:  group.ID,
		}
		err := userRepo.Create(user)
		require.NoError(t, err)

		// 查询用户
		foundUser, err := userRepo.GetByID(user.ID)
		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.ID, foundUser.ID)
		assert.Equal(t, "testuser2", foundUser.Username)
	})

	t.Run("更新用户", func(t *testing.T) {
		// 先创建一个用户
		user := &models.User{
			Username: "testuser3",
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}
		err := userRepo.Create(user)
		require.NoError(t, err)

		// 更新用户
		user.Status = models.UserStatusLocked
		user.UserType = models.UserTypeSecurity
		err = userRepo.Update(user)
		assert.NoError(t, err)

		// 验证更新
		updatedUser, err := userRepo.GetByID(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, models.UserStatusLocked, updatedUser.Status)
		assert.Equal(t, models.UserTypeSecurity, updatedUser.UserType)
	})

	t.Run("删除用户", func(t *testing.T) {
		// 先创建一个用户
		user := &models.User{
			Username: "testuser4",
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}
		err := userRepo.Create(user)
		require.NoError(t, err)

		// 删除用户
		err = userRepo.Delete(user.ID)
		assert.NoError(t, err)

		// 验证删除（软删除）
		deletedUser, err := userRepo.GetByID(user.ID)
		assert.Error(t, err)
		assert.Nil(t, deletedUser)
	})

	t.Run("用户列表查询", func(t *testing.T) {
		users, total, err := userRepo.List(0, 10)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.GreaterOrEqual(t, total, int64(0))
	})

	t.Run("根据用户类型查询", func(t *testing.T) {
		// 创建不同类型的用户
		systemUser := &models.User{
			Username: "sysuser",
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeSystem,
			GroupID:  group.ID,
		}
		err := userRepo.Create(systemUser)
		require.NoError(t, err)

		// 查询系统管理员
		users, total, err := userRepo.GetByUserType(models.UserTypeSystem, 0, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 1)
		assert.GreaterOrEqual(t, total, int64(1))
		
		// 验证查询结果
		found := false
		for _, user := range users {
			if user.UserType == models.UserTypeSystem {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestUserGroupRepository(t *testing.T) {
	// 清理数据库
	cleanupDatabase()

	t.Run("创建用户组", func(t *testing.T) {
		group := &models.UserGroup{
			Name:        "测试组1",
			Description: "测试用户组1",
			Status:      1,
		}

		err := groupRepo.Create(group)
		assert.NoError(t, err)
		assert.NotZero(t, group.ID)
	})

	t.Run("根据名称查询用户组", func(t *testing.T) {
		group, err := groupRepo.GetByName("测试组1")
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, "测试组1", group.Name)
	})

	t.Run("根据ID查询用户组", func(t *testing.T) {
		// 先创建一个用户组
		group := &models.UserGroup{
			Name:        "测试组2",
			Description: "测试用户组2",
			Status:      1,
		}
		err := groupRepo.Create(group)
		require.NoError(t, err)

		// 查询用户组
		foundGroup, err := groupRepo.GetByID(group.ID)
		assert.NoError(t, err)
		assert.NotNil(t, foundGroup)
		assert.Equal(t, group.ID, foundGroup.ID)
		assert.Equal(t, "测试组2", foundGroup.Name)
	})

	t.Run("更新用户组", func(t *testing.T) {
		// 先创建一个用户组
		group := &models.UserGroup{
			Name:        "测试组3",
			Description: "测试用户组3",
			Status:      1,
		}
		err := groupRepo.Create(group)
		require.NoError(t, err)

		// 更新用户组
		group.Description = "更新后的描述"
		group.Status = 0
		err = groupRepo.Update(group)
		assert.NoError(t, err)

		// 验证更新
		updatedGroup, err := groupRepo.GetByID(group.ID)
		assert.NoError(t, err)
		assert.Equal(t, "更新后的描述", updatedGroup.Description)
		assert.Equal(t, 0, updatedGroup.Status)
	})

	t.Run("删除用户组", func(t *testing.T) {
		// 先创建一个用户组
		group := &models.UserGroup{
			Name:        "测试组4",
			Description: "测试用户组4",
			Status:      1,
		}
		err := groupRepo.Create(group)
		require.NoError(t, err)

		// 删除用户组
		err = groupRepo.Delete(group.ID)
		assert.NoError(t, err)

		// 验证删除（软删除）
		deletedGroup, err := groupRepo.GetByID(group.ID)
		assert.Error(t, err)
		assert.Nil(t, deletedGroup)
	})

	t.Run("用户组列表查询", func(t *testing.T) {
		groups, total, err := groupRepo.List(0, 10)
		assert.NoError(t, err)
		assert.NotNil(t, groups)
		assert.GreaterOrEqual(t, total, int64(0))
	})

	t.Run("查询用户组及其用户", func(t *testing.T) {
		// 创建用户组
		group := &models.UserGroup{
			Name:        "测试组5",
			Description: "测试用户组5",
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

		// 查询用户组及其用户
		groupWithUsers, err := groupRepo.GetWithUsers(group.ID)
		assert.NoError(t, err)
		assert.NotNil(t, groupWithUsers)
		assert.Equal(t, group.ID, groupWithUsers.ID)
		assert.GreaterOrEqual(t, len(groupWithUsers.Users), 1)
	})
}

func TestHealthCheckRepository(t *testing.T) {
	// 清理数据库
	cleanupDatabase()

	t.Run("创建健康检查记录", func(t *testing.T) {
		healthCheck := &models.HealthCheck{
			ClientIP:  "127.0.0.1",
			UserAgent: "test-agent",
			Status:    "success",
		}

		err := healthRepo.Create(healthCheck)
		assert.NoError(t, err)
		assert.NotZero(t, healthCheck.ID)
	})

	t.Run("根据ID查询健康检查记录", func(t *testing.T) {
		// 先创建一个记录
		healthCheck := &models.HealthCheck{
			ClientIP:  "192.168.1.1",
			UserAgent: "test-agent-2",
			Status:    "success",
		}
		err := healthRepo.Create(healthCheck)
		require.NoError(t, err)

		// 查询记录
		foundRecord, err := healthRepo.GetByID(healthCheck.ID)
		assert.NoError(t, err)
		assert.NotNil(t, foundRecord)
		assert.Equal(t, healthCheck.ID, foundRecord.ID)
		assert.Equal(t, "192.168.1.1", foundRecord.ClientIP)
	})

	t.Run("健康检查记录列表查询", func(t *testing.T) {
		records, total, err := healthRepo.List(0, 10)
		assert.NoError(t, err)
		assert.NotNil(t, records)
		assert.GreaterOrEqual(t, total, int64(0))
	})

	t.Run("获取健康检查统计", func(t *testing.T) {
		// 创建多个记录
		for i := 0; i < 5; i++ {
			healthCheck := &models.HealthCheck{
				ClientIP:  "127.0.0.1",
				UserAgent: "test-agent",
				Status:    "success",
			}
			err := healthRepo.Create(healthCheck)
			require.NoError(t, err)
		}

		stats, err := healthRepo.GetStats()
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "total")
		assert.GreaterOrEqual(t, stats["total"], int64(5))
	})
}