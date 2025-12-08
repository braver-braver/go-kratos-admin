package gormcli

import (
	"encoding/json"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Default DSN from configs/data.yaml; can override with DSN env.
const defaultDSN = "host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=Abcd1234 dbname=postgres sslmode=disable"

type snapshot struct {
	Counts  map[string]int64            `json:"counts"`
	Samples map[string][]map[string]any `json:"samples"`
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	return db
}

func loadSnapshot(t *testing.T) snapshot {
	t.Helper()
	paths := []string{
		"app/admin/service/internal/data/gormcli/testdata/ent_goldens.json",
		"testdata/ent_goldens.json",
	}
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			var snap snapshot
			if err := json.Unmarshal(data, &snap); err != nil {
				t.Fatalf("decode goldens from %s: %v", p, err)
			}
			return snap
		}
	}
	t.Fatalf("read goldens from any of %v", paths)
	return snapshot{}
}

func TestContractCountsMatchEntBaseline(t *testing.T) {
	db := openTestDB(t)
	snap := loadSnapshot(t)

	for table, expected := range snap.Counts {
		var got int64
		if err := db.Table(table).Count(&got).Error; err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if got != expected {
			t.Fatalf("table %s count mismatch: got %d want %d", table, got, expected)
		}
	}
}

func TestContractSamplesMatchShape(t *testing.T) {
	db := openTestDB(t)
	snap := loadSnapshot(t)

	for table, rows := range snap.Samples {
		if len(rows) == 0 {
			continue
		}
		var got []map[string]any
		if err := db.Table(table).Limit(len(rows)).Find(&got).Error; err != nil {
			t.Fatalf("select %s: %v", table, err)
		}
		// compare field presence (not values) to ensure shape parity
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
