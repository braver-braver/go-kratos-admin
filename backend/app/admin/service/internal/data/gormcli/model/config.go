package model

import (
	"database/sql"

	"gorm.io/cli/gorm/field"
	"gorm.io/cli/gorm/genconfig"
)

var _ = genconfig.Config{
	OutPath: "../../internal/data/gormcli",

	// Map Go types to helper kinds
	FieldTypeMap: map[any]any{
		sql.NullTime{}: field.Time{},
	},

	// Narrow what gets generated (patterns or type literals)
	// IncludeInterfaces: []any{"Query*", models.Query(nil)},
	// IncludeStructs:    []any{"User", "Account*", models.User{}},
}
