package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// RoleRepo delegates to the gorm implementation.
type RoleRepo = gormcli.RoleRepo

func NewRoleRepo(data *Data, logger log.Logger) *RoleRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "role/repo/admin-service")).Fatal("gorm DB is required for RoleRepo")
	}
	return gormcli.NewRoleRepo(data.GormDB(), logger)
}
