package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli"
)

// NewGormDB creates a gorm DB using the same DSN config as ent.
func NewGormDB(cfg *conf.Bootstrap, logger log.Logger) *gorm.DB {
	l := log.NewHelper(log.With(logger, "module", "gorm/data/admin-service"))
	dsn := cfg.Data.Database.GetSource()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		l.Fatalf("failed opening gorm connection to db: %v", err)
		return nil
	}

	gormcli.InstallGormTelemetry(db, logger)
	return db
}
