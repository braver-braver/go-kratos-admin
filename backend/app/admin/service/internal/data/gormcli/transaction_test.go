package gormcli

import (
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Ensures failed transactions roll back without leaving partial writes.
func TestTransactionRollback_NoPartialWrite(t *testing.T) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	table := "sys_admin_operation_logs"
	var before int64
	if err := db.Table(table).Count(&before).Error; err != nil {
		t.Fatalf("count before: %v", err)
	}

	reqID := time.Now().Format("rollback-%Y%m%d-%H%M%S-%f")
	err = db.Transaction(func(tx *gorm.DB) error {
		payload := map[string]any{
			"request_id": reqID,
			"method":     "POST",
			"path":       "/rollback/probe",
			"success":    true,
		}
		if err := tx.Table(table).Create(payload).Error; err != nil {
			return err
		}
		return assertError // force rollback
	})
	if err == nil {
		t.Fatalf("expected rollback error, got nil")
	}

	var after int64
	if err := db.Table(table).Count(&after).Error; err != nil {
		t.Fatalf("count after: %v", err)
	}
	if after != before {
		t.Fatalf("rollback failed: count changed from %d to %d", before, after)
	}

	var exists int64
	if err := db.Table(table).Where("request_id = ?", reqID).Count(&exists).Error; err != nil {
		t.Fatalf("count inserted row: %v", err)
	}
	if exists != 0 {
		t.Fatalf("rollback failed: found inserted row with request_id %s", reqID)
	}
}

var assertError = &sentinelError{msg: "force rollback"}

type sentinelError struct {
	msg string
}

func (s *sentinelError) Error() string {
	return s.msg
}
