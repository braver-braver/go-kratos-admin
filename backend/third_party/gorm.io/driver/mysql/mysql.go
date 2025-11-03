package mysql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"gorm.io/gorm"
)

type Dialector struct {
	DSN string
}

func Open(dsn string) Dialector {
	return Dialector{DSN: dsn}
}

func (d Dialector) DialectName() string {
	return "mysql"
}

func (d Dialector) Open(cfg *gorm.Config) (*sql.DB, error) {
	return sql.Open("mysql", d.DSN)
}
