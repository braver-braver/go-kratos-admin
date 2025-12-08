package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// RoleApiRepo delegates to the gorm implementation.
type RoleApiRepo = gormcli.RoleApiRepo

func NewRoleApiRepo(data *Data, logger log.Logger) *RoleApiRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "role-api/repo/admin-service")).Fatal("gorm DB is required for RoleApiRepo")
	}
	return gormcli.NewRoleApiRepo(data.GormDB(), logger)
}
