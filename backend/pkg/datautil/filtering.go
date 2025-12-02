package datautil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/gorm"
)

// FilterSet holds parsed paging/filter information.
type FilterSet struct {
	And       []map[string]any
	Or        []map[string]any
	OrderBy   []string
	FieldMask []string
	NoPaging  bool
	Page      uint32
	PageSize  uint32
}

// PagingRequest is the expected interface of pagination.PagingRequest.
type PagingRequest interface {
	GetPage() int32
	GetPageSize() int32
	GetQuery() string
	GetOrQuery() string
	GetOrderBy() []string
	GetFieldMask() *fieldmaskpb.FieldMask
	GetNoPaging() bool
	GetTenantId() uint32
}

// ParsePagingRequest parses pagination.PagingRequest fields into FilterSet.
// It tolerates varying generated method signatures by using reflection lookups.
func ParsePagingRequest(req any) (*FilterSet, error) {
	fs := &FilterSet{}

	// Preferred path: strongly typed pagination.PagingRequest
	if pr, ok := req.(PagingRequest); ok {
		fs.NoPaging = pr.GetNoPaging()
		fs.Page = uint32(pr.GetPage())
		fs.PageSize = uint32(pr.GetPageSize())
		andFilters, err := parseFilterJSON(pr.GetQuery())
		if err != nil {
			return nil, err
		}
		orFilters, err := parseFilterJSON(pr.GetOrQuery())
		if err != nil {
			return nil, err
		}
		fs.And = andFilters
		fs.Or = orFilters
		fs.OrderBy = pr.GetOrderBy()
		if fm := pr.GetFieldMask(); fm != nil {
			fs.FieldMask = fm.GetPaths()
		}
		if tid := pr.GetTenantId(); tid > 0 {
			fs.And = append(fs.And, map[string]any{"tenant_id": tid})
		}
	} else {
		// Fallback: reflection for legacy shapes
		fs.NoPaging = callBoolMethod(req, "GetNoPaging")
		fs.Page = callUint32Method(req, "GetPage")
		fs.PageSize = callUint32Method(req, "GetPageSize")

		andFilters, err := parseFilterJSON(callStringMethod(req, "GetQuery"))
		if err != nil {
			return nil, err
		}
		orFilters, err := parseFilterJSON(callStringMethod(req, "GetOrQuery"))
		if err != nil {
			return nil, err
		}
		fs.And = andFilters
		fs.Or = orFilters

		fs.OrderBy = callStringSliceMethod(req, "GetOrderBy")
		fs.FieldMask = getFieldMaskPaths(req)
		if tid := callUint32Method(req, "GetTenantId"); tid > 0 {
			fs.And = append(fs.And, map[string]any{"tenant_id": tid})
		}
	}

	if fs.Page == 0 {
		fs.Page = 1
	}
	if fs.PageSize == 0 {
		fs.PageSize = 10
	}

	return fs, nil
}

// FilterSetFromConditions builds a FilterSet using a single AND map.
// Useful for legacy Count(ctx, conditions) signatures.
func FilterSetFromConditions(conds map[string]any) *FilterSet {
	if conds == nil || len(conds) == 0 {
		return &FilterSet{}
	}
	return &FilterSet{And: []map[string]any{conds}}
}

func callStringMethod(obj any, name string) string {
	m := reflect.ValueOf(obj).MethodByName(name)
	if !m.IsValid() || m.Type().NumIn() != 0 || m.Type().NumOut() != 1 {
		return ""
	}
	if m.Type().Out(0).Kind() != reflect.String {
		return ""
	}
	out := m.Call(nil)
	return out[0].String()
}

func callStringSliceMethod(obj any, name string) []string {
	m := reflect.ValueOf(obj).MethodByName(name)
	if !m.IsValid() || m.Type().NumIn() != 0 || m.Type().NumOut() != 1 {
		return nil
	}
	if m.Type().Out(0).Kind() != reflect.Slice || m.Type().Out(0).Elem().Kind() != reflect.String {
		return nil
	}
	out := m.Call(nil)
	if out[0].IsNil() {
		return nil
	}
	s := out[0]
	result := make([]string, s.Len())
	for i := 0; i < s.Len(); i++ {
		result[i] = s.Index(i).String()
	}
	return result
}

