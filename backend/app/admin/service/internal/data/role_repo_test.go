package data

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/model"
	"kratos-admin/app/admin/service/internal/data/gorm/query"
)

func newTestRoleRepo(t *testing.T) (*RoleRepo, *gorm.DB, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.SysRole{})
	require.NoError(t, err)

	query.SetDefault(db)

	data := &Data{
		db:  db,
		log: log.NewHelper(log.With(log.DefaultLogger, "module", "role-repo-test")),
	}

	repo := NewRoleRepo(data, log.DefaultLogger)

	cleanup := func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}

	return repo, db, cleanup
}

func createRoleForTest(t *testing.T, repo *RoleRepo, ctx context.Context, name, code string, sort int32, tenant uint32, parentID *uint32) *userV1.Role {
	t.Helper()

	status := userV1.Role_ON
	remark := fmt.Sprintf("remark-%s", code)

	req := &userV1.CreateRoleRequest{
		Data: &userV1.Role{
			Name:      &name,
			Code:      &code,
			Status:    &status,
			SortOrder: &sort,
			Remark:    &remark,
			TenantId:  proto.Uint32(tenant),
			Menus:     []uint32{1, 2, 3},
			Apis:      []uint32{101, 202},
		},
	}

	if parentID != nil {
		req.Data.ParentId = parentID
	}

	err := repo.Create(ctx, req)
	require.NoError(t, err)

	role, err := repo.GetRoleByCode(ctx, code)
	require.NoError(t, err)
	require.NotNil(t, role)

	return role
}

func TestRoleRepo_CreateAndGet(t *testing.T) {
	repo, _, cleanup := newTestRoleRepo(t)
	defer cleanup()

	ctx := context.Background()

	role := createRoleForTest(t, repo, ctx, "Administrator", "admin", 1, 10, nil)

	require.NotZero(t, role.GetId())
	require.Equal(t, "admin", role.GetCode())
	require.ElementsMatch(t, []uint32{1, 2, 3}, role.GetMenus())
	require.Equal(t, userV1.Role_ON, role.GetStatus())

	count, err := repo.Count(ctx, query.SysRole.Code.Eq("admin"))
	require.NoError(t, err)
	require.Equal(t, 1, count)

	fetched, err := repo.Get(ctx, role.GetId())
	require.NoError(t, err)
	require.Equal(t, role.GetId(), fetched.GetId())
	require.Equal(t, role.GetRemark(), fetched.GetRemark())
}

func TestRoleRepo_UpdateWithFieldMask(t *testing.T) {
	repo, _, cleanup := newTestRoleRepo(t)
	defer cleanup()

	ctx := context.Background()

	role := createRoleForTest(t, repo, ctx, "Auditor", "auditor", 2, 20, nil)

	newRemark := "updated-remark"
	updatedMenus := []uint32{9, 8, 7}

	updateReq := &userV1.UpdateRoleRequest{
		Data: &userV1.Role{
			Id:        proto.Uint32(role.GetId()),
			Remark:    &newRemark,
			Menus:     updatedMenus,
			UpdatedBy: proto.Uint32(99),
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"remark", "menus", "updated_by"}},
	}

	err := repo.Update(ctx, updateReq)
	require.NoError(t, err)

	updated, err := repo.Get(ctx, role.GetId())
	require.NoError(t, err)
	require.Equal(t, newRemark, updated.GetRemark())
	require.ElementsMatch(t, updatedMenus, updated.GetMenus())
	require.Equal(t, uint32(99), updated.GetUpdatedBy())
}

func TestRoleRepo_ListWithQuery(t *testing.T) {
	repo, _, cleanup := newTestRoleRepo(t)
	defer cleanup()

	ctx := context.Background()

	createRoleForTest(t, repo, ctx, "Admin Root", "admin_root", 1, 1, nil)
	createRoleForTest(t, repo, ctx, "Admin Audit", "admin_audit", 2, 1, nil)
	createRoleForTest(t, repo, ctx, "Guest", "guest", 3, 1, nil)

	page := int32(1)
	pageSize := int32(10)
	queryJSON := "{\"name__contains\":\"Admin\"}"

	req := &pagination.PagingRequest{
		Page:     &page,
		PageSize: &pageSize,
		Query:    &queryJSON,
		OrderBy:  []string{"-created_at"},
	}

	resp, err := repo.List(ctx, req)
	require.NoError(t, err)
	require.Equal(t, uint32(2), resp.Total)
	require.Len(t, resp.Items, 2)
	for _, item := range resp.Items {
		require.Contains(t, item.GetName(), "Admin")
	}
}

func TestRoleRepo_DeleteCascade(t *testing.T) {
	repo, db, cleanup := newTestRoleRepo(t)
	defer cleanup()

	ctx := context.Background()

	parent := createRoleForTest(t, repo, ctx, "Parent", "parent", 1, 1, nil)

	parentID := parent.GetId()
	child := createRoleForTest(t, repo, ctx, "Child", "child", 2, 1, &parentID)

	deleteReq := &userV1.DeleteRoleRequest{Id: parent.GetId()}

	err := repo.Delete(ctx, deleteReq)
	require.NoError(t, err)

	exist, err := repo.IsExist(ctx, parent.GetId())
	require.NoError(t, err)
	require.False(t, exist)

	exist, err = repo.IsExist(ctx, child.GetId())
	require.NoError(t, err)
	require.False(t, exist)

	// ensure no residual rows in table
	var count int64
	err = db.WithContext(ctx).Table(model.TableNameSysRole).Count(&count).Error
	require.NoError(t, err)
	require.Zero(t, count)
}
