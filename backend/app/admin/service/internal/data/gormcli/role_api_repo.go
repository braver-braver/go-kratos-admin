package gormcli

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type RoleApiRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewRoleApiRepo(db *gorm.DB, logger log.Logger) *RoleApiRepo {
	return &RoleApiRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "role-api/gormcli")),
	}
}

// AssignApis 给角色分配API
func (r *RoleApiRepo) AssignApis(ctx context.Context, roleId uint32, apiIds []uint32, operatorId uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleId).Delete(&model.SysRoleAPI{}).Error; err != nil {
			r.log.Errorf("delete old role apis failed: %s", err.Error())
			return userV1.ErrorInternalServerError("delete old role apis failed")
		}
		if len(apiIds) == 0 {
			return nil
		}
		now := time.Now()
		rows := make([]*model.SysRoleAPI, 0, len(apiIds))
		for _, apiId := range apiIds {
			rows = append(rows, &model.SysRoleAPI{
				RoleID:    int64(roleId),
				APIID:     int64(apiId),
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: int64(operatorId),
				UpdatedBy: int64(operatorId),
			})
		}
		if err := tx.CreateInBatches(rows, 100).Error; err != nil {
			r.log.Errorf("assign apis to role failed: %s", err.Error())
			return userV1.ErrorInternalServerError("assign apis to role failed")
		}
		return nil
	})
}

// GetApiIdsByRoleId 获取角色分配的API ID列表
func (r *RoleApiRepo) GetApiIdsByRoleId(ctx context.Context, roleId uint32) ([]uint32, error) {
	var ids []uint32
	if err := r.db.WithContext(ctx).
		Table(model.TableNameSysRoleAPI).
		Where("role_id = ?", roleId).
		Pluck("api_id", &ids).Error; err != nil {
		r.log.Errorf("query api ids by role id failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query api ids by role id failed")
	}
	return ids, nil
}

// RemoveApis 从角色移除API
func (r *RoleApiRepo) RemoveApis(ctx context.Context, roleId uint32, apiIds []uint32) error {
	if err := r.db.WithContext(ctx).
		Where("role_id = ? AND api_id IN ?", roleId, apiIds).
		Delete(&model.SysRoleAPI{}).Error; err != nil {
		r.log.Errorf("remove apis from role failed: %s", err.Error())
		return userV1.ErrorInternalServerError("remove apis from role failed")
	}
	return nil
}
