package gormcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDocRuleCompliance 验证查询解析函数是否符合文档中的查询规则
func TestDocRuleCompliance(t *testing.T) {
	// 测试基本查询格式（单个map）
	query1 := `{"username":"admin"}`
	filter := scopeFilters(query1, false)
	assert.NotNil(t, filter, "基本查询格式应该被正确解析")

	// 测试数组格式查询
	query2 := `[{"username":"admin"},{"method":"POST"}]`
	filter = scopeFilters(query2, false)
	assert.NotNil(t, filter, "数组格式查询应该被正确解析")

	// 测试带操作符的查询 - 文档中定义的各种操作符
	// not 操作符
	query3 := `{"username__not":"admin"}`
	filter = scopeFilters(query3, false)
	assert.NotNil(t, filter, "not 操作符应该被正确解析")

	// in 操作符
	query4 := `{"status__in":"[\"active\",\"inactive\"]"}`
	filter = scopeFilters(query4, false)
	assert.NotNil(t, filter, "in 操作符应该被正确解析")

	// not_in 操作符
	query5 := `{"status__not_in":"[\"deleted\"]"}`
	filter = scopeFilters(query5, false)
	assert.NotNil(t, filter, "not_in 操作符应该被正确解析")

	// gte 操作符
	query6 := `{"create_time__gte":"2023-10-25"}`
	filter = scopeFilters(query6, false)
	assert.NotNil(t, filter, "gte 操作符应该被正确解析")

	// gt 操作符
	query7 := `{"create_time__gt":"2023-10-25"}`
	filter = scopeFilters(query7, false)
	assert.NotNil(t, filter, "gt 操作符应该被正确解析")

	// lte 操作符
	query8 := `{"create_time__lte":"2023-10-25"}`
	filter = scopeFilters(query8, false)
	assert.NotNil(t, filter, "lte 操作符应该被正确解析")

	// lt 操作符
	query9 := `{"create_time__lt":"2023-10-25"}`
	filter = scopeFilters(query9, false)
	assert.NotNil(t, filter, "lt 操作符应该被正确解析")

	// range 操作符
	query10 := `{"create_time__range":"[\"2023-01-01\", \"2023-12-31\"]"}`
	filter = scopeFilters(query10, false)
	assert.NotNil(t, filter, "range 操作符应该被正确解析")

	// isnull 操作符
	query11 := `{"description__isnull":"true"}`
	filter = scopeFilters(query11, false)
	assert.NotNil(t, filter, "isnull 操作符应该被正确解析")

	// not_isnull 操作符
	query12 := `{"description__not_isnull":"false"}`
	filter = scopeFilters(query12, false)
	assert.NotNil(t, filter, "not_isnull 操作符应该被正确解析")

	// contains 操作符
	query13 := `{"name__contains":"test"}`
	filter = scopeFilters(query13, false)
	assert.NotNil(t, filter, "contains 操作符应该被正确解析")

	// icontains 操作符
	query14 := `{"name__icontains":"test"}`
	filter = scopeFilters(query14, false)
	assert.NotNil(t, filter, "icontains 操作符应该被正确解析")

	// startswith 操作符
	query15 := `{"name__startswith":"pre"}`
	filter = scopeFilters(query15, false)
	assert.NotNil(t, filter, "startswith 操作符应该被正确解析")

	// istartswith 操作符
	query16 := `{"name__istartswith":"pre"}`
	filter = scopeFilters(query16, false)
	assert.NotNil(t, filter, "istartswith 操作符应该被正确解析")

	// endswith 操作符
	query17 := `{"name__endswith":"suf"}`
	filter = scopeFilters(query17, false)
	assert.NotNil(t, filter, "endswith 操作符应该被正确解析")

	// iendswith 操作符
	query18 := `{"name__iendswith":"suf"}`
	filter = scopeFilters(query18, false)
	assert.NotNil(t, filter, "iendswith 操作符应该被正确解析")

	// exact 操作符
	query19 := `{"name__exact":"test"}`
	filter = scopeFilters(query19, false)
	assert.NotNil(t, filter, "exact 操作符应该被正确解析")

	// iexact 操作符
	query20 := `{"name__iexact":"test"}`
	filter = scopeFilters(query20, false)
	assert.NotNil(t, filter, "iexact 操作符应该被正确解析")

	// 日期提取操作符测试
	// date 操作符
	query21 := `{"create_time__date":"2023-01-01"}`
	filter = scopeFilters(query21, false)
	assert.NotNil(t, filter, "date 操作符应该被正确解析")

	// year 操作符
	query22 := `{"create_time__year":"2023"}`
	filter = scopeFilters(query22, false)
	assert.NotNil(t, filter, "year 操作符应该被正确解析")

	// month 操作符
	query23 := `{"create_time__month":"10"}`
	filter = scopeFilters(query23, false)
	assert.NotNil(t, filter, "month 操作符应该被正确解析")

	// day 操作符
	query24 := `{"create_time__day":"15"}`
	filter = scopeFilters(query24, false)
	assert.NotNil(t, filter, "day 操作符应该被正确解析")

	// hour 操作符
	query25 := `{"create_time__hour":"12"}`
	filter = scopeFilters(query25, false)
	assert.NotNil(t, filter, "hour 操作符应该被正确解析")

	// minute 操作符
	query26 := `{"create_time__minute":"30"}`
	filter = scopeFilters(query26, false)
	assert.NotNil(t, filter, "minute 操作符应该被正确解析")

	// second 操作符
	query27 := `{"create_time__second":"45"}`
	filter = scopeFilters(query27, false)
	assert.NotNil(t, filter, "second 操作符应该被正确解析")

	// 测试复杂的组合查询
	query28 := `{"username":"admin","status__in":"[\"active\",\"pending\"]","create_time__gte":"2023-01-01","create_time__lte":"2023-12-31"}`
	filter = scopeFilters(query28, false)
	assert.NotNil(t, filter, "复杂组合查询应该被正确解析")

	// 测试OR查询
	orQuery := `{"status__in":"[\"active\",\"inactive\"]"}`
	filter = scopeFilters(orQuery, true)
	assert.NotNil(t, filter, "OR查询应该被正确解析")
}

