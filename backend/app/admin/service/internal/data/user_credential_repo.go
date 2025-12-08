package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// UserCredentialRepo delegates to the gorm implementation.
type UserCredentialRepo = gormcli.UserCredentialRepo

func NewUserCredentialRepo(data *Data, logger log.Logger) *UserCredentialRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "user-credential/repo/admin-service")).Fatal("gorm DB is required for UserCredentialRepo")
	}
	return gormcli.NewUserCredentialRepo(data.GormDB(), logger)
}
