package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// RolePositionRepo delegates to the gorm implementation.
type RolePositionRepo = gormcli.RolePositionRepo

func NewRolePositionRepo(data *Data, logger log.Logger) *RolePositionRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "role-position/repo/admin-service")).Fatal("gorm DB is required for RolePositionRepo")
	}
	return gormcli.NewRolePositionRepo(data.GormDB(), logger)
}
