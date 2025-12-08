package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// OrganizationRepo delegates to the gorm implementation.
type OrganizationRepo = gormcli.OrganizationRepo

func NewOrganizationRepo(data *Data, logger log.Logger) *OrganizationRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "organization/repo/admin-service")).Fatal("gorm DB is required for OrganizationRepo")
	}
	return gormcli.NewOrganizationRepo(data.GormDB(), logger)
}
