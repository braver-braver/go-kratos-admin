package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// TenantRepo delegates to the gorm implementation.
type TenantRepo = gormcli.TenantRepo

func NewTenantRepo(data *Data, logger log.Logger) *TenantRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "tenant/repo/admin-service")).Fatal("gorm DB is required for TenantRepo")
	}
	return gormcli.NewTenantRepo(data.GormDB(), logger)
}
