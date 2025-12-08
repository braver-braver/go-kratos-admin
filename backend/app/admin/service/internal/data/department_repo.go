package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// DepartmentRepo delegates to the gorm implementation.
type DepartmentRepo = gormcli.DepartmentRepo

func NewDepartmentRepo(data *Data, logger log.Logger) *DepartmentRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "department/repo/admin-service")).Fatal("gorm DB is required for DepartmentRepo")
	}
	return gormcli.NewDepartmentRepo(data.GormDB(), logger)
}
