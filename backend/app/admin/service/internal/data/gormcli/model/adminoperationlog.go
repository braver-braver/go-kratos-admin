package model

import (
	"time"
)

type AdminOperationLog struct {

	// id
	ID uint32 `json:"id,omitempty" gorm:"column:id;primaryKey"`
	// 创建时间
	CreatedAt *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	// 请求ID
	RequestID *string `json:"request_id,omitempty" gorm:"column:request_id"`
	// 请求方法
	Method *string `json:"method,omitempty" gorm:"column:method"`
	// 操作方法
	Operation *string `json:"operation,omitempty" gorm:"column:operation"`
	// 请求路径
	Path *string `json:"path,omitempty" gorm:"column:path"`
	// 请求源
	Referer *string `json:"referer,omitempty" gorm:"column:referer"`
	// 请求URI
	RequestURI *string `json:"request_uri,omitempty" gorm:"column:request_uri"`
	// 请求体
	RequestBody *string `json:"request_body,omitempty" gorm:"column:request_body"`
	// 请求头
	RequestHeader *string `json:"request_header,omitempty" gorm:"column:request_header"`
	// 响应信息
	Response *string `json:"response,omitempty" gorm:"column:response"`
	// 操作耗时
	CostTime *float64 `json:"cost_time,omitempty" gorm:"column:cost_time"`
	// 操作者用户ID
	UserID *uint32 `json:"user_id,omitempty" gorm:"column:user_id"`
	// 操作者账号名
	Username *string `json:"username,omitempty" gorm:"column:username"`
	// 操作者IP
	ClientIP *string `json:"client_ip,omitempty" gorm:"column:client_ip"`
	// 状态码
	StatusCode *int32 `json:"status_code,omitempty" gorm:"column:status_code"`
	// 操作失败原因
	Reason *string `json:"reason,omitempty" gorm:"column:reason"`
	// 操作成功
	Success *bool `json:"success,omitempty" gorm:"column:success"`
	// 操作地理位置
	Location *string `json:"location,omitempty" gorm:"column:location"`
	// 浏览器的用户代理信息
	UserAgent *string `json:"user_agent,omitempty" gorm:"column:user_agent"`
	// 浏览器名称
	BrowserName *string `json:"browser_name,omitempty" gorm:"column:browser_name"`
	// 浏览器版本
	BrowserVersion *string `json:"browser_version,omitempty" gorm:"column:browser_version"`
	// 客户端ID
	ClientID *string `json:"client_id,omitempty" gorm:"column:client_id"`
	// 客户端名称
	ClientName *string `json:"client_name,omitempty" gorm:"column:client_name"`
	// 操作系统名称
	OsName *string `json:"os_name,omitempty" gorm:"column:os_name"`
	// 操作系统版本
	OsVersion *string `json:"os_version,omitempty" gorm:"column:os_version"`
}

func (AdminOperationLog) TableName() string {
	return "sys_admin_operation_logs"
}
