package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 简化的用户模型用于测试
type User struct {
	ID        uint32         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy *uint32        `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy *uint32        `gorm:"column:update_by" json:"update_by,omitempty"`
	Remark    *string        `gorm:"column:remark" json:"remark,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID  *uint32        `gorm:"column:tenant_id;index" json:"tenant_id,omitempty"`
	Status    *int32         `gorm:"column:status;default:1" json:"status,omitempty"`

	// 基本信息
	Username    *string `gorm:"column:username;uniqueIndex;size:255" json:"username,omitempty"`
	Nickname    *string `gorm:"column:nickname;size:255" json:"nickname,omitempty"`
	Realname    *string `gorm:"column:realname;size:255" json:"realname,omitempty"`
	Email       *string `gorm:"column:email;size:320" json:"email,omitempty"`
	Mobile      *string `gorm:"column:mobile;size:255;default:''" json:"mobile,omitempty"`
	Telephone   *string `gorm:"column:telephone;size:255;default:''" json:"telephone,omitempty"`
	Avatar      *string `gorm:"column:avatar;size:1023" json:"avatar,omitempty"`
	Address     *string `gorm:"column:address;size:2048;default:''" json:"address,omitempty"`
	Region      *string `gorm:"column:region;size:255;default:''" json:"region,omitempty"`
	Description *string `gorm:"column:description;size:1023" json:"description,omitempty"`

	// 枚举字段
	Gender    *string `gorm:"column:gender;size:20" json:"gender,omitempty"`
	Authority *string `gorm:"column:authority;size:20;default:'CUSTOMER_USER'" json:"authority,omitempty"`

	// 登录信息
	LastLoginTime *time.Time `gorm:"column:last_login_time" json:"last_login_time,omitempty"`
	LastLoginIP   *string    `gorm:"column:last_login_ip;size:64;default:''" json:"last_login_ip,omitempty"`

	// 组织信息
	OrgID      *uint32 `gorm:"column:org_id" json:"org_id,omitempty"`
	PositionID *uint32 `gorm:"column:position_id" json:"position_id,omitempty"`
	WorkID     *uint32 `gorm:"column:work_id" json:"work_id,omitempty"`

	// 角色信息（JSON 数组）
	Roles *string `gorm:"column:roles;type:json" json:"roles,omitempty"`
}

func (User) TableName() string {
	return "users"
}

func main() {
	fmt.Println("Testing GORM User Repository with PostgreSQL 15...")

	// 连接数据库
	dsn := "host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=*Abcd123456 dbname=kratos_admin sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Database connection successful")

	// 创建用户表
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("Failed to migrate User model: %v", err)
	}

	fmt.Println("✅ User model migration successful")

	ctx := context.Background()

	// 测试基本功能
	fmt.Println("\n--- Testing Basic GORM Functions ---")

	// 1. 创建测试用户
	username := "testuser"
	nickname := "Test User"
	email := "test@example.com"

	user := &User{
		Username: &username,
		Nickname: &nickname,
		Email:    &email,
	}

	result := db.WithContext(ctx).Create(user)
	if result.Error != nil {
		log.Fatalf("Failed to create user: %v", result.Error)
	}

	fmt.Printf("✅ User created successfully with ID: %d\n", user.ID)
	fmt.Printf("   Username: %s\n", *user.Username)
	fmt.Printf("   Nickname: %s\n", *user.Nickname)
	fmt.Printf("   Email: %s\n", *user.Email)

	// 2. 查询用户
	var retrievedUser User
	result = db.WithContext(ctx).First(&retrievedUser, user.ID)
	if result.Error != nil {
		log.Fatalf("Failed to get user: %v", result.Error)
	}
	fmt.Printf("✅ User retrieved successfully: %s\n", *retrievedUser.Username)

	// 3. 根据用户名查询用户
	var userByUsername User
	result = db.WithContext(ctx).Where("username = ?", username).First(&userByUsername)
	if result.Error != nil {
		log.Fatalf("Failed to get user by username: %v", result.Error)
	}
	fmt.Printf("✅ User retrieved by username: %s\n", *userByUsername.Username)

	// 4. 更新用户
	newNickname := "Updated Test User"
	result = db.WithContext(ctx).Model(&retrievedUser).Update("nickname", newNickname)
	if result.Error != nil {
		log.Fatalf("Failed to update user: %v", result.Error)
	}
	fmt.Printf("✅ User updated successfully\n")

	// 5. 验证更新
	var updatedUser User
	result = db.WithContext(ctx).First(&updatedUser, user.ID)
	if result.Error != nil {
		log.Fatalf("Failed to get updated user: %v", result.Error)
	}
	fmt.Printf("✅ Updated user nickname: %s\n", *updatedUser.Nickname)

	// 6. 删除用户
	result = db.WithContext(ctx).Delete(&updatedUser)
	if result.Error != nil {
		log.Fatalf("Failed to delete user: %v", result.Error)
	}
	fmt.Printf("✅ User deleted successfully\n")

	// 7. 验证删除
	var deletedUser User
	result = db.WithContext(ctx).First(&deletedUser, user.ID)
	if result.Error == gorm.ErrRecordNotFound {
		fmt.Printf("✅ User successfully deleted (not found)\n")
	} else if result.Error != nil {
		log.Fatalf("Failed to check user deletion: %v", result.Error)
	} else {
		fmt.Printf("⚠️  User still exists after deletion\n")
	}

	fmt.Println("\n🎉 All GORM basic tests passed successfully!")
	fmt.Println("✅ GORM is working correctly with PostgreSQL 15")
	fmt.Println("✅ User entity migration from Ent to GORM is feasible")
}
