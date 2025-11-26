package data

import (
	"context"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// TenantRepo is the GORM-backed repository for tenants.
type TenantRepo struct {
	repo *repositories.TenantRepo
	log  *log.Helper
}

func NewTenantRepo(data *Data, logger log.Logger) *TenantRepo {
	return &TenantRepo{
		repo: repositories.NewTenantRepo(data.db, logger),
		log:  log.NewHelper(log.With(logger, "module", "tenant/repo/admin-service")),
	}
}

func (r *TenantRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.repo.Count(ctx, nil)
	return int(count), err
}

func (r *TenantRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListTenantResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *TenantRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *TenantRepo) Get(ctx context.Context, req *userV1.GetTenantRequest) (*userV1.Tenant, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Get(ctx, req.GetId())
}

func (r *TenantRepo) Create(ctx context.Context, req *userV1.CreateTenantRequest) (*userV1.Tenant, error) {
	return r.repo.Create(ctx, req.GetData())
}

func (r *TenantRepo) Update(ctx context.Context, req *userV1.UpdateTenantRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *TenantRepo) Delete(ctx context.Context, req *userV1.DeleteTenantRequest) error {
	if req == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Delete(ctx, req.GetId())
}

func (r *TenantRepo) TenantExists(ctx context.Context, req *userV1.TenantExistsRequest) (*userV1.TenantExistsResponse, error) {
	return r.repo.TenantExists(ctx, req)
}

func (r *TenantRepo) GetTenantByTenantCode(ctx context.Context, code string) (*userV1.Tenant, error) {
	return r.repo.GetTenantByTenantCode(ctx, code)
}

func (r *TenantRepo) GetTenantsByIds(ctx context.Context, ids []uint32) ([]*userV1.Tenant, error) {
	return r.repo.GetTenantsByIds(ctx, ids)
}
