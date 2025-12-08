package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// MenuRepo delegates to the gorm implementation.
type MenuRepo = gormcli.MenuRepo

func NewMenuRepo(data *Data, logger log.Logger) *MenuRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "menu/repo/admin-service")).Fatal("gorm DB is required for MenuRepo")
	}
	return gormcli.NewMenuRepo(data.GormDB(), logger)
}
