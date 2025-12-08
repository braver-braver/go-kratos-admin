package gormcli

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SafeLimit applies a maximum rows cap when a query omits pagination.
func SafeLimit(limit int) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if db == nil || limit <= 0 {
			return
		}
		if _, ok := db.Clauses["LIMIT"]; ok {
			return
		}
		db.AddClause(clause.Limit{Limit: &limit})
	}
}
