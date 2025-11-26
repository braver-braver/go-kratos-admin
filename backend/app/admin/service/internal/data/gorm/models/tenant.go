package models

import "time"

// Tenant 租户模型
type Tenant struct {
	ID        uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      *string   `gorm:"column:name;size:255;not null" json:"name,omitempty"`
	Code      *string   `gorm:"column:code;size:64;uniqueIndex;not null" json:"code,omitempty"`
	Status    *int32    `gorm:"column:status;default:1" json:"status,omitempty"`
	Remark    *string   `gorm:"column:remark;size:512" json:"remark,omitempty"`
	CreatedAt time.Time `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy *uint32   `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy *uint32   `gorm:"column:update_by" json:"update_by,omitempty"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "tenants"
}
