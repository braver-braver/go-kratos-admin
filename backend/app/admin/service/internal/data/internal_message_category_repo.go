package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// InternalMessageCategoryRepo delegates to the gorm implementation.
type InternalMessageCategoryRepo = gormcli.InternalMessageCategoryRepo

func NewInternalMessageCategoryRepo(data *Data, logger log.Logger) *InternalMessageCategoryRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "internal-message-category/repo/admin-service")).Fatal("gorm DB is required for InternalMessageCategoryRepo")
	}
	return gormcli.NewInternalMessageCategoryRepo(data.GormDB(), logger)
}
