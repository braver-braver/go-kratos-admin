package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// PositionRepo delegates to the gorm implementation.
type PositionRepo = gormcli.PositionRepo

func NewPositionRepo(data *Data, logger log.Logger) *PositionRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "position/repo/admin-service")).Fatal("gorm DB is required for PositionRepo")
	}
	return gormcli.NewPositionRepo(data.GormDB(), logger)
}
