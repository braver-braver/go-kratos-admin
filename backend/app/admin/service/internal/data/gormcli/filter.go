package gormcli

import (
	"strings"

	"gorm.io/gorm"
)

type FilterOp int

const (
	FilterNot                   FilterOp = iota // 不等于
	FilterIn                                    // 检查值是否在列表中
	FilterNotIn                                 // 不在列表中
	FilterGTE                                   // 大于或等于传递的值
	FilterGT                                    // 大于传递值
	FilterLTE                                   // 小于或等于传递值
	FilterLT                                    // 小于传递值
	FilterRange                                 // 是否介于和给定的两个值之间
	FilterIsNull                                // 是否为空
	FilterNotIsNull                             // 是否不为空
	FilterContains                              // 是否包含指定的子字符串
	FilterInsensitiveContains                   // 不区分大小写，是否包含指定的子字符串
	FilterStartsWith                            // 以值开头
	FilterInsensitiveStartsWith                 // 不区分大小写，以值开头
	FilterEndsWith                              // 以值结尾
	FilterInsensitiveEndsWith                   // 不区分大小写，以值结尾
	FilterExact                                 // 精确匹配
	FilterInsensitiveExact                      // 不区分大小写，精确匹配
	FilterRegex                                 // 正则表达式
	FilterInsensitiveRegex                      // 不区分大小写，正则表达式
	FilterSearch                                // 全文搜索
)

var ops = [...]string{
	"not",
	"in",
	"not_in",
	"gte",
	"gt",
	"lte",
	"lt",
	"range",
	"isnull",
	"not_isnull",
	"contains",
	"icontains",
	"startswith",
	"istartswith",
	"endswith",
	"iendswith",
	"exact",
	"iexact",
	"regex",
	"iregex",
	"search",
}

type DatePart int

const (
	DatePartDate        DatePart = iota // 日期
	DatePartYear                        // 年
	DatePartISOYear                     // ISO 8601 一年中的周数
	DatePartQuarter                     // 季度
	DatePartMonth                       // 月
	DatePartWeek                        // ISO 8601 周编号 一年中的周数
	DatePartWeekDay                     // 星期几
	DatePartISOWeekDay                  // 星期几
	DatePartDay                         // 日
	DatePartTime                        // 小时：分钟：秒
	DatePartHour                        // 小时
	DatePartMinute                      // 分钟
	DatePartSecond                      // 秒
	DatePartMicrosecond                 // 微秒
)

var dateParts = [...]string{
	DatePartDate:        "date",
	DatePartYear:        "year",
	DatePartISOYear:     "iso_year",
	DatePartQuarter:     "quarter",
	DatePartMonth:       "month",
	DatePartWeek:        "week",
	DatePartWeekDay:     "week_day",
	DatePartISOWeekDay:  "iso_week_day",
	DatePartDay:         "day",
	DatePartTime:        "time",
	DatePartHour:        "hour",
	DatePartMinute:      "minute",
	DatePartSecond:      "second",
	DatePartMicrosecond: "microsecond",
}

// splitJsonFieldKey 分割JSON字段键
func splitJsonFieldKey(key string) []string {
	return strings.Split(key, JsonFieldDelimiter)
}

// isJsonFieldKey 是否为JSON字段键
func isJsonFieldKey(key string) bool {
	return strings.Contains(key, JsonFieldDelimiter)
}

// defaultLimitGuard ensures unbounded scans get capped to a reasonable default.
func defaultLimitGuard() func(db *gorm.Statement) {
	return SafeLimit(100)
}
