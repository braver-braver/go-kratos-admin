package data

import (
	"context"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// DepartmentRepo is the GORM-backed repository for departments.
type DepartmentRepo struct {
	repo *repositories.DepartmentRepo
	log  *log.Helper
}

func NewDepartmentRepo(data *Data, logger log.Logger) *DepartmentRepo {
	return &DepartmentRepo{
		repo: repositories.NewDepartmentRepo(data.db, logger),
		log:  log.NewHelper(log.With(logger, "module", "department/repo/admin-service")),
	}
}

func (r *DepartmentRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	return r.repo.Count(ctx, conditions)
}

func (r *DepartmentRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListDepartmentResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *DepartmentRepo) Get(ctx context.Context, departmentId uint32) (*userV1.Department, error) {
	return r.repo.Get(ctx, departmentId)
}

func (r *DepartmentRepo) Create(ctx context.Context, req *userV1.CreateDepartmentRequest) (*userV1.Department, error) {
	return r.repo.Create(ctx, req)
}

func (r *DepartmentRepo) Update(ctx context.Context, req *userV1.UpdateDepartmentRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *DepartmentRepo) Delete(ctx context.Context, departmentId uint32) error {
	return r.repo.Delete(ctx, departmentId)
}

func (r *DepartmentRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *DepartmentRepo) GetDepartmentsByIds(ctx context.Context, ids []uint32) ([]*userV1.Department, error) {
	return r.repo.GetDepartmentsByIds(ctx, ids)
}
