package data

import (
	"context"

	fileV1 "kratos-admin/api/gen/go/file/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// FileRepo is the GORM-backed repository for files.
type FileRepo struct {
	repo *repositories.FileRepo
	log  *log.Helper
}

func NewFileRepo(data *Data, logger log.Logger) *FileRepo {
	return &FileRepo{
		repo: repositories.NewFileRepo(data.db, logger),
		log:  log.NewHelper(log.With(logger, "module", "file/repo/admin-service")),
	}
}

func (r *FileRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.repo.Count(ctx, nil)
	return int(count), err
}

func (r *FileRepo) List(ctx context.Context, req *pagination.PagingRequest) (*fileV1.ListFileResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *FileRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *FileRepo) Get(ctx context.Context, req *fileV1.GetFileRequest) (*fileV1.File, error) {
	if req == nil {
		return nil, fileV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Get(ctx, req.GetId())
}

func (r *FileRepo) Create(ctx context.Context, req *fileV1.CreateFileRequest) (*fileV1.File, error) {
	return r.repo.Create(ctx, req)
}

func (r *FileRepo) Update(ctx context.Context, req *fileV1.UpdateFileRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *FileRepo) Delete(ctx context.Context, req *fileV1.DeleteFileRequest) error {
	if req == nil {
		return fileV1.ErrorBadRequest("invalid parameter")
	}
	return r.repo.Delete(ctx, req.GetId())
}
