package models

import "time"

// File 文件模型
type File struct {
	ID        uint32    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      *string   `gorm:"column:name;size:255;not null" json:"name,omitempty"`
	Path      *string   `gorm:"column:path;size:512;not null" json:"path,omitempty"`
	Size      *int64    `gorm:"column:size" json:"size,omitempty"`
	Type      *string   `gorm:"column:type;size:64" json:"type,omitempty"`
	MimeType  *string   `gorm:"column:mime_type;size:128" json:"mime_type,omitempty"`
	Status    *int32    `gorm:"column:status;default:1" json:"status,omitempty"`
	Remark    *string   `gorm:"column:remark;size:512" json:"remark,omitempty"`
	CreatedAt time.Time `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy *uint32   `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy *uint32   `gorm:"column:update_by" json:"update_by,omitempty"`
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}
