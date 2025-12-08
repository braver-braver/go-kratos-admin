package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// RoleOrgRepo delegates to the gorm implementation.
type RoleOrgRepo = gormcli.RoleOrgRepo

func NewRoleOrgRepo(data *Data, logger log.Logger) *RoleOrgRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "role-org/repo/admin-service")).Fatal("gorm DB is required for RoleOrgRepo")
	}
	return gormcli.NewRoleOrgRepo(data.GormDB(), logger)
}
