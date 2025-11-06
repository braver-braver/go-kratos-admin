package postgres

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"gorm.io/gorm"
)

type Dialector struct {
	DSN string
}

func Open(dsn string) Dialector {
	return Dialector{DSN: dsn}
}

func (d Dialector) DialectName() string {
	return "postgres"
}

func (d Dialector) Open(cfg *gorm.Config) (*sql.DB, error) {
	return sql.Open("pgx", d.DSN)
}
