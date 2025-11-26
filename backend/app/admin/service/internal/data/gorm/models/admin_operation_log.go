package models

import "time"

// AdminOperationLog 管理员操作日志模型
type AdminOperationLog struct {
	ID             uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestId      *string   `gorm:"column:request_id;size:64;index" json:"request_id,omitempty"`
	Method         *string   `gorm:"column:method;size:10" json:"method,omitempty"`
	Operation      *string   `gorm:"column:operation;size:255" json:"operation,omitempty"`
	Path           *string   `gorm:"column:path;size:512" json:"path,omitempty"`
	Referer        *string   `gorm:"column:referer;size:1024" json:"referer,omitempty"`
	RequestUri     *string   `gorm:"column:request_uri;size:1024" json:"request_uri,omitempty"`
	RequestBody    *string   `gorm:"column:request_body;type:text" json:"request_body,omitempty"`
	RequestHeader  *string   `gorm:"column:request_header;type:text" json:"request_header,omitempty"`
	Response       *string   `gorm:"column:response;type:text" json:"response,omitempty"`
	CostTime       *float64  `gorm:"column:cost_time" json:"cost_time,omitempty"`
	UserID         *uint32   `gorm:"column:user_id;index" json:"user_id,omitempty"`
	Username       *string   `gorm:"column:username;size:255;index" json:"username,omitempty"`
	ClientIp       *string   `gorm:"column:client_ip;size:64" json:"client_ip,omitempty"`
	UserAgent      *string   `gorm:"column:user_agent;size:512" json:"user_agent,omitempty"`
	BrowserName    *string   `gorm:"column:browser_name;size:64" json:"browser_name,omitempty"`
	BrowserVersion *string   `gorm:"column:browser_version;size:32" json:"browser_version,omitempty"`
	ClientID       *string   `gorm:"column:client_id;size:64" json:"client_id,omitempty"`
	ClientName     *string   `gorm:"column:client_name;size:64" json:"client_name,omitempty"`
	OsName         *string   `gorm:"column:os_name;size:64" json:"os_name,omitempty"`
	OsVersion      *string   `gorm:"column:os_version;size:32" json:"os_version,omitempty"`
	StatusCode     *int32    `gorm:"column:status_code" json:"status_code,omitempty"`
	Success        *bool     `gorm:"column:success" json:"success,omitempty"`
	Reason         *string   `gorm:"column:reason;size:255" json:"reason,omitempty"`
	Location       *string   `gorm:"column:location;size:255" json:"location,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreatedAt" json:"created_at"`
}

// TableName 指定表名
func (AdminOperationLog) TableName() string {
	return "admin_operation_logs"
}
