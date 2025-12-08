package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// UserRepo delegates to the gorm implementation.
type UserRepo = gormcli.UserRepo

func NewUserRepo(logger log.Logger, data *Data) *UserRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "user/repo/admin-service")).Fatal("gorm DB is required for UserRepo")
	}
	return gormcli.NewUserRepo(data.GormDB(), logger)
}
