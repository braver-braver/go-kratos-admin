package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// DictEntryRepo delegates to the gorm implementation.
type DictEntryRepo = gormcli.DictEntryRepo

func NewDictEntryRepo(data *Data, logger log.Logger) *DictEntryRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "dict-entry/repo/admin-service")).Fatal("gorm DB is required for DictEntryRepo")
	}
	return gormcli.NewDictEntryRepo(data.GormDB(), logger)
}
