package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// AdminLoginRestrictionRepo delegates to the gorm implementation.
type AdminLoginRestrictionRepo = gormcli.AdminLoginRestrictionRepo

func NewAdminLoginRestrictionRepo(data *Data, logger log.Logger) *AdminLoginRestrictionRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "admin-login-restriction/repo/admin-service")).Fatal("gorm DB is required for AdminLoginRestrictionRepo")
	}
	return gormcli.NewAdminLoginRestrictionRepo(data.GormDB(), logger)
}
