package data

import (
	"context"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// AdminOperationLogRepo is the GORM-backed repository for admin operation logs.
type AdminOperationLogRepo struct {
	repo *repositories.AdminOperationLogRepo
	log  *log.Helper
}

func NewAdminOperationLogRepo(data *Data, logger log.Logger) *AdminOperationLogRepo {
	return &AdminOperationLogRepo{
		repo: repositories.NewAdminOperationLogRepo(data.db, logger),
		log:  log.NewHelper(log.With(logger, "module", "admin-operation-log/repo/admin-service")),
	}
}

func (r *AdminOperationLogRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.repo.Count(ctx, nil)
	return int(count), err
}

func (r *AdminOperationLogRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListAdminOperationLogResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *AdminOperationLogRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *AdminOperationLogRepo) Get(ctx context.Context, req *adminV1.GetAdminOperationLogRequest) (*adminV1.AdminOperationLog, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Get(ctx, req.GetId())
}

func (r *AdminOperationLogRepo) Create(ctx context.Context, req *adminV1.CreateAdminOperationLogRequest) error {
	return r.repo.Create(ctx, req)
}
