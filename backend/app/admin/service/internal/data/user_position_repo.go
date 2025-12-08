package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// UserPositionRepo delegates to the gorm implementation.
type UserPositionRepo = gormcli.UserPositionRepo

func NewUserPositionRepo(data *Data, logger log.Logger) *UserPositionRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "user-position/repo/admin-service")).Fatal("gorm DB is required for UserPositionRepo")
	}
	return gormcli.NewUserPositionRepo(data.GormDB(), logger)
}
