package models

import "time"

// AdminLoginLog 管理员登录日志模型
type AdminLoginLog struct {
	ID             uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	LoginIP        *string   `gorm:"column:login_ip;size:64" json:"login_ip,omitempty"`
	LoginMAC       *string   `gorm:"column:login_mac;size:32" json:"login_mac,omitempty"`
	UserAgent      *string   `gorm:"column:user_agent;size:512" json:"user_agent,omitempty"`
	BrowserName    *string   `gorm:"column:browser_name;size:64" json:"browser_name,omitempty"`
	BrowserVersion *string   `gorm:"column:browser_version;size:32" json:"browser_version,omitempty"`
	ClientID       *string   `gorm:"column:client_id;size:64" json:"client_id,omitempty"`
	ClientName     *string   `gorm:"column:client_name;size:64" json:"client_name,omitempty"`
	OsName         *string   `gorm:"column:os_name;size:64" json:"os_name,omitempty"`
	OsVersion      *string   `gorm:"column:os_version;size:32" json:"os_version,omitempty"`
	UserID         *uint32   `gorm:"column:user_id;index" json:"user_id,omitempty"`
	Username       *string   `gorm:"column:username;size:255;index" json:"username,omitempty"`
	StatusCode     *int32    `gorm:"column:status_code" json:"status_code,omitempty"`
	Success        *bool     `gorm:"column:success" json:"success,omitempty"`
	Reason         *string   `gorm:"column:reason;size:255" json:"reason,omitempty"`
	Location       *string   `gorm:"column:location;size:255" json:"location,omitempty"`
	LoginTime      time.Time `gorm:"column:login_time;autoCreatedAt" json:"login_time"`
	CreatedAt      time.Time `gorm:"autoCreatedAt" json:"created_at"`
}

// TableName 指定表名
func (AdminLoginLog) TableName() string {
	return "admin_login_logs"
}