// TestQueryDelimiterCompliance 测试双下划线分隔符的使用
func TestQueryDelimiterCompliance(t *testing.T) {
	// 验证查询分隔符常量
	assert.Equal(t, "__", QueryDelimiter, "查询分隔符应该是双下划线")
}

// TestFilterConstantsMapping 测试过滤器常量映射的正确性
func TestFilterConstantsMapping(t *testing.T) {
	// 验证所有操作符映射到正确的字符串
	assert.Equal(t, "not", ops[FilterNot])
	assert.Equal(t, "in", ops[FilterIn])
	assert.Equal(t, "not_in", ops[FilterNotIn])
	assert.Equal(t, "gte", ops[FilterGTE])
	assert.Equal(t, "gt", ops[FilterGT])
	assert.Equal(t, "lte", ops[FilterLTE])
	assert.Equal(t, "lt", ops[FilterLT])
	assert.Equal(t, "range", ops[FilterRange])
	assert.Equal(t, "isnull", ops[FilterIsNull])
	assert.Equal(t, "not_isnull", ops[FilterNotIsNull])
	assert.Equal(t, "contains", ops[FilterContains])
	assert.Equal(t, "icontains", ops[FilterInsensitiveContains])
	assert.Equal(t, "startswith", ops[FilterStartsWith])
	assert.Equal(t, "istartswith", ops[FilterInsensitiveStartsWith])
	assert.Equal(t, "endswith", ops[FilterEndsWith])
	assert.Equal(t, "iendswith", ops[FilterInsensitiveEndsWith])
	assert.Equal(t, "exact", ops[FilterExact])
	assert.Equal(t, "iexact", ops[FilterInsensitiveExact])
	assert.Equal(t, "regex", ops[FilterRegex])
	assert.Equal(t, "iregex", ops[FilterInsensitiveRegex])
	assert.Equal(t, "search", ops[FilterSearch])
}

// TestFieldSnakeCaseConversion 测试字段名转换为蛇形命名
func TestFieldSnakeCaseConversion(t *testing.T) {
	// 这个测试验证字段名会被转换为蛇形命名（在 makeFieldFilter 中使用）
	// 虽然无法直接测试内部实现，但我们可以验证 columnWhitelist 中的映射
	assert.NotNil(t, columnWhitelist["id"])
	assert.NotNil(t, columnWhitelist["created_at"])
	assert.NotNil(t, columnWhitelist["request_id"])
	assert.NotNil(t, columnWhitelist["method"])
	assert.NotNil(t, columnWhitelist["operation"])
	assert.NotNil(t, columnWhitelist["path"])
	assert.NotNil(t, columnWhitelist["username"])
	assert.NotNil(t, columnWhitelist["client_ip"])
	assert.NotNil(t, columnWhitelist["status_code"])
	assert.NotNil(t, columnWhitelist["reason"])
}
