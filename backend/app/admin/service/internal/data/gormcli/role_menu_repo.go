package gormcli

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type RoleMenuRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewRoleMenuRepo(db *gorm.DB, logger log.Logger) *RoleMenuRepo {
	return &RoleMenuRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "role-menu/gormcli")),
	}
}

// AssignMenus replaces all menu bindings for a role in a single transaction.
func (r *RoleMenuRepo) AssignMenus(ctx context.Context, roleId uint32, menuIds []uint32, operatorId uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleId).Delete(&model.SysRoleMenu{}).Error; err != nil {
			r.log.Errorf("delete old role menus failed: %s", err.Error())
			return userV1.ErrorInternalServerError("delete old role menus failed")
		}
		if len(menuIds) == 0 {
			return nil
		}
		now := time.Now()
		rows := make([]*model.SysRoleMenu, 0, len(menuIds))
		for _, menuId := range menuIds {
			rows = append(rows, &model.SysRoleMenu{
				RoleID:    int64(roleId),
				MenuID:    int64(menuId),
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: int64(operatorId),
				UpdatedBy: int64(operatorId),
			})
		}
		if err := tx.CreateInBatches(rows, 100).Error; err != nil {
			r.log.Errorf("assign menus to role failed: %s", err.Error())
			return userV1.ErrorInternalServerError("assign menus to role failed")
		}
		return nil
	})
}

// GetMenuIdsByRoleId returns menu ids bound to the role.
func (r *RoleMenuRepo) GetMenuIdsByRoleId(ctx context.Context, roleId uint32) ([]uint32, error) {
	var ids []uint32
	if err := r.db.WithContext(ctx).
		Table(model.TableNameSysRoleMenu).
		Where("role_id = ?", roleId).
		Pluck("menu_id", &ids).Error; err != nil {
		r.log.Errorf("query menu ids by role id failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query menu ids by role id failed")
	}
	return ids, nil
}

// RemoveMenus removes specific menu bindings for a role.
func (r *RoleMenuRepo) RemoveMenus(ctx context.Context, roleId uint32, menuIds []uint32) error {
	if len(menuIds) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).
		Where("role_id = ? AND menu_id IN ?", roleId, menuIds).
		Delete(&model.SysRoleMenu{}).Error; err != nil {
		r.log.Errorf("remove menus from role failed: %s", err.Error())
		return userV1.ErrorInternalServerError("remove menus from role failed")
	}
	return nil
}
