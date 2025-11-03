package data

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-kratos/kratos/v2/log"

	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"

	"gorm.io/driver/mysql"
	"gorm.io/driver/oracle"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewGormClient 创建 GORM ORM 数据库客户端
func NewGormClient(cfg *conf.Bootstrap, logger log.Logger) *gorm.DB {
	l := log.NewHelper(log.With(logger, "module", "gorm/data/admin-service"))

	dbCfg := cfg.GetData().GetDatabase()
	if dbCfg == nil {
		l.Fatalf("database configuration is missing")
		return nil
	}

	gormCfg := &gorm.Config{
		Logger: loggerAdapter{helper: l},
	}

	var (
		db  *gorm.DB
		err error
	)

	switch driver := dbCfg.GetDriver(); driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(dbCfg.GetSource()), gormCfg)
	case "postgres", "pgx", "postgresql":
		db, err = gorm.Open(postgres.Open(dbCfg.GetSource()), gormCfg)
	case "oracle":
		// Oracle connections require a database/sql driver registered under the "oracle" name
		// (for example github.com/sijms/go-ora/v2). The driver can be linked in the binary
		// via a blank import from the application layer.
		db, err = gorm.Open(oracle.Open(dbCfg.GetSource()), gormCfg)
	default:
		err = fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		l.Fatalf("failed opening connection to db: %v", err)
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		l.Fatalf("failed to obtain sql DB from gorm: %v", err)
		return nil
	}

	sqlDB.SetMaxIdleConns(int(dbCfg.GetMaxIdleConnections()))
	sqlDB.SetMaxOpenConns(int(dbCfg.GetMaxOpenConnections()))
	sqlDB.SetConnMaxLifetime(dbCfg.GetConnectionMaxLifetime().AsDuration())

	return db
}

type loggerAdapter struct {
	helper *log.Helper
}

func (l loggerAdapter) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l loggerAdapter) Info(_ context.Context, msg string, data ...interface{}) {
	l.helper.Infof(msg, data...)
}

func (l loggerAdapter) Warn(_ context.Context, msg string, data ...interface{}) {
	l.helper.Warnf(msg, data...)
}

func (l loggerAdapter) Error(_ context.Context, msg string, data ...interface{}) {
	l.helper.Errorf(msg, data...)
}

func (l loggerAdapter) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	if err != nil {
		l.helper.Errorf("sql: %s | rows: %d | err: %v", sql, rows, err)
		return
	}
	l.helper.Debugf("sql: %s | rows: %d | cost: %s", sql, rows, time.Since(begin))
}
