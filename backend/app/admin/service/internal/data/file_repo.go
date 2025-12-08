package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// FileRepo delegates to the gorm implementation.
type FileRepo = gormcli.FileRepo

func NewFileRepo(data *Data, logger log.Logger) *FileRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "file/repo/admin-service")).Fatal("gorm DB is required for FileRepo")
	}
	return gormcli.NewFileRepo(data.GormDB(), logger)
}
