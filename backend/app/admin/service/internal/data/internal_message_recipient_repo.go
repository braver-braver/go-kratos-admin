package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// InternalMessageRecipientRepo delegates to the gorm implementation.
type InternalMessageRecipientRepo = gormcli.InternalMessageRecipientRepo

func NewInternalMessageRecipientRepo(data *Data, logger log.Logger) *InternalMessageRecipientRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "internal-message-recipient/repo/admin-service")).Fatal("gorm DB is required for InternalMessageRecipientRepo")
	}
	return gormcli.NewInternalMessageRecipientRepo(data.GormDB(), logger)
}
