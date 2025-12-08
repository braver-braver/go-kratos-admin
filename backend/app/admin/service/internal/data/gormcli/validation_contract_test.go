package gormcli

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Ensures FK/validation errors surface as failures (parity with ent behavior).
func TestValidationErrorParity_ForeignKeyViolation(t *testing.T) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	// Insert role_menu with non-existent role_id/menu_id to trigger FK violation.
	payload := map[string]any{
		"role_id": 999999,
		"menu_id": 999999,
	}
	err = db.Table("sys_role_menu").Create(payload).Error
	if err == nil {
		t.Fatalf("expected constraint violation, got nil")
	}
	// Accept any constraint-style error (FK or unique) as parity guard.
	t.Logf("constraint error: %v", err)
}
