package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// AdminLoginLogRepo delegates to the gorm implementation.
type AdminLoginLogRepo = gormcli.AdminLoginLogRepo

func NewAdminLoginLogRepo(data *Data, logger log.Logger) *AdminLoginLogRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "admin-login-log/repo/admin-service")).Fatal("gorm DB is required for AdminLoginLogRepo")
	}
	return gormcli.NewAdminLoginLogRepo(data.GormDB(), logger)
}
