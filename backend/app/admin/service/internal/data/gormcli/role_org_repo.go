package gormcli

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type RoleOrgRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewRoleOrgRepo(db *gorm.DB, logger log.Logger) *RoleOrgRepo {
	return &RoleOrgRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "role-org/gormcli")),
	}
}

// AssignOrganizations replaces all org bindings for a role.
func (r *RoleOrgRepo) AssignOrganizations(ctx context.Context, roleId uint32, orgIds []uint32, operatorId uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleId).Delete(&model.SysRoleOrg{}).Error; err != nil {
			r.log.Errorf("delete old role organizations failed: %s", err.Error())
			return userV1.ErrorInternalServerError("delete old role organizations failed")
		}
		if len(orgIds) == 0 {
			return nil
		}
		now := time.Now()
		rows := make([]*model.SysRoleOrg, 0, len(orgIds))
		for _, orgId := range orgIds {
			rows = append(rows, &model.SysRoleOrg{
				RoleID:    int64(roleId),
				OrgID:     int64(orgId),
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: int64(operatorId),
				UpdatedBy: int64(operatorId),
			})
		}
		if err := tx.CreateInBatches(rows, 100).Error; err != nil {
			r.log.Errorf("assign organizations to role failed: %s", err.Error())
			return userV1.ErrorInternalServerError("assign organizations to role failed")
		}
		return nil
	})
}

// GetOrganizationIdsByRoleId returns org ids bound to the role.
func (r *RoleOrgRepo) GetOrganizationIdsByRoleId(ctx context.Context, roleId uint32) ([]uint32, error) {
	var ids []uint32
	if err := r.db.WithContext(ctx).
		Table(model.TableNameSysRoleOrg).
		Where("role_id = ?", roleId).
		Pluck("org_id", &ids).Error; err != nil {
		r.log.Errorf("query organization ids by role id failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query organization ids by role id failed")
	}
	return ids, nil
}

// RemoveOrganizations removes specific org bindings for a role.
func (r *RoleOrgRepo) RemoveOrganizations(ctx context.Context, roleId uint32, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).
		Where("role_id = ? AND org_id IN ?", roleId, ids).
		Delete(&model.SysRoleOrg{}).Error; err != nil {
		r.log.Errorf("remove organizations from role failed: %s", err.Error())
		return userV1.ErrorInternalServerError("remove organizations from role failed")
	}
	return nil
}
