package data

import (
	"context"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// AdminLoginLogRepo is the GORM-backed repository for admin login logs.
// It replaces the previous Ent implementation.
type AdminLoginLogRepo struct {
	repo *repositories.AdminLoginLogRepo
	log  *log.Helper
}

func NewAdminLoginLogRepo(data *Data, logger log.Logger) *AdminLoginLogRepo {
	return &AdminLoginLogRepo{
		repo: repositories.NewAdminLoginLogRepo(data.db, logger),
		log:  log.NewHelper(log.With(logger, "module", "admin-login-log/repo/admin-service")),
	}
}

func (r *AdminLoginLogRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.repo.Count(ctx, nil)
	return int(count), err
}

func (r *AdminLoginLogRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListAdminLoginLogResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *AdminLoginLogRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *AdminLoginLogRepo) Get(ctx context.Context, req *adminV1.GetAdminLoginLogRequest) (*adminV1.AdminLoginLog, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Get(ctx, req.GetId())
}

func (r *AdminLoginLogRepo) Create(ctx context.Context, req *adminV1.CreateAdminLoginLogRequest) error {
	return r.repo.Create(ctx, req)
}

func (r *AdminLoginLogRepo) Delete(ctx context.Context, req *adminV1.DeleteAdminLoginLogRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Delete(ctx, req.GetId())
}
