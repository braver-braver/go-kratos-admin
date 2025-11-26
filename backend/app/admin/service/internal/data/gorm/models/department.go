package models

import "time"

// Department 部门模型
type Department struct {
	ID             uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           *string   `gorm:"column:name;size:255;not null" json:"name,omitempty"`
	Code           *string   `gorm:"column:code;size:64;index" json:"code,omitempty"`
	ParentID       *uint32   `gorm:"column:parent_id;default:0;index" json:"parent_id,omitempty"`
	Path           *string   `gorm:"column:path;size:512" json:"path,omitempty"`
	SortOrder      *int32    `gorm:"column:sort_order;default:0" json:"sort_order,omitempty"` // 排序
	Leader         *uint32   `gorm:"column:leader" json:"leader,omitempty"`
	Phone          *string   `gorm:"column:phone;size:255" json:"phone,omitempty"`
	Email          *string   `gorm:"column:email;size:320" json:"email,omitempty"`
	Description    *string   `gorm:"column:description;size:1023" json:"description,omitempty"`
	Status         *int32    `gorm:"column:status;default:1" json:"status,omitempty"`
	Remark         *string   `gorm:"column:remark;size:512" json:"remark,omitempty"`
	TenantID       *uint32   `gorm:"column:tenant_id;index" json:"tenant_id,omitempty"`
	OrganizationID *uint32   `gorm:"column:organization_id;index" json:"organization_id,omitempty"` // 组织ID
	ManagerID      *uint32   `gorm:"column:manager_id;index" json:"manager_id,omitempty"`           // 部门经理ID
	CreatedAt      time.Time `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy      *uint32   `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy      *uint32   `gorm:"column:update_by" json:"update_by,omitempty"`
}

// TableName 指定表名
func (Department) TableName() string {
	return "departments"
}