func callBoolMethod(obj any, name string) bool {
	m := reflect.ValueOf(obj).MethodByName(name)
	if !m.IsValid() || m.Type().NumIn() != 0 || m.Type().NumOut() != 1 || m.Type().Out(0).Kind() != reflect.Bool {
		return false
	}
	out := m.Call(nil)
	return out[0].Bool()
}

func callUint32Method(obj any, name string) uint32 {
	m := reflect.ValueOf(obj).MethodByName(name)
	if !m.IsValid() || m.Type().NumIn() != 0 || m.Type().NumOut() != 1 {
		return 0
	}
	switch m.Type().Out(0).Kind() {
	case reflect.Uint32, reflect.Uint, reflect.Uint64:
		out := m.Call(nil)
		return uint32(out[0].Uint())
	case reflect.Int, reflect.Int32, reflect.Int64:
		out := m.Call(nil)
		return uint32(out[0].Int())
	default:
		return 0
	}
}

func getFieldMaskPaths(obj any) []string {
	// Try GetFieldMask() returning a message with GetPaths
	m := reflect.ValueOf(obj).MethodByName("GetFieldMask")
	if m.IsValid() && m.Type().NumIn() == 0 && m.Type().NumOut() == 1 {
		fmVal := m.Call(nil)[0]
		if fmVal.IsValid() && !fmVal.IsNil() {
			if paths := fmVal.MethodByName("GetPaths"); paths.IsValid() && paths.Type().NumIn() == 0 && paths.Type().NumOut() == 1 {
				if paths.Type().Out(0).Kind() == reflect.Slice && paths.Type().Out(0).Elem().Kind() == reflect.String {
					res := paths.Call(nil)[0]
					result := make([]string, res.Len())
					for i := 0; i < res.Len(); i++ {
						result[i] = res.Index(i).String()
					}
					return result
				}
			}
		}
	}

	// Try GetFieldMask() returning []string directly
	if m.IsValid() && m.Type().NumOut() == 1 && m.Type().Out(0).Kind() == reflect.Slice && m.Type().Out(0).Elem().Kind() == reflect.String {
		res := m.Call(nil)[0]
		if res.IsNil() {
			return nil
		}
		result := make([]string, res.Len())
		for i := 0; i < res.Len(); i++ {
			result[i] = res.Index(i).String()
		}
		return result
	}

	// Fallback: check GetFieldMaskPaths()
	m = reflect.ValueOf(obj).MethodByName("GetFieldMaskPaths")
	if m.IsValid() && m.Type().NumIn() == 0 && m.Type().NumOut() == 1 && m.Type().Out(0).Kind() == reflect.Slice && m.Type().Out(0).Elem().Kind() == reflect.String {
		res := m.Call(nil)[0]
		if res.IsNil() {
			return nil
		}
		result := make([]string, res.Len())
		for i := 0; i < res.Len(); i++ {
			result[i] = res.Index(i).String()
		}
		return result
	}

	return nil
}

func parseFilterJSON(raw string) ([]map[string]any, error) {
	if raw == "" {
		return nil, nil
	}

	// single object
	var single map[string]any
	if err := json.Unmarshal([]byte(raw), &single); err == nil {
		return []map[string]any{single}, nil
	}

	// array of objects
	var multi []map[string]any
	if err := json.Unmarshal([]byte(raw), &multi); err == nil {
		return multi, nil
	}

	return nil, fmt.Errorf("invalid filter json: %s", raw)
}

