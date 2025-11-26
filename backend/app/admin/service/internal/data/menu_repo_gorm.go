package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gorm"
)

// MenuRepoGorm GORM 版本的 Menu Repository
type MenuRepoGorm struct {
	repo *repositories.MenuRepo
	log  *log.Helper
}

func NewMenuRepoGorm(logger log.Logger, db interface{}) *MenuRepoGorm {
	// 这里需要从 Data 结构体中获取 GORM 客户端
	// 暂时使用类型断言，实际使用时需要调整
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("invalid database client type")
	}

	repo := repositories.NewMenuRepo(gormDB, logger)

	return &MenuRepoGorm{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "menu/repo-gorm/admin-service")),
	}
}

func (r *MenuRepoGorm) Count(ctx context.Context, whereCond []interface{}) (int, error) {
	// 将 Ent 的 whereCond 转换为 GORM 的 conditions
	conditions := make(map[string]interface{})
	// 这里需要实现条件转换逻辑
	// 简化实现
	count, err := r.repo.Count(ctx, conditions)
	return int(count), err
}

func (r *MenuRepoGorm) List(ctx context.Context, req *pagination.PagingRequest, treeTravel bool) (*adminV1.ListMenuResponse, error) {
	return r.repo.List(ctx, req, treeTravel)
}

func (r *MenuRepoGorm) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *MenuRepoGorm) Get(ctx context.Context, req *adminV1.GetMenuRequest) (*adminV1.Menu, error) {
	return r.repo.Get(ctx, req)
}

func (r *MenuRepoGorm) Create(ctx context.Context, req *adminV1.CreateMenuRequest) error {
	return r.repo.Create(ctx, req)
}

func (r *MenuRepoGorm) Update(ctx context.Context, req *adminV1.UpdateMenuRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *MenuRepoGorm) Delete(ctx context.Context, req *adminV1.DeleteMenuRequest) error {
	return r.repo.Delete(ctx, req)
}
