package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，包含公共字段
type BaseModel struct {
	ID        uint32         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreatedAt" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdatedAt" json:"updated_at"`
	CreatedBy *uint32        `gorm:"column:create_by" json:"create_by,omitempty"`
	UpdatedBy *uint32        `gorm:"column:update_by" json:"update_by,omitempty"`
	Remark    *string        `gorm:"column:remark" json:"remark,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TenantModel 租户模型
type TenantModel struct {
	BaseModel
	TenantID *uint32 `gorm:"column:tenant_id;index" json:"tenant_id,omitempty"`
}

// StatusModel 状态模型
type StatusModel struct {
	TenantModel
	Status *int32 `gorm:"column:status;default:1" json:"status,omitempty"`
}

// User 用户模型
type User struct {
	StatusModel

	// 基本信息
	Username    *string `gorm:"column:username;uniqueIndex;size:255" json:"username,omitempty"`
	Nickname    *string `gorm:"column:nickname;size:255" json:"nickname,omitempty"`
	Realname    *string `gorm:"column:realname;size:255" json:"realname,omitempty"`
	Email       *string `gorm:"column:email;size:320" json:"email,omitempty"`
	Mobile      *string `gorm:"column:mobile;size:255;default:''" json:"mobile,omitempty"`
	Telephone   *string `gorm:"column:telephone;size:255;default:''" json:"telephone,omitempty"`
	Avatar      *string `gorm:"column:avatar;size:1023" json:"avatar,omitempty"`
	Address     *string `gorm:"column:address;size:2048;default:''" json:"address,omitempty"`
	Region      *string `gorm:"column:region;size:255;default:''" json:"region,omitempty"`
	Description *string `gorm:"column:description;size:1023" json:"description,omitempty"`

	// 枚举字段
	Gender    *string `gorm:"column:gender;size:20" json:"gender,omitempty"`
	Authority *string `gorm:"column:authority;size:20;default:'CUSTOMER_USER'" json:"authority,omitempty"`

	// 登录信息
	LastLoginTime *time.Time `gorm:"column:last_login_time" json:"last_login_time,omitempty"`
	LastLoginIP   *string    `gorm:"column:last_login_ip;size:64;default:''" json:"last_login_ip,omitempty"`

	// 组织信息
	OrgID      *uint32 `gorm:"column:org_id" json:"org_id,omitempty"`
	PositionID *uint32 `gorm:"column:position_id" json:"position_id,omitempty"`
	WorkID     *uint32 `gorm:"column:work_id" json:"work_id,omitempty"`

	// 角色信息（JSON 数组）
	Roles *string `gorm:"column:roles;type:json" json:"roles,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM 钩子：创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 可以在这里添加创建前的逻辑
	return nil
}

// BeforeUpdate GORM 钩子：更新前
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 可以在这里添加更新前的逻辑
	return nil
}
