package main

import (
	"encoding/json"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DSN sourced from configs/data.yaml; override with DSN env for local runs.
const dsn = "host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=Abcd1234 dbname=postgres sslmode=disable"

var tables = []string{
	"sys_users",
	"sys_roles",
	"sys_menus",
	"sys_tenants",
	"sys_admin_operation_logs",
}

type Snapshot struct {
	Counts  map[string]int64            `json:"counts"`
	Samples map[string][]map[string]any `json:"samples"`
}

func main() {
	database := os.Getenv("DSN")
	if database == "" {
		database = dsn
	}

	db, err := gorm.Open(postgres.Open(database), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	snap := Snapshot{Counts: map[string]int64{}, Samples: map[string][]map[string]any{}}

	for _, table := range tables {
		var cnt int64
		if err := db.Table(table).Count(&cnt).Error; err != nil {
			log.Fatalf("count %s: %v", table, err)
		}
		snap.Counts[table] = cnt

		var rows []map[string]any
		if err := db.Table(table).Limit(5).Find(&rows).Error; err != nil {
			log.Fatalf("select %s: %v", table, err)
		}
		snap.Samples[table] = rows
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}

	outPath := "app/admin/service/internal/data/gormcli/testdata/gorm_results.json"
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		log.Fatalf("write %s: %v", outPath, err)
	}
	log.Printf("written gorm results to %s", outPath)
}
