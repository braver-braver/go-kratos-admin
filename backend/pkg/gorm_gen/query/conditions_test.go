package query

import (
	"testing"

	"kratos-admin/pkg/datautil"

	"gorm.io/gen/field"
)

// Tests cover README-like scenarios on string/int fields.

func TestBuildConditions_StringContainsAndEq(t *testing.T) {
	fs := &datautil.FilterSet{
		And: []map[string]any{
			{"name__contains": "tom", "status": "ON"},
		},
	}
	fieldMap := map[string]field.Expr{
		"name":   field.NewString("t", "name"),
		"status": field.NewString("t", "status"),
	}
	conds := BuildConditions(fs, fieldMap)
	if len(conds) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(conds))
	}
}

func TestBuildConditions_IntRangeAndIn(t *testing.T) {
	fs := &datautil.FilterSet{
		And: []map[string]any{
			{"age__range": []any{10, 20}, "status__in": []any{"ON", "OFF"}, "code__not": "X"},
		},
	}
	fieldMap := map[string]field.Expr{
		"age":    field.NewInt("t", "age"),
		"status": field.NewString("t", "status"),
		"code":   field.NewString("t", "code"),
	}
	conds := BuildConditions(fs, fieldMap)
	// range -> 2 conds, status__in -> 1, code__not ->1 => 4
	if len(conds) != 4 {
		t.Fatalf("expected 4 conditions, got %d", len(conds))
	}
}

func TestBuildConditions_OR(t *testing.T) {
	fs := &datautil.FilterSet{
		Or: []map[string]any{
			{"title__startswith": "A"},
			{"title__endswith": "Z"},
		},
	}
	fieldMap := map[string]field.Expr{
		"title": field.NewString("t", "title"),
	}
	conds := BuildConditions(fs, fieldMap)
	if len(conds) != 1 {
		t.Fatalf("expected 1 OR condition, got %d", len(conds))
	}
}

func TestBuildConditions_UnsupportedFieldSkipped(t *testing.T) {
	fs := &datautil.FilterSet{
		And: []map[string]any{
			{"unknown": "x", "name": "ok"},
		},
	}
	fieldMap := map[string]field.Expr{
		"name": field.NewString("t", "name"),
		// "unknown" intentionally missing
	}
	conds := BuildConditions(fs, fieldMap)
	if len(conds) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(conds))
	}
}

func TestBuildConditions_NumberCoercion(t *testing.T) {
	fs := &datautil.FilterSet{
		And: []map[string]any{
			{"age__gte": "5", "age__lt": 10},
		},
	}
	fieldMap := map[string]field.Expr{
		"age": field.NewInt32("t", "age"),
	}
	conds := BuildConditions(fs, fieldMap)
	if len(conds) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(conds))
	}
}
