package data

import (
	"context"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// MenuRepo is the GORM-backed repository for menus.
type MenuRepo struct {
	repo *repositories.MenuRepo
	log  *log.Helper
}

func NewMenuRepo(data *Data, logger log.Logger) *MenuRepo {
	return &MenuRepo{
		repo: repositories.NewMenuRepo(data.db, logger),
		log:  log.NewHelper(log.With(logger, "module", "menu/repo/admin-service")),
	}
}

func (r *MenuRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.repo.Count(ctx, nil)
	return int(count), err
}

func (r *MenuRepo) List(ctx context.Context, req *pagination.PagingRequest, treeTravel bool) (*adminV1.ListMenuResponse, error) {
	return r.repo.List(ctx, req, treeTravel)
}

func (r *MenuRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *MenuRepo) Get(ctx context.Context, req *adminV1.GetMenuRequest) (*adminV1.Menu, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Get(ctx, req)
}

func (r *MenuRepo) Create(ctx context.Context, req *adminV1.CreateMenuRequest) error {
	return r.repo.Create(ctx, req)
}

func (r *MenuRepo) Update(ctx context.Context, req *adminV1.UpdateMenuRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *MenuRepo) Delete(ctx context.Context, req *adminV1.DeleteMenuRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Delete(ctx, req)
}
