package gormcli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Verifies live gorm snapshot matches ent baselines (counts + field presence).
func TestDataParitySnapshotMatchesEntGoldens(t *testing.T) {
	snap := loadSnapshot(t)

	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	for table, expected := range snap.Counts {
		var got int64
		if err := db.Table(table).Count(&got).Error; err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if got != expected {
			t.Fatalf("table %s count mismatch: got %d want %d", table, got, expected)
		}
	}

	for table, rows := range snap.Samples {
		if len(rows) == 0 {
			continue
		}
		var got []map[string]any
		if err := db.Table(table).Limit(len(rows)).Find(&got).Error; err != nil {
			t.Fatalf("select %s: %v", table, err)
		}
		wantFields := map[string]struct{}{}
		for k := range rows[0] {
			wantFields[k] = struct{}{}
		}
		if len(got) == 0 {
			t.Fatalf("table %s returned 0 rows, expected fields: %v", table, wantFields)
		}
		for k := range wantFields {
			if _, ok := got[0][k]; !ok {
				t.Fatalf("table %s missing field %s in gorm result", table, k)
			}
		}
	}
}

// Compares stored gorm_results.json against ent_goldens to ensure captured parity file is valid.
func TestStoredGormResultsMatchGoldens(t *testing.T) {
	_, err := os.Stat("app/admin/service/internal/data/gormcli/testdata/gorm_results.json")
	if err != nil {
		t.Skip("gorm_results.json not present; run cmd/results to generate")
	}

	entSnap := loadSnapshot(t)
	gormSnap := snapshotFromFile(t, "app/admin/service/internal/data/gormcli/testdata/gorm_results.json")

	if len(entSnap.Counts) != len(gormSnap.Counts) {
		t.Fatalf("count map size mismatch: ent=%d gorm=%d", len(entSnap.Counts), len(gormSnap.Counts))
	}
	for table, entCnt := range entSnap.Counts {
		got, ok := gormSnap.Counts[table]
		if !ok {
			t.Fatalf("gorm results missing table %s", table)
		}
		if got != entCnt {
			t.Fatalf("table %s count mismatch in stored results: got %d want %d", table, got, entCnt)
		}
	}

	for table, entRows := range entSnap.Samples {
		gormRows := gormSnap.Samples[table]
		if len(entRows) == 0 || len(gormRows) == 0 {
			continue
		}
		wantFields := map[string]struct{}{}
		for k := range entRows[0] {
			wantFields[k] = struct{}{}
		}
		for k := range wantFields {
			if _, ok := gormRows[0][k]; !ok {
				t.Fatalf("table %s missing field %s in stored gorm results", table, k)
			}
		}
	}
}

func snapshotFromFile(t *testing.T, path string) snapshot {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var snap snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return snap
}
