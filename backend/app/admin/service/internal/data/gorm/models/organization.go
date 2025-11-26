package models

import "time"

// Organization 组织模型
type Organization struct {
	ID        uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      *string   `gorm:"column:name;size:255;not null" json:"name,omitempty"`
	Code      *string   `gorm:"column:code;size:64;index;not null" json:"code,omitempty"`
	ParentID  *uint32   `gorm:"column:parent_id;default:0;index" json:"parent_id,omitempty"`
	Path      *string   `gorm:"column:path;size:512" json:"path,omitempty"`
	Sort      *int32    `gorm:"column:sort;default:0" json:"sort,omitempty"`
	Status    *int32    `gorm:"column:status;default:1" json:"status,omitempty"`
	Remark    *string   `gorm:"column:remark;size:512" json:"remark,omitempty"`
	TenantID  *uint32   `gorm:"column:tenant_id;index" json:"tenant_id,omitempty"`
	CreatedAt time.Time `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy *uint32   `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy *uint32   `gorm:"column:update_by" json:"update_by,omitempty"`
}

// TableName 指定表名
func (Organization) TableName() string {
	return "organizations"
}
