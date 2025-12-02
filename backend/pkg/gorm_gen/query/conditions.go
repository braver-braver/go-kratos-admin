package query

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"kratos-admin/pkg/datautil"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

// BuildConditions builds gorm/gen conditions from a FilterSet using concrete field.* types.
// fieldMap keys are logical names; values are generated fields (field.String, field.Int, etc.).
func BuildConditions(fs *datautil.FilterSet, fieldMap map[string]field.Expr) []gen.Condition {
	if fs == nil {
		return nil
	}
	var conds []gen.Condition

	for _, m := range fs.And {
		conds = append(conds, buildConditionList(m, fieldMap)...)
	}

	if len(fs.Or) > 0 {
		var orExprs []field.Expr
		for _, m := range fs.Or {
			for _, c := range buildConditionList(m, fieldMap) {
				if ce, ok := c.(field.Expr); ok {
					orExprs = append(orExprs, ce)
				}
			}
		}
		if len(orExprs) > 0 {
			conds = append(conds, field.Or(orExprs...))
		}
	}

	return conds
}

func buildConditionList(m map[string]any, fieldMap map[string]field.Expr) []gen.Condition {
	var res []gen.Condition
	for key, v := range m {
		name, op := datautil.ParseFilterKey(key)
		expr, ok := fieldMap[name]
		if !ok {
			continue
		}
		switch f := expr.(type) {
		case field.String:
			res = append(res, stringCond(f, op, v)...)
		case field.Int:
			res = append(res, intCond(f, op, v)...)
		case field.Int32:
			res = append(res, int32Cond(f, op, v)...)
		case field.Int64:
			res = append(res, int64Cond(f, op, v)...)
		case field.Uint:
			res = append(res, uintCond(f, op, v)...)
		case field.Uint32:
			res = append(res, uint32Cond(f, op, v)...)
		case field.Uint64:
			res = append(res, uint64Cond(f, op, v)...)
		default:
			// unsupported type: skip
		}
	}
	return res
}

func stringCond(f field.String, op string, v any) []gen.Condition {
	val := fmt.Sprint(v)
	switch op {
	case "", "exact", "eq":
		return []gen.Condition{f.Eq(val)}
	case "not":
		return []gen.Condition{f.Neq(val)}
	case "in":
		return []gen.Condition{f.In(toStringSlice(parseArray(v))...)}
	case "not_in":
		return []gen.Condition{f.NotIn(toStringSlice(parseArray(v))...)}
	case "contains", "icontains":
		return []gen.Condition{f.Like("%" + val + "%")}
	case "startswith", "istartswith":
		return []gen.Condition{f.Like(val + "%")}
	case "endswith", "iendswith":
		return []gen.Condition{f.Like("%" + val)}
	case "iexact":
		return []gen.Condition{f.Like(val)}
	default:
		return []gen.Condition{f.Eq(val)}
	}
}

func intCond(f field.Int, op string, v any) []gen.Condition {
	val := int(toInt64(v))
	switch op {
	case "", "exact", "eq":
		return []gen.Condition{f.Eq(val)}
	case "not":
		return []gen.Condition{f.Neq(val)}
	case "in":
		return []gen.Condition{f.In(toIntSlice(parseArray(v))...)}
	case "not_in":
		return []gen.Condition{f.NotIn(toIntSlice(parseArray(v))...)}
	case "gte":
		return []gen.Condition{f.Gte(val)}
	case "gt":
		return []gen.Condition{f.Gt(val)}
	case "lte":
		return []gen.Condition{f.Lte(val)}
	case "lt":
		return []gen.Condition{f.Lt(val)}
	case "range":
		arr := toIntSlice(parseArray(v))
		if len(arr) >= 2 {
			return []gen.Condition{f.Gte(arr[0]), f.Lte(arr[1])}
		}
		return nil
	default:
		return []gen.Condition{f.Eq(val)}
	}
}

func int32Cond(f field.Int32, op string, v any) []gen.Condition {
	val := int32(toInt64(v))
	switch op {
	case "", "exact", "eq":
		return []gen.Condition{f.Eq(val)}
	case "not":
		return []gen.Condition{f.Neq(val)}
	case "in":
		return []gen.Condition{f.In(toInt32Slice(parseArray(v))...)}
	case "not_in":
		return []gen.Condition{f.NotIn(toInt32Slice(parseArray(v))...)}
	case "gte":
		return []gen.Condition{f.Gte(val)}
	case "gt":
		return []gen.Condition{f.Gt(val)}
	case "lte":
		return []gen.Condition{f.Lte(val)}
	case "lt":
		return []gen.Condition{f.Lt(val)}
	case "range":
		arr := toInt32Slice(parseArray(v))
		if len(arr) >= 2 {
			return []gen.Condition{f.Gte(arr[0]), f.Lte(arr[1])}
		}
		return nil
	default:
		return []gen.Condition{f.Eq(val)}
	}
}

func int64Cond(f field.Int64, op string, v any) []gen.Condition {
	val := toInt64(v)
	switch op {
	case "", "exact", "eq":
		return []gen.Condition{f.Eq(val)}
	case "not":
		return []gen.Condition{f.Neq(val)}
	case "in":
		return []gen.Condition{f.In(toInt64Slice(parseArray(v))...)}
	case "not_in":
		return []gen.Condition{f.NotIn(toInt64Slice(parseArray(v))...)}
	case "gte":
		return []gen.Condition{f.Gte(val)}
	case "gt":
		return []gen.Condition{f.Gt(val)}
	case "lte":
		return []gen.Condition{f.Lte(val)}
	case "lt":
		return []gen.Condition{f.Lt(val)}
	case "range":
		arr := toInt64Slice(parseArray(v))
		if len(arr) >= 2 {
			return []gen.Condition{f.Gte(arr[0]), f.Lte(arr[1])}
		}
		return nil
	default:
		return []gen.Condition{f.Eq(val)}
	}
}

