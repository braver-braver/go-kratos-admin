package gormcli

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type RolePositionRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewRolePositionRepo(db *gorm.DB, logger log.Logger) *RolePositionRepo {
	return &RolePositionRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "role-position/gormcli")),
	}
}

// AssignPositions replaces all position bindings for a role.
func (r *RolePositionRepo) AssignPositions(ctx context.Context, roleId uint32, positionIds []uint32, operatorId uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleId).Delete(&model.SysRolePosition{}).Error; err != nil {
			r.log.Errorf("delete old role positions failed: %s", err.Error())
			return userV1.ErrorInternalServerError("delete old role positions failed")
		}
		if len(positionIds) == 0 {
			return nil
		}
		now := time.Now()
		rows := make([]*model.SysRolePosition, 0, len(positionIds))
		for _, pid := range positionIds {
			rows = append(rows, &model.SysRolePosition{
				RoleID:     int64(roleId),
				PositionID: int64(pid),
				CreatedAt:  now,
				UpdatedAt:  now,
				CreatedBy:  int64(operatorId),
				UpdatedBy:  int64(operatorId),
			})
		}
		if err := tx.CreateInBatches(rows, 100).Error; err != nil {
			r.log.Errorf("assign positions to role failed: %s", err.Error())
			return userV1.ErrorInternalServerError("assign positions to role failed")
		}
		return nil
	})
}

// GetPositionIdsByRoleId returns position ids bound to the role.
func (r *RolePositionRepo) GetPositionIdsByRoleId(ctx context.Context, roleId uint32) ([]uint32, error) {
	var ids []uint32
	if err := r.db.WithContext(ctx).
		Table(model.TableNameSysRolePosition).
		Where("role_id = ?", roleId).
		Pluck("position_id", &ids).Error; err != nil {
		r.log.Errorf("query position ids by role id failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query position ids by role id failed")
	}
	return ids, nil
}

// RemovePositions removes specific position bindings for a role.
func (r *RolePositionRepo) RemovePositions(ctx context.Context, roleId uint32, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).
		Where("role_id = ? AND position_id IN ?", roleId, ids).
		Delete(&model.SysRolePosition{}).Error; err != nil {
		r.log.Errorf("remove positions from role failed: %s", err.Error())
		return userV1.ErrorInternalServerError("remove positions from role failed")
	}
	return nil
}
