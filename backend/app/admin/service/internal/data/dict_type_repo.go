package data

import (
	"context"

	dictV1 "kratos-admin/api/gen/go/dict/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// DictTypeRepo is the GORM-backed repository for dictionary types.
type DictTypeRepo struct {
	repo *repositories.DictTypeRepo
	log  *log.Helper
}

func NewDictTypeRepo(data *Data, logger log.Logger) *DictTypeRepo {
	return &DictTypeRepo{
		repo: repositories.NewDictTypeRepo(data.db, logger),
		log:  log.NewHelper(log.With(logger, "module", "dict-type/repo/admin-service")),
	}
}

func (r *DictTypeRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.repo.Count(ctx, nil)
	return int(count), err
}

func (r *DictTypeRepo) List(ctx context.Context, req *pagination.PagingRequest) (*dictV1.ListDictTypeResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *DictTypeRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *DictTypeRepo) Get(ctx context.Context, req *dictV1.GetDictTypeRequest) (*dictV1.DictType, error) {
	if req == nil {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Get(ctx, req.GetId())
}

func (r *DictTypeRepo) Create(ctx context.Context, req *dictV1.CreateDictTypeRequest) (*dictV1.DictType, error) {
	return r.repo.Create(ctx, req)
}

func (r *DictTypeRepo) Update(ctx context.Context, req *dictV1.UpdateDictTypeRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *DictTypeRepo) Delete(ctx context.Context, req *dictV1.BatchDeleteDictRequest) error {
	if req == nil {
		return dictV1.ErrorBadRequest("invalid parameter")
	}
	for _, id := range req.GetIds() {
		if err := r.repo.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (r *DictTypeRepo) BatchDelete(ctx context.Context, ids []uint32) error {
	for _, id := range ids {
		if err := r.repo.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
