package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// DictTypeRepo delegates to the gorm implementation.
type DictTypeRepo = gormcli.DictTypeRepo

func NewDictTypeRepo(data *Data, logger log.Logger) *DictTypeRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "dict-type/repo/admin-service")).Fatal("gorm DB is required for DictTypeRepo")
	}
	return gormcli.NewDictTypeRepo(data.GormDB(), logger)
}
