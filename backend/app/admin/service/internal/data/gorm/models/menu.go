package models

import "time"

// Menu 菜单模型
type Menu struct {
	ID        uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      *string   `gorm:"column:name;size:255;not null" json:"name,omitempty"`
	Code      *string   `gorm:"column:code;size:64;index;not null" json:"code,omitempty"`
	ParentID  *uint32   `gorm:"column:parent_id;default:0;index" json:"parent_id,omitempty"`
	Path      *string   `gorm:"column:path;size:512" json:"path,omitempty"`
	Redirect  *string   `gorm:"column:redirect;size:1024" json:"redirect,omitempty"`
	Component *string   `gorm:"column:component;size:512" json:"component,omitempty"`
	Icon      *string   `gorm:"column:icon;size:128" json:"icon,omitempty"`
	Link      *string   `gorm:"column:link;size:512" json:"link,omitempty"`
	Level     *int32    `gorm:"column:level;default:0" json:"level,omitempty"` // 菜单层级
	Sort      *int32    `gorm:"column:sort;default:0" json:"sort,omitempty"`
	Visible   *bool     `gorm:"column:visible;default:true" json:"visible,omitempty"`        // 是否可见
	Disabled  *bool     `gorm:"column:disabled;default:false" json:"disabled,omitempty"`     // 是否禁用
	KeepAlive *bool     `gorm:"column:keep_alive;default:false" json:"keep_alive,omitempty"` // 是否缓存
	Type      *string   `gorm:"column:type;size:32" json:"type,omitempty"`                   // 菜单类型（dir, menu, button）
	Status    *int32    `gorm:"column:status;default:1" json:"status,omitempty"`
	Remark    *string   `gorm:"column:remark;size:512" json:"remark,omitempty"`
	TenantID  *uint32   `gorm:"column:tenant_id;index" json:"tenant_id,omitempty"`
	CreatedAt time.Time `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy *uint32   `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy *uint32   `gorm:"column:update_by" json:"update_by,omitempty"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return "menus"
}
