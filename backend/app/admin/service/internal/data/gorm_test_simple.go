package data

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/models"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"
)

func TestGormBasicConnection(t *testing.T) {
	// 测试基本的数据库连接
	dsn := "host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=*Abcd123456 dbname=kratos_admin sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	t.Log("Database connection successful")
}

func TestGormUserModel(t *testing.T) {
	// 测试用户模型创建
	dsn := "host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=*Abcd123456 dbname=kratos_admin sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建用户表
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate User model: %v", err)
	}

	t.Log("User model migration successful")
}

func TestGormUserRepoBasic(t *testing.T) {
	// 测试基本的 Repository 功能
	dsn := "host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=*Abcd123456 dbname=kratos_admin sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 迁移
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建 Repository
	logger := log.DefaultLogger
	repo := repositories.NewUserRepo(db, logger)

	// 测试基本查询
	ctx := context.Background()

	// 测试用户是否存在（应该返回 false，因为数据库是空的）
	exists, err := repo.IsExist(ctx, 1)
	assert.NoError(t, err)
	assert.False(t, exists)

	t.Log("Basic Repository test successful")
}

func TestGormUserCreate(t *testing.T) {
	// 测试用户创建
	dsn := "host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=*Abcd123456 dbname=kratos_admin sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 迁移
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建 Repository
	logger := log.DefaultLogger
	repo := repositories.NewUserRepo(db, logger)

	ctx := context.Background()

	// 创建测试用户
	username := "testuser"
	nickname := "Test User"
	email := "test@example.com"

	req := &userV1.CreateUserRequest{
		Data: &userV1.User{
			Username: &username,
			Nickname: &nickname,
			Email:    &email,
		},
	}

	// 创建用户
	user, err := repo.Create(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, *user.Username)
	assert.Equal(t, nickname, *user.Nickname)
	assert.Equal(t, email, *user.Email)
	assert.NotZero(t, user.Id)

	t.Logf("User created successfully with ID: %d", *user.Id)

	// 验证用户存在
	exists, err := repo.IsExist(ctx, *user.Id)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 获取用户
	retrievedUser, err := repo.Get(ctx, *user.Id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedUser)
	assert.Equal(t, username, *retrievedUser.Username)

	t.Log("User creation and retrieval test successful")
}
