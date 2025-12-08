package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// RoleMenuRepo delegates to the gorm implementation.
type RoleMenuRepo = gormcli.RoleMenuRepo

func NewRoleMenuRepo(data *Data, logger log.Logger) *RoleMenuRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "role-menu/repo/admin-service")).Fatal("gorm DB is required for RoleMenuRepo")
	}
	return gormcli.NewRoleMenuRepo(data.GormDB(), logger)
}
