package data

import (
	"context"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// ApiResourceRepo is the GORM-backed repository for API resources.
type ApiResourceRepo struct {
	repo *repositories.ApiResourceRepo
	log  *log.Helper
}

func NewApiResourceRepo(data *Data, logger log.Logger) *ApiResourceRepo {
	return &ApiResourceRepo{
		repo: repositories.NewApiResourceRepo(data.db, logger),
		log:  log.NewHelper(log.With(logger, "module", "api-resource/repo/admin-service")),
	}
}

func (r *ApiResourceRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.repo.Count(ctx, nil)
	return int(count), err
}

func (r *ApiResourceRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListApiResourceResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *ApiResourceRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *ApiResourceRepo) Get(ctx context.Context, req *adminV1.GetApiResourceRequest) (*adminV1.ApiResource, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Get(ctx, req.GetId())
}

func (r *ApiResourceRepo) Create(ctx context.Context, req *adminV1.CreateApiResourceRequest) (*adminV1.ApiResource, error) {
	return r.repo.Create(ctx, req)
}

func (r *ApiResourceRepo) Update(ctx context.Context, req *adminV1.UpdateApiResourceRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *ApiResourceRepo) Delete(ctx context.Context, req *adminV1.DeleteApiResourceRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Delete(ctx, req.GetId())
}

func (r *ApiResourceRepo) DeleteAll(ctx context.Context) error {
	return adminV1.ErrorBadRequest("delete all not supported in gorm repo")
}

func (r *ApiResourceRepo) GetResourceByPathAndMethod(ctx context.Context, path string, method string) (*adminV1.ApiResource, error) {
	return r.repo.GetResourceByPathAndMethod(ctx, path, method)
}
