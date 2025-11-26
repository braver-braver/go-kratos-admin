package model

import "time"

const TableNameSysUserRole = "sys_user_role"

// SysUserRole models the user-role relation table.
type SysUserRole struct {
	ID        int32      `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement"`
	CreatedAt *time.Time `gorm:"column:created_at;type:timestamp;comment:创建时间"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:timestamp;comment:更新时间"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp;comment:删除时间"`
	CreatedBy *uint32    `gorm:"column:created_by;type:int unsigned;comment:创建者ID"`
	UpdatedBy *uint32    `gorm:"column:updated_by;type:int unsigned;comment:更新者ID"`
	DeletedBy *uint32    `gorm:"column:deleted_by;type:int unsigned;comment:删除者ID"`
	UserID    *uint32    `gorm:"column:user_id;type:int unsigned;index:idx_sys_user_role_user_id,priority:1;comment:用户ID"`
	RoleID    *uint32    `gorm:"column:role_id;type:int unsigned;index:idx_sys_user_role_role_id,priority:1;uniqueIndex:idx_sys_user_role_user_id_role_id,priority:2;comment:角色ID"`
}

// TableName declares the mapped table.
func (*SysUserRole) TableName() string { return TableNameSysUserRole }
