package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// InternalMessageRepo delegates to the gorm implementation.
type InternalMessageRepo = gormcli.InternalMessageRepo

func NewInternalMessageRepo(data *Data, logger log.Logger) *InternalMessageRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "internal-message/repo/admin-service")).Fatal("gorm DB is required for InternalMessageRepo")
	}
	return gormcli.NewInternalMessageRepo(data.GormDB(), logger)
}
