package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// DSN sourced from backend/app/admin/service/configs/data.yaml
const dsn = "host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=Abcd1234 dbname=postgres sslmode=disable"

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:      "app/admin/service/internal/data/gormcli/generated",
		ModelPkgPath: "app/admin/service/internal/data/gormcli/model",
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	g.UseDB(db)

	all := g.GenerateAllTable()
	g.ApplyBasic(all...)

	g.Execute()
	log.Println("gorm gen executed; check generated/ and model/ for outputs")
}
