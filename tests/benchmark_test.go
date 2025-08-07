package tests

import (
	"pnas/internal/models"
	"strconv"
	"testing"
)

func BenchmarkUserRepository(b *testing.B) {
	// 清理数据库
	cleanupDatabase()

	// 创建测试用户组
	group := &models.UserGroup{
		Name:        "性能测试组",
		Description: "性能测试用户组",
		Status:      1,
	}
	err := groupRepo.Create(group)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("CreateUser", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := &models.User{
				Username: "benchuser" + strconv.Itoa(i),
				Password: "password123",
				Status:   models.UserStatusActive,
				UserType: models.UserTypeNormal,
				GroupID:  group.ID,
			}
			err := userRepo.Create(user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 创建一些用户用于查询测试
	for i := 0; i < 100; i++ {
		user := &models.User{
			Username: "queryuser" + strconv.Itoa(i),
			Password: "password123",
			Status:   models.UserStatusActive,
			UserType: models.UserTypeNormal,
			GroupID:  group.ID,
		}
		err := userRepo.Create(user)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Run("GetUserByUsername", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			username := "queryuser" + strconv.Itoa(i%100)
			_, err := userRepo.GetByUsername(username)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ListUsers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := userRepo.List(0, 10)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GetUsersByType", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := userRepo.GetByUserType(models.UserTypeNormal, 0, 10)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkUserGroupRepository(b *testing.B) {
	// 清理数据库
	cleanupDatabase()

	b.Run("CreateUserGroup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			group := &models.UserGroup{
				Name:        "benchgroup" + strconv.Itoa(i),
				Description: "性能测试用户组" + strconv.Itoa(i),
				Status:      1,
			}
			err := groupRepo.Create(group)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 创建一些用户组用于查询测试
	for i := 0; i < 50; i++ {
		group := &models.UserGroup{
			Name:        "querygroup" + strconv.Itoa(i),
			Description: "查询测试用户组" + strconv.Itoa(i),
			Status:      1,
		}
		err := groupRepo.Create(group)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Run("GetUserGroupByName", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			name := "querygroup" + strconv.Itoa(i%50)
			_, err := groupRepo.GetByName(name)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ListUserGroups", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := groupRepo.List(0, 10)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkHealthCheckRepository(b *testing.B) {
	// 清理数据库
	cleanupDatabase()

	b.Run("CreateHealthCheck", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			healthCheck := &models.HealthCheck{
				ClientIP:  "127.0.0.1",
				UserAgent: "benchmark-agent-" + strconv.Itoa(i),
				Status:    "success",
			}
			err := healthRepo.Create(healthCheck)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 创建一些健康检查记录用于查询测试
	for i := 0; i < 100; i++ {
		healthCheck := &models.HealthCheck{
			ClientIP:  "192.168.1." + strconv.Itoa(i%255),
			UserAgent: "query-agent-" + strconv.Itoa(i),
			Status:    "success",
		}
		err := healthRepo.Create(healthCheck)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Run("ListHealthChecks", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := healthRepo.List(0, 10)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GetHealthCheckStats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := healthRepo.GetStats()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}