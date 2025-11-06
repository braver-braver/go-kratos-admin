package gorm

import (
	"database/sql"
	"errors"

	"gorm.io/gorm/logger"
)

type Dialector interface {
	DialectName() string
	Open(*Config) (*sql.DB, error)
}

type Config struct {
	Logger logger.Interface
}

type DB struct {
	sqlDB     *sql.DB
	config    *Config
	dialector Dialector
}

func Open(d Dialector, cfg *Config) (*DB, error) {
	if d == nil {
		return nil, errors.New("gorm: nil dialector")
	}
	if cfg == nil {
		cfg = &Config{}
	}
	sqlDB, err := d.Open(cfg)
	if err != nil {
		return nil, err
	}
	return &DB{
		sqlDB:     sqlDB,
		config:    cfg,
		dialector: d,
	}, nil
}

func (db *DB) DB() (*sql.DB, error) {
	if db == nil || db.sqlDB == nil {
		return nil, errors.New("gorm: database is not initialized")
	}
	return db.sqlDB, nil
}

func (db *DB) Dialector() Dialector {
	if db == nil {
		return nil
	}
	return db.dialector
}

func (db *DB) Config() *Config {
	if db == nil {
		return nil
	}
	return db.config
}