func uintCond(f field.Uint, op string, v any) []gen.Condition {
	val := uint(toUint64(v))
	switch op {
	case "", "exact", "eq":
		return []gen.Condition{f.Eq(val)}
	case "not":
		return []gen.Condition{f.Neq(val)}
	case "in":
		return []gen.Condition{f.In(toUintSlice(parseArray(v))...)}
	case "not_in":
		return []gen.Condition{f.NotIn(toUintSlice(parseArray(v))...)}
	case "gte":
		return []gen.Condition{f.Gte(val)}
	case "gt":
		return []gen.Condition{f.Gt(val)}
	case "lte":
		return []gen.Condition{f.Lte(val)}
	case "lt":
		return []gen.Condition{f.Lt(val)}
	case "range":
		arr := toUintSlice(parseArray(v))
		if len(arr) >= 2 {
			return []gen.Condition{f.Gte(arr[0]), f.Lte(arr[1])}
		}
		return nil
	default:
		return []gen.Condition{f.Eq(val)}
	}
}

func uint32Cond(f field.Uint32, op string, v any) []gen.Condition {
	val := uint32(toUint64(v))
	switch op {
	case "", "exact", "eq":
		return []gen.Condition{f.Eq(val)}
	case "not":
		return []gen.Condition{f.Neq(val)}
	case "in":
		return []gen.Condition{f.In(toUint32Slice(parseArray(v))...)}
	case "not_in":
		return []gen.Condition{f.NotIn(toUint32Slice(parseArray(v))...)}
	case "gte":
		return []gen.Condition{f.Gte(val)}
	case "gt":
		return []gen.Condition{f.Gt(val)}
	case "lte":
		return []gen.Condition{f.Lte(val)}
	case "lt":
		return []gen.Condition{f.Lt(val)}
	case "range":
		arr := toUint32Slice(parseArray(v))
		if len(arr) >= 2 {
			return []gen.Condition{f.Gte(arr[0]), f.Lte(arr[1])}
		}
		return nil
	default:
		return []gen.Condition{f.Eq(val)}
	}
}

func uint64Cond(f field.Uint64, op string, v any) []gen.Condition {
	val := toUint64(v)
	switch op {
	case "", "exact", "eq":
		return []gen.Condition{f.Eq(val)}
	case "not":
		return []gen.Condition{f.Neq(val)}
	case "in":
		return []gen.Condition{f.In(toUint64Slice(parseArray(v))...)}
	case "not_in":
		return []gen.Condition{f.NotIn(toUint64Slice(parseArray(v))...)}
	case "gte":
		return []gen.Condition{f.Gte(val)}
	case "gt":
		return []gen.Condition{f.Gt(val)}
	case "lte":
		return []gen.Condition{f.Lte(val)}
	case "lt":
		return []gen.Condition{f.Lt(val)}
	case "range":
		arr := toUint64Slice(parseArray(v))
		if len(arr) >= 2 {
			return []gen.Condition{f.Gte(arr[0]), f.Lte(arr[1])}
		}
		return nil
	default:
		return []gen.Condition{f.Eq(val)}
	}
}

func toStringSlice(arr []any) []string {
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		out = append(out, fmt.Sprint(v))
	}
	return out
}

func toIntSlice(arr []any) []int {
	out := make([]int, 0, len(arr))
	for _, v := range arr {
		out = append(out, int(toInt64(v)))
	}
	return out
}

func toInt32Slice(arr []any) []int32 {
	out := make([]int32, 0, len(arr))
	for _, v := range arr {
		out = append(out, int32(toInt64(v)))
	}
	return out
}

func toInt64Slice(arr []any) []int64 {
	out := make([]int64, 0, len(arr))
	for _, v := range arr {
		out = append(out, toInt64(v))
	}
	return out
}

func toUintSlice(arr []any) []uint {
	out := make([]uint, 0, len(arr))
	for _, v := range arr {
		out = append(out, uint(toUint64(v)))
	}
	return out
}

func toUint32Slice(arr []any) []uint32 {
	out := make([]uint32, 0, len(arr))
	for _, v := range arr {
		out = append(out, uint32(toUint64(v)))
	}
	return out
}

func toUint64Slice(arr []any) []uint64 {
	out := make([]uint64, 0, len(arr))
	for _, v := range arr {
		out = append(out, toUint64(v))
	}
	return out
}

func toInt64(v any) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	case uint:
		return int64(val)
	case uint32:
		return int64(val)
	case uint64:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	default:
		return 0
	}
}

func toUint64(v any) uint64 {
	switch val := v.(type) {
	case int:
		return uint64(val)
	case int32:
		return uint64(val)
	case int64:
		return uint64(val)
	case uint:
		return uint64(val)
	case uint32:
		return uint64(val)
	case uint64:
		return val
	case float64:
		return uint64(val)
	case string:
		i, _ := strconv.ParseUint(val, 10, 64)
		return i
	default:
		return 0
	}
}

func parseArray(v any) []any {
	switch vv := v.(type) {
	case []any:
		return vv
	case string:
		if vv == "" {
			return nil
		}
		var arr []any
		if err := json.Unmarshal([]byte(vv), &arr); err == nil {
			return arr
		}
		parts := strings.Split(vv, ",")
		out := make([]any, 0, len(parts))
		for _, p := range parts {
			out = append(out, strings.TrimSpace(p))
		}
		return out
	default:
		return []any{vv}
	}
}
