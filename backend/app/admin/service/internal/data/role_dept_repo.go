package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// RoleDeptRepo delegates to the gorm implementation.
type RoleDeptRepo = gormcli.RoleDeptRepo

func NewRoleDeptRepo(data *Data, logger log.Logger) *RoleDeptRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "role-dept/repo/admin-service")).Fatal("gorm DB is required for RoleDeptRepo")
	}
	return gormcli.NewRoleDeptRepo(data.GormDB(), logger)
}
