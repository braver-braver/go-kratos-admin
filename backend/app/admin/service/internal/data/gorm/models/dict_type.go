package models

import "time"

// DictType 字典类型模型
type DictType struct {
	ID          uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	TypeCode    *string   `gorm:"column:type_code;size:64;uniqueIndex;not null" json:"type_code,omitempty"` // 类型代码
	TypeName    *string   `gorm:"column:type_name;size:255;not null" json:"type_name,omitempty"`            // 类型名称
	IsEnabled   *bool     `gorm:"column:is_enabled;default:true" json:"is_enabled,omitempty"`               // 状态
	SortOrder   *int32    `gorm:"column:sort_order;default:0" json:"sort_order,omitempty"`                  // 排序
	Description *string   `gorm:"column:description;size:512" json:"description,omitempty"`                 // 描述
	CreatedAt   time.Time `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy   *uint32   `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy   *uint32   `gorm:"column:update_by" json:"update_by,omitempty"`
}

// TableName 指定表名
func (DictType) TableName() string {
	return "dict_types"
}
