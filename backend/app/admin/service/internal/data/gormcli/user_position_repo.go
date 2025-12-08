package gormcli

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type UserPositionRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewUserPositionRepo(db *gorm.DB, logger log.Logger) *UserPositionRepo {
	return &UserPositionRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "user-position/gormcli")),
	}
}

// AssignPositions replaces all position bindings for a user.
func (r *UserPositionRepo) AssignPositions(ctx context.Context, userId uint32, positionIds []uint32, operatorId uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).Delete(&model.SysUserPosition{}).Error; err != nil {
			r.log.Errorf("delete old user positions failed: %s", err.Error())
			return userV1.ErrorInternalServerError("delete old user positions failed")
		}
		if len(positionIds) == 0 {
			return nil
		}
		now := time.Now()
		rows := make([]*model.SysUserPosition, 0, len(positionIds))
		for _, pid := range positionIds {
			rows = append(rows, &model.SysUserPosition{
				UserID:     int64(userId),
				PositionID: int64(pid),
				CreatedAt:  now,
				UpdatedAt:  now,
				CreatedBy:  int64(operatorId),
				UpdatedBy:  int64(operatorId),
			})
		}
		if err := tx.CreateInBatches(rows, 100).Error; err != nil {
			r.log.Errorf("assign positions to user failed: %s", err.Error())
			return userV1.ErrorInternalServerError("assign positions to user failed")
		}
		return nil
	})
}

// GetPositionIdsByUserId returns position ids bound to a user.
func (r *UserPositionRepo) GetPositionIdsByUserId(ctx context.Context, userId uint32) ([]uint32, error) {
	var ids []uint32
	if err := r.db.WithContext(ctx).
		Table(model.TableNameSysUserPosition).
		Where("user_id = ?", userId).
		Pluck("position_id", &ids).Error; err != nil {
		r.log.Errorf("query position ids by user id failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query position ids by user id failed")
	}
	return ids, nil
}

// RemovePositions removes specific position bindings for a user.
func (r *UserPositionRepo) RemovePositions(ctx context.Context, userId uint32, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND position_id IN ?", userId, ids).
		Delete(&model.SysUserPosition{}).Error; err != nil {
		r.log.Errorf("remove positions from user failed: %s", err.Error())
		return userV1.ErrorInternalServerError("remove positions from user failed")
	}
	return nil
}
