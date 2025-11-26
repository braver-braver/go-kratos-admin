package data

import (
	"context"

	dictV1 "kratos-admin/api/gen/go/dict/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// DictEntryRepo is the GORM-backed repository for dictionary entries.
type DictEntryRepo struct {
	repo *repositories.DictEntryRepo
	log  *log.Helper
}

func NewDictEntryRepo(data *Data, logger log.Logger) *DictEntryRepo {
	return &DictEntryRepo{
		repo: repositories.NewDictEntryRepo(data.db, logger),
		log:  log.NewHelper(log.With(logger, "module", "dict-entry/repo/admin-service")),
	}
}

func (r *DictEntryRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.repo.Count(ctx, nil)
	return int(count), err
}

func (r *DictEntryRepo) List(ctx context.Context, req *pagination.PagingRequest) (*dictV1.ListDictEntryResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *DictEntryRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *DictEntryRepo) Get(ctx context.Context, req *dictV1.GetDictEntryRequest) (*dictV1.DictEntry, error) {
	if req == nil {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Get(ctx, req.GetId())
}

func (r *DictEntryRepo) Create(ctx context.Context, req *dictV1.CreateDictEntryRequest) (*dictV1.DictEntry, error) {
	return r.repo.Create(ctx, req)
}

func (r *DictEntryRepo) Update(ctx context.Context, req *dictV1.UpdateDictEntryRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *DictEntryRepo) Delete(ctx context.Context, req *dictV1.BatchDeleteDictRequest) error {
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

func (r *DictEntryRepo) GetDictEntryByDictTypeCode(ctx context.Context, dictTypeCode string) ([]*dictV1.DictEntry, error) {
	return r.repo.GetDictEntryByDictTypeCode(ctx, dictTypeCode)
}

func (r *DictEntryRepo) BatchDelete(ctx context.Context, ids []uint32) error {
	for _, id := range ids {
		if err := r.repo.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
