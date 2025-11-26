package data

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/models"
)

func setupTestDBForIntegration(t *testing.T) *Data {
	// 使用 SQLite 内存数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移模型
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// 创建 Data 实例
	logger := log.DefaultLogger
	data, _, err := NewData(logger, db, nil)
	if err != nil {
		t.Fatalf("Failed to create Data instance: %v", err)
	}

	return data
}

func TestUserRepoGormIntegration(t *testing.T) {
	data := setupTestDBForIntegration(t)

	// 创建 Repository 实例
	userRepo := NewUserRepo(log.DefaultLogger, data)

	ctx := context.Background()

	// 测试创建用户
	username := "testuser"
	nickname := "Test User"
	email := "test@example.com"

	createReq := &userV1.CreateUserRequest{
		Data: &userV1.User{
			Username: &username,
			Nickname: &nickname,
			Email:    &email,
		},
	}

	createdUser, err := userRepo.Create(ctx, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, username, *createdUser.Username)

	// 测试获取用户
	user, err := userRepo.Get(ctx, *createdUser.Id)
	assert.NoError(t, err)
	assert.Equal(t, username, *user.Username)

	// 测试更新用户
	newNickname := "Updated Nickname"
	updateReq := &userV1.UpdateUserRequest{
		Data: &userV1.User{
			Id:       createdUser.Id,
			Username: &username,
			Nickname: &newNickname,
		},
	}

	err = userRepo.Update(ctx, updateReq)
	assert.NoError(t, err)

	// 验证更新结果
	updatedUser, err := userRepo.Get(ctx, *createdUser.Id)
	assert.NoError(t, err)
	assert.Equal(t, newNickname, *updatedUser.Nickname)

	// 测试检查用户是否存在
	exists, err := userRepo.IsExist(ctx, *createdUser.Id)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 测试根据用户名获取用户
	userByUsername, err := userRepo.GetUserByUserName(ctx, username)
	assert.NoError(t, err)
	assert.Equal(t, username, *userByUsername.Username)

	// 测试删除用户
	err = userRepo.Delete(ctx, *createdUser.Id)
	assert.NoError(t, err)

	// 验证用户已被删除
	_, err = userRepo.Get(ctx, *createdUser.Id)
	assert.Error(t, err)
}
