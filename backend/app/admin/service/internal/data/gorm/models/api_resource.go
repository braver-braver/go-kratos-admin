package models

import "time"

// ApiResource API资源模型
type ApiResource struct {
	ID         uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	Operation  *string   `gorm:"column:operation;size:255;not null" json:"operation,omitempty"` // 操作名
	Path       *string   `gorm:"column:path;size:512;not null" json:"path,omitempty"`
	Method     *string   `gorm:"column:method;size:10;not null" json:"method,omitempty"`
	Scope      *int32    `gorm:"column:scope;default:0" json:"scope,omitempty"`                          // 作用域 (0: 无效, 1: 管理后台API, 2: 前台应用API)
	Desc       *string   `gorm:"column:description;size:512" json:"description,omitempty"`               // 描述
	Module     *string   `gorm:"column:module;size:128" json:"module,omitempty"`                         // 模块
	ModuleDesc *string   `gorm:"column:module_description;size:255" json:"module_description,omitempty"` // 模块描述
	CreatedAt  time.Time `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy  *uint32   `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy  *uint32   `gorm:"column:update_by" json:"update_by,omitempty"`
}

// TableName 指定表名
func (ApiResource) TableName() string {
	return "sys_api_resources"
}
