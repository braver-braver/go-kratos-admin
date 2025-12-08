package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// TaskRepo delegates to the gorm implementation.
type TaskRepo = gormcli.TaskRepo

func NewTaskRepo(data *Data, logger log.Logger) *TaskRepo {
	if data.GormDB() == nil {
		log.NewHelper(log.With(logger, "module", "task/repo/admin-service")).Fatal("gorm DB is required for TaskRepo")
	}
	return gormcli.NewTaskRepo(data.GormDB(), logger)
}
