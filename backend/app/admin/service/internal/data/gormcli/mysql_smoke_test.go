//go:build mysql
// +build mysql

package gormcli

import (
	"os"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Optional MySQL smoke check; runs only when MYSQL_DSN is provided and build tag `mysql` is set.
func TestMySQLSmoke_QueryCount(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN not provided")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect mysql: %v", err)
	}

	var cnt int64
	if err := db.Table("sys_users").Count(&cnt).Error; err != nil {
		t.Fatalf("mysql smoke count: %v", err)
	}
	t.Logf("mysql sys_users count=%d", cnt)
}
