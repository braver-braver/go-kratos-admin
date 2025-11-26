package gorm

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"
)

// NewGormClient 创建 GORM 数据库客户端
func NewGormClient(cfg *conf.Bootstrap, logHelper log.Logger) *gorm.DB {
	l := log.NewHelper(log.With(logHelper, "module", "gorm/data/admin-service"))

	var dialector gorm.Dialector
	switch cfg.Data.Database.GetDriver() {
	case "postgres":
		dialector = postgres.Open(cfg.Data.Database.GetSource())
	case "mysql":
		dialector = mysql.Open(cfg.Data.Database.GetSource())
	default:
		l.Fatalf("unsupported database driver: %s", cfg.Data.Database.GetDriver())
	}

	// 配置 GORM
	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",   // 表前缀
			SingularTable: true, // 使用单数表名
		},
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	db, err := gorm.Open(dialector, config)
	if err != nil {
		l.Fatalf("failed to connect database: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		l.Fatalf("failed to get sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(int(cfg.Data.Database.GetMaxIdleConnections()))
	sqlDB.SetMaxOpenConns(int(cfg.Data.Database.GetMaxOpenConnections()))
	sqlDB.SetConnMaxLifetime(cfg.Data.Database.GetConnectionMaxLifetime().AsDuration())

	// 自动迁移
	if cfg.Data.Database.GetMigrate() {
		if err := autoMigrate(db); err != nil {
			l.Fatalf("failed to auto migrate: %v", err)
		}
	}

	return db
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	// 导入所有模型
	// import "kratos-admin/app/admin/service/internal/data/gorm/models"

	return db.AutoMigrate(
	// &models.User{},
	// &models.Role{},
	// &models.Menu{},
	// ... 其他模型
	)
}
