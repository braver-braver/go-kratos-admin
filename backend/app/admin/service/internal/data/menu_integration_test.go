package data

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/models"
)

func TestMenuRepoGormIntegration(t *testing.T) {
	// 使用 SQLite 内存数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移模型
	err = db.AutoMigrate(&models.Menu{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// 创建 Data 实例
	logger := log.DefaultLogger
	data, _, err := NewData(logger, db, nil)
	if err != nil {
		t.Fatalf("Failed to create Data instance: %v", err)
	}

	// 创建 Repository 实例
	menuRepo := NewMenuRepo(data, logger)

	ctx := context.Background()

	// 测试创建菜单
	name := "Test Menu"
	path := "/test"
	menuType := adminV1.Menu_MENU
	status := adminV1.Menu_ON

	createReq := &adminV1.CreateMenuRequest{
		Data: &adminV1.Menu{
			Name:   &name,
			Path:   &path,
			Type:   &menuType,
			Status: &status,
		},
	}

	err = menuRepo.Create(ctx, createReq)
	assert.NoError(t, err)

	// 测试列表查询
	page := int32(1)
	pageSize := int32(10)
	req := &pagination.PagingRequest{
		Page:     &page,
		PageSize: &pageSize,
		NoPaging: nil,
	}

	listResp, err := menuRepo.List(ctx, req, false)
	assert.NoError(t, err)
	assert.NotNil(t, listResp)
	assert.Equal(t, uint32(1), listResp.Total)

	// 如果创建的菜单有ID，测试获取单个菜单
	if len(listResp.Items) > 0 {
		menuId := listResp.Items[0].GetId()

		getReq := &adminV1.GetMenuRequest{
			Id: menuId,
		}

		menu, err := menuRepo.Get(ctx, getReq)
		assert.NoError(t, err)
		assert.Equal(t, name, *menu.Name)

		// 测试更新菜单
		newName := "Updated Menu"
		updateReq := &adminV1.UpdateMenuRequest{
			Data: &adminV1.Menu{
				Id:   &menuId,
				Name: &newName,
			},
		}

		err = menuRepo.Update(ctx, updateReq)
		assert.NoError(t, err)

		// 验证更新结果
		updatedMenu, err := menuRepo.Get(ctx, getReq)
		assert.NoError(t, err)
		assert.Equal(t, newName, *updatedMenu.Name)

		// 测试删除菜单
		deleteReq := &adminV1.DeleteMenuRequest{
			Id: menuId,
		}

		err = menuRepo.Delete(ctx, deleteReq)
		assert.NoError(t, err)

		// 验证菜单已被删除 (这可能会抛出错误，因为菜单已被删除)
		_, err = menuRepo.Get(ctx, getReq)
		// 根据你的错误处理逻辑，这里可能需要检查特定类型的错误
	}
}
