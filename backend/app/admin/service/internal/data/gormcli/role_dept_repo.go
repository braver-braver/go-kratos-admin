package gormcli

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type RoleDeptRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewRoleDeptRepo(db *gorm.DB, logger log.Logger) *RoleDeptRepo {
	return &RoleDeptRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "role-dept/gormcli")),
	}
}

// AssignDepartments replaces all department bindings for a role.
func (r *RoleDeptRepo) AssignDepartments(ctx context.Context, roleId uint32, deptIds []uint32, operatorId uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleId).Delete(&model.SysRoleDept{}).Error; err != nil {
			r.log.Errorf("delete old role departments failed: %s", err.Error())
			return userV1.ErrorInternalServerError("delete old role departments failed")
		}
		if len(deptIds) == 0 {
			return nil
		}
		now := time.Now()
		rows := make([]*model.SysRoleDept, 0, len(deptIds))
		for _, deptId := range deptIds {
			rows = append(rows, &model.SysRoleDept{
				RoleID:    int64(roleId),
				DeptID:    int64(deptId),
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: int64(operatorId),
				UpdatedBy: int64(operatorId),
			})
		}
		if err := tx.CreateInBatches(rows, 100).Error; err != nil {
			r.log.Errorf("assign departments to role failed: %s", err.Error())
			return userV1.ErrorInternalServerError("assign departments to role failed")
		}
		return nil
	})
}

// GetDepartmentIdsByRoleId returns department ids bound to the role.
func (r *RoleDeptRepo) GetDepartmentIdsByRoleId(ctx context.Context, roleId uint32) ([]uint32, error) {
	var ids []uint32
	if err := r.db.WithContext(ctx).
		Table(model.TableNameSysRoleDept).
		Where("role_id = ?", roleId).
		Pluck("dept_id", &ids).Error; err != nil {
		r.log.Errorf("query department ids by role id failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query department ids by role id failed")
	}
	return ids, nil
}

// RemoveDepartments removes specific department bindings for a role.
func (r *RoleDeptRepo) RemoveDepartments(ctx context.Context, roleId uint32, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).
		Where("role_id = ? AND dept_id IN ?", roleId, ids).
		Delete(&model.SysRoleDept{}).Error; err != nil {
		r.log.Errorf("remove departments from role failed: %s", err.Error())
		return userV1.ErrorInternalServerError("remove departments from role failed")
	}
	return nil
}
