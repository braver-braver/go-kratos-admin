package models

import "time"

// DictEntry 字典条目模型
type DictEntry struct {
	ID           uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	TypeID       *uint32   `gorm:"column:type_id;index" json:"type_id,omitempty"`                     // 字典类型ID
	EntryLabel   *string   `gorm:"column:entry_label;size:255;not null" json:"entry_label,omitempty"` // 显示标签
	EntryValue   *string   `gorm:"column:entry_value;size:255;not null" json:"entry_value,omitempty"` // 实际值
	NumericValue *int32    `gorm:"column:numeric_value" json:"numeric_value,omitempty"`               // 数值型值
	LanguageCode *string   `gorm:"column:language_code;size:10" json:"language_code,omitempty"`       // 语言代码
	IsEnabled    *bool     `gorm:"column:is_enabled;default:true" json:"is_enabled,omitempty"`        // 状态
	SortOrder    *int32    `gorm:"column:sort_order;default:0" json:"sort_order,omitempty"`           // 排序
	Description  *string   `gorm:"column:description;size:512" json:"description,omitempty"`          // 描述
	CreatedAt    time.Time `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy    *uint32   `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy    *uint32   `gorm:"column:update_by" json:"update_by,omitempty"`
}

// TableName 指定表名
func (DictEntry) TableName() string {
	return "dict_entries"
}
