package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	gormRepo "kratos-admin/app/admin/service/internal/data/gorm/repositories"
)

type IDepartmentRepo interface {
	Count(ctx context.Context, conditions map[string]interface{}) (int64, error)
	List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListDepartmentResponse, error)
	Get(ctx context.Context, departmentId uint32) (*userV1.Department, error)
	Create(ctx context.Context, req *userV1.CreateDepartmentRequest) (*userV1.Department, error)
	Update(ctx context.Context, req *userV1.UpdateDepartmentRequest) error
	Delete(ctx context.Context, departmentId uint32) error
	IsExist(ctx context.Context, id uint32) (bool, error)
	GetDepartmentsByIds(ctx context.Context, ids []uint32) ([]*userV1.Department, error)
}

type DepartmentRepoGorm struct {
	repo *gormRepo.DepartmentRepo
}

func NewDepartmentRepoGorm(db *gorm.DB, logger log.Logger) *DepartmentRepoGorm {
	return &DepartmentRepoGorm{
		repo: gormRepo.NewDepartmentRepo(db, logger),
	}
}

func (r *DepartmentRepoGorm) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	return r.repo.Count(ctx, conditions)
}

func (r *DepartmentRepoGorm) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListDepartmentResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *DepartmentRepoGorm) Get(ctx context.Context, departmentId uint32) (*userV1.Department, error) {
	return r.repo.Get(ctx, departmentId)
}

func (r *DepartmentRepoGorm) Create(ctx context.Context, req *userV1.CreateDepartmentRequest) (*userV1.Department, error) {
	return r.repo.Create(ctx, req)
}

func (r *DepartmentRepoGorm) Update(ctx context.Context, req *userV1.UpdateDepartmentRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *DepartmentRepoGorm) Delete(ctx context.Context, departmentId uint32) error {
	return r.repo.Delete(ctx, departmentId)
}

func (r *DepartmentRepoGorm) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *DepartmentRepoGorm) GetDepartmentsByIds(ctx context.Context, ids []uint32) ([]*userV1.Department, error) {
	return r.repo.GetDepartmentsByIds(ctx, ids)
}
