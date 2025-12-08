package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// ApiResourceRepo delegates to the gorm implementation.
type ApiResourceRepo = gormcli.ApiResourceRepo

func NewApiResourceRepo(data *Data, logger log.Logger) *ApiResourceRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "api-resource/repo/admin-service")).Fatal("gorm DB is required for ApiResourceRepo")
	}
	return gormcli.NewApiResourceRepo(data.GormDB(), logger)
}
