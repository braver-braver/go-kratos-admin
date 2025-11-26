package models

import "time"

// AdminLoginRestriction 管理员登录限制模型
type AdminLoginRestriction struct {
	ID        uint32     `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetID  *uint32    `gorm:"column:target_id;index" json:"target_id,omitempty"`
	Type      *string    `gorm:"column:type;size:32" json:"type,omitempty"`             // 限制类型 (IP, USER, DEVICE等)
	Method    *string    `gorm:"column:method;size:32" json:"method,omitempty"`         // 限制方法 (BLACKLIST, WHITELIST等)
	Value     *string    `gorm:"column:value;size:255;not null" json:"value,omitempty"` // 限制值
	Reason    *string    `gorm:"column:reason;size:255" json:"reason,omitempty"`        // 限制原因
	ExpiredAt *time.Time `gorm:"column:expired_at" json:"expired_at,omitempty"`         // 过期时间
	CreatedAt time.Time  `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy *uint32    `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy *uint32    `gorm:"column:update_by" json:"update_by,omitempty"`
}

// TableName 指定表名
func (AdminLoginRestriction) TableName() string {
	return "admin_login_restrictions"
}
