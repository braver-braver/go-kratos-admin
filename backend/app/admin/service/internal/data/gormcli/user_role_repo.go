package gormcli

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type UserRoleRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewUserRoleRepo(db *gorm.DB, logger log.Logger) *UserRoleRepo {
	return &UserRoleRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "user-role/gormcli")),
	}
}

func (r *UserRoleRepo) GetUserRoleIdsByUserIds(ctx context.Context, userIds []uint32) (map[uint32][]uint32, error) {
	if len(userIds) == 0 {
		return map[uint32][]uint32{}, nil
	}
	var rows []model.SysUserRole
	if err := r.db.WithContext(ctx).
		Where("user_id IN ?", userIds).
		Find(&rows).Error; err != nil {
		r.log.Errorf("query user roles failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query user roles failed")
	}
	result := make(map[uint32][]uint32)
	for _, row := range rows {
		uid := uint32(row.UserID)
		result[uid] = append(result[uid], uint32(row.RoleID))
	}
	return result, nil
}

// Bulk replace roles for a user.
func (r *UserRoleRepo) ReplaceUserRoles(ctx context.Context, userId uint32, roleIds []uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).Delete(&model.SysUserRole{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, rid := range roleIds {
			rec := &model.SysUserRole{
				UserID:    int64(userId),
				RoleID:    int64(rid),
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := tx.Create(rec).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AssignRoles replaces all role bindings for a user in a single transaction.
func (r *UserRoleRepo) AssignRoles(ctx context.Context, userId uint32, ids []uint32, operatorId uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).Delete(&model.SysUserRole{}).Error; err != nil {
			r.log.Errorf("delete old user roles failed: %s", err.Error())
			return userV1.ErrorInternalServerError("delete old user roles failed")
		}
		if len(ids) == 0 {
			return nil
		}
		now := time.Now()
		records := make([]*model.SysUserRole, 0, len(ids))
		for _, rid := range ids {
			records = append(records, &model.SysUserRole{
				UserID:    int64(userId),
				RoleID:    int64(rid),
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: int64(operatorId),
				UpdatedBy: int64(operatorId),
			})
		}
		if err := tx.CreateInBatches(records, 100).Error; err != nil {
			r.log.Errorf("assign roles to user failed: %s", err.Error())
			return userV1.ErrorInternalServerError("assign roles to user failed")
		}
		return nil
	})
}

// GetRoleIdsByUserId returns all role IDs for a user.
func (r *UserRoleRepo) GetRoleIdsByUserId(ctx context.Context, userId uint32) ([]uint32, error) {
	var ids []uint32
	if err := r.db.WithContext(ctx).
		Table(model.TableNameSysUserRole).
		Where("user_id = ?", userId).
		Pluck("role_id", &ids).Error; err != nil {
		r.log.Errorf("query role ids by user id failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query role ids by user id failed")
	}
	return ids, nil
}

// RemoveRoles deletes specific role bindings for a user.
func (r *UserRoleRepo) RemoveRoles(ctx context.Context, userId uint32, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id IN ?", userId, ids).
		Delete(&model.SysUserRole{}).Error; err != nil {
		r.log.Errorf("remove roles from user failed: %s", err.Error())
		return userV1.ErrorInternalServerError("remove roles from user failed")
	}
	return nil
}
