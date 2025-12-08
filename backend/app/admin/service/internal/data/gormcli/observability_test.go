package gormcli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Validates telemetry hooks emit logs with context propagation.
func TestObservabilityHooksEmitLogs(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := log.With(log.NewStdLogger(buf), "trace_id", "obs-test")

	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	InstallGormTelemetry(db, baseLogger)

	var rows []map[string]any
	if err := db.WithContext(context.Background()).Table("sys_users").Limit(1).Find(&rows).Error; err != nil {
		t.Fatalf("query with telemetry: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "gorm query") {
		t.Fatalf("expected telemetry log output, got: %s", out)
	}
	if !strings.Contains(out, "trace_id=obs-test") {
		t.Fatalf("expected trace_id propagation in logs, got: %s", out)
	}
}
