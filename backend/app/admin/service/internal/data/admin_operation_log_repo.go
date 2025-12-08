package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// AdminOperationLogRepo delegates to the gorm implementation.
type AdminOperationLogRepo = gormcli.AdminOperationLogRepo

func NewAdminOperationLogRepo(data *Data, logger log.Logger) *AdminOperationLogRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "admin-operation-log/repo/admin-service")).Fatal("gorm DB is required for AdminOperationLogRepo")
	}
	return gormcli.NewAdminOperationLogRepo(data.GormDB(), logger)
}
