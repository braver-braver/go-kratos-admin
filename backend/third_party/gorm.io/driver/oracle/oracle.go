package oracle

import (
	"database/sql"

	"gorm.io/gorm"
)

type Dialector struct {
	DSN string
}

func Open(dsn string) Dialector {
	return Dialector{DSN: dsn}
}

func (d Dialector) DialectName() string {
	return "oracle"
}

func (d Dialector) Open(cfg *gorm.Config) (*sql.DB, error) {
	return sql.Open("oracle", d.DSN)
}