// ApplyFilterSet applies filters/order/fieldMask to a gorm query.
// fieldMap allows translating logical names to DB columns; if absent, the logical name is used.
func ApplyFilterSet(db *gorm.DB, fs *FilterSet, fieldMap map[string]string) *gorm.DB {
	if fs == nil {
		return db
	}
	if len(fs.And) > 0 {
		for _, m := range fs.And {
			db = applyFilterMap(db, m, fieldMap, false)
		}
	}
	if len(fs.Or) > 0 {
		db = db.Where(
			func(tx *gorm.DB) *gorm.DB {
				orTx := tx
				for idx, m := range fs.Or {
					condTx := applyFilterMap(tx.Session(&gorm.Session{}), m, fieldMap, false)
					if idx == 0 {
						orTx = condTx
					} else {
						orTx = orTx.Or(condTx)
					}
				}
				return orTx
			},
		)
	}

	if len(fs.FieldMask) > 0 {
		db = db.Select(fs.FieldMask)
	}

	if len(fs.OrderBy) > 0 {
		for _, ob := range fs.OrderBy {
			desc := strings.HasPrefix(ob, "-")
			field := strings.TrimPrefix(ob, "-")
			col := mapField(fieldMap, field)
			if desc {
				db = db.Order(col + " DESC")
			} else {
				db = db.Order(col + " ASC")
			}
		}
	}

	return db
}

func mapField(fieldMap map[string]string, field string) string {
	if fieldMap != nil {
		if col, ok := fieldMap[field]; ok {
			return col
		}
	}
	return field
}

func applyFilterMap(db *gorm.DB, m map[string]any, fieldMap map[string]string, orMode bool) *gorm.DB {
	for key, v := range m {
		field, op := ParseFilterKey(key)
		col := mapField(fieldMap, field)
		switch op {
		case "", "exact":
			db = db.Where(fmt.Sprintf("%s = ?", col), v)
		case "not":
			db = db.Where(fmt.Sprintf("%s <> ?", col), v)
		case "in":
			db = db.Where(fmt.Sprintf("%s IN ?", col), parseArray(v))
		case "not_in":
			db = db.Where(fmt.Sprintf("%s NOT IN ?", col), parseArray(v))
		case "gte":
			db = db.Where(fmt.Sprintf("%s >= ?", col), v)
		case "gt":
			db = db.Where(fmt.Sprintf("%s > ?", col), v)
		case "lte":
			db = db.Where(fmt.Sprintf("%s <= ?", col), v)
		case "lt":
			db = db.Where(fmt.Sprintf("%s < ?", col), v)
		case "range":
			arr := parseArray(v)
			if len(arr) >= 2 {
				db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", col), arr[0], arr[1])
			}
		case "isnull":
			if toBool(v) {
				db = db.Where(fmt.Sprintf("%s IS NULL", col))
			} else {
				db = db.Where(fmt.Sprintf("%s IS NOT NULL", col))
			}
		case "not_isnull":
			if toBool(v) {
				db = db.Where(fmt.Sprintf("%s IS NOT NULL", col))
			} else {
				db = db.Where(fmt.Sprintf("%s IS NULL", col))
			}
		case "contains":
			db = db.Where(fmt.Sprintf("%s LIKE ?", col), "%"+fmt.Sprint(v)+"%")
		case "icontains":
			db = db.Where(fmt.Sprintf("%s ILIKE ?", col), "%"+fmt.Sprint(v)+"%")
		case "startswith":
			db = db.Where(fmt.Sprintf("%s LIKE ?", col), fmt.Sprint(v)+"%")
		case "istartswith":
			db = db.Where(fmt.Sprintf("%s ILIKE ?", col), fmt.Sprint(v)+"%")
		case "endswith":
			db = db.Where(fmt.Sprintf("%s LIKE ?", col), "%"+fmt.Sprint(v))
		case "iendswith":
			db = db.Where(fmt.Sprintf("%s ILIKE ?", col), "%"+fmt.Sprint(v))
		case "iexact":
			db = db.Where(fmt.Sprintf("%s ILIKE ?", col), fmt.Sprint(v))
		case "regex":
			db = db.Where(fmt.Sprintf("%s ~ ?", col), fmt.Sprint(v))
		case "iregex":
			db = db.Where(fmt.Sprintf("%s ~* ?", col), fmt.Sprint(v))
		default:
			// unsupported operator: fallback to equality
			db = db.Where(fmt.Sprintf("%s = ?", col), v)
		}
	}
	return db
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

func toBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		valLower := strings.ToLower(val)
		return valLower == "true" || valLower == "1" || valLower == "t" || valLower == "yes"
	case int:
		return val != 0
	case int64:
		return val != 0
	case float64:
		return val != 0
	case time.Time:
		return !val.IsZero()
	default:
		return false
	}
}
