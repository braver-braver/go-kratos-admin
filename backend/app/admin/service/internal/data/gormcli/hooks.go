package gormcli

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

const (
	startKey           = "gorm_start_time"
	defaultSlowQueryMs = 200
)

// InstallGormTelemetry attaches query/exec callbacks that log duration and slow queries.
func InstallGormTelemetry(db *gorm.DB, logger log.Logger) {
	helper := log.NewHelper(log.With(logger, "component", "gorm"))
	slowMs := defaultSlowQueryMs

	db.Callback().Query().Before("gorm:query").Register("telemetry:before_query", func(db *gorm.DB) {
		_ = db.InstanceSet(startKey, time.Now())
	})
	db.Callback().Create().Before("gorm:create").Register("telemetry:before_create", func(db *gorm.DB) {
		_ = db.InstanceSet(startKey, time.Now())
	})
	db.Callback().Update().Before("gorm:update").Register("telemetry:before_update", func(db *gorm.DB) {
		_ = db.InstanceSet(startKey, time.Now())
	})
	db.Callback().Delete().Before("gorm:delete").Register("telemetry:before_delete", func(db *gorm.DB) {
		_ = db.InstanceSet(startKey, time.Now())
	})

	after := func(db *gorm.DB) {
		ctx := db.Statement.Context
		start, _ := db.InstanceGet(startKey)
		var dur time.Duration
		if startTime, ok := start.(time.Time); ok {
			dur = time.Since(startTime)
		}

		sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
		helper.WithContext(ctx).Infof("gorm query duration=%s rows=%d sql=%s", dur.String(), db.RowsAffected, sql)

		if dur > time.Duration(slowMs)*time.Millisecond {
			helper.WithContext(ctx).Warnf("slow query duration=%s sql=%s", dur.String(), sql)
		}
	}

	db.Callback().Query().After("gorm:after_query").Register("telemetry:after_query", after)
	db.Callback().Create().After("gorm:after_create").Register("telemetry:after_create", after)
	db.Callback().Update().After("gorm:after_update").Register("telemetry:after_update", after)
	db.Callback().Delete().After("gorm:after_delete").Register("telemetry:after_delete", after)
}
