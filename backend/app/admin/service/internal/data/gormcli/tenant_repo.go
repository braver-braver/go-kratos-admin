package gormcli

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type TenantRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewTenantRepo(db *gorm.DB, logger log.Logger) *TenantRepo {
	return &TenantRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "tenant/gormcli")),
	}
}

func (r *TenantRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysTenant](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, userV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *TenantRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListTenantResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderTenant(req.GetOrderBy()),
		scopeFieldMaskTenant(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysTenant](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
	total, err := r.Count(ctx, fScopes...)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query count failed")
	}

	entities, err := g.Find(ctx)
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query list failed")
	}

	dtos := make([]*userV1.Tenant, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toTenantDTO(&entities[i]))
	}

	return &userV1.ListTenantResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *TenantRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysTenant](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *TenantRepo) Get(ctx context.Context, req *userV1.GetTenantRequest) (*userV1.Tenant, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysTenant](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskTenant(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userV1.ErrorTenantNotFound("tenant not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}
	return toTenantDTO(&entity), nil
}

func (r *TenantRepo) Create(ctx context.Context, data *userV1.Tenant) (*userV1.Tenant, error) {
	if data == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	entity := toTenantModel(data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.SysTenant](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("insert data failed")
	}
	return toTenantDTO(entity), nil
}

func (r *TenantRepo) Update(ctx context.Context, req *userV1.UpdateTenantRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createdBy := req.Data.CreatedBy
			createReq := &userV1.Tenant{
				Id:          req.Data.Id,
				Name:        req.Data.Name,
				Code:        req.Data.Code,
				LogoUrl:     req.Data.LogoUrl,
				Remark:      req.Data.Remark,
				Industry:    req.Data.Industry,
				AdminUserId: req.Data.AdminUserId,
				Status:      req.Data.Status,
				Type:        req.Data.Type,
				AuditStatus: req.Data.AuditStatus,
				CreatedBy:   createdBy,
			}
			req.Data.CreatedBy = nil
			_, err := r.Create(ctx, createReq)
			return err
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch strings.ToLower(p) {
		case "name":
			updates["name"] = req.Data.GetName()
		case "code":
			updates["code"] = req.Data.GetCode()
		case "logo_url":
			updates["logo_url"] = req.Data.GetLogoUrl()
		case "remark":
			updates["remark"] = req.Data.GetRemark()
		case "industry":
			updates["industry"] = req.Data.GetIndustry()
		case "admin_user_id":
			updates["admin_user_id"] = req.Data.GetAdminUserId()
		case "status":
			updates["status"] = req.Data.GetStatus().String()
		case "type":
			updates["type"] = req.Data.GetType().String()
		case "audit_status":
			updates["audit_status"] = req.Data.GetAuditStatus().String()
		case "subscription_at":
			updates["subscription_at"] = toTime(req.Data.GetSubscriptionAt())
		case "unsubscribe_at":
			updates["unsubscribe_at"] = toTime(req.Data.GetUnsubscribeAt())
		case "subscription_plan":
			updates["subscription_plan"] = req.Data.GetSubscriptionPlan()
		case "expired_at":
			updates["expired_at"] = toTime(req.Data.GetExpiredAt())
		case "last_login_time":
			updates["last_login_time"] = toTime(req.Data.GetLastLoginTime())
		case "last_login_ip":
			updates["last_login_ip"] = req.Data.GetLastLoginIp()
		case "updated_by":
			updates["updated_by"] = req.Data.GetUpdatedBy()
		case "updated_at":
			updates["updated_at"] = toTime(req.Data.GetUpdatedAt())
		}
	}
	if len(updates) == 0 {
		return nil
	}
	if _, ok := updates["updated_at"]; !ok {
		updates["updated_at"] = time.Now()
	}

	db := r.db.WithContext(ctx).Model(&model.SysTenant{}).Where("id = ?", req.Data.GetId())
	if err := db.Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *TenantRepo) Delete(ctx context.Context, req *userV1.DeleteTenantRequest) error {
	if req == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).Where("id = ?", req.GetId()).Delete(&model.SysTenant{}).Error; err != nil {
		r.log.Errorf("delete tenant failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

// GetTenantByTenantName gets tenant by tenant name.
func (r *TenantRepo) GetTenantByTenantName(ctx context.Context, userName string) (*userV1.Tenant, error) {
	if userName == "" {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysTenant](r.db.WithContext(ctx)).
		Where("name = ?", userName).
		Take(ctx)
	switch {
	case err == nil:
		return toTenantDTO(&entity), nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, userV1.ErrorNotFound("tenant not found")
	default:
		r.log.Errorf("query tenant data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}
}

// GetTenantByTenantCode gets tenant by tenant code.
func (r *TenantRepo) GetTenantByTenantCode(ctx context.Context, code string) (*userV1.Tenant, error) {
	if code == "" {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysTenant](r.db.WithContext(ctx)).
		Where("code = ?", code).
		Take(ctx)
	switch {
	case err == nil:
		return toTenantDTO(&entity), nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, userV1.ErrorNotFound("tenant not found")
	default:
		r.log.Errorf("query tenant data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}
}

// TenantExists checks if a tenant with the given code exists.
func (r *TenantRepo) TenantExists(ctx context.Context, req *userV1.TenantExistsRequest) (*userV1.TenantExistsResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	var count int64
	if err := r.db.WithContext(ctx).
		Table(model.TableNameSysTenant).
		Where("code = ?", req.GetCode()).
		Count(&count).Error; err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query exist failed")
	}
	return &userV1.TenantExistsResponse{Exist: count > 0}, nil
}

// GetTenantsByIds gets tenants by a list of IDs.
func (r *TenantRepo) GetTenantsByIds(ctx context.Context, ids []uint32) ([]*userV1.Tenant, error) {
	if len(ids) == 0 {
		return []*userV1.Tenant{}, nil
	}

	entities, err := gorm.G[model.SysTenant](r.db.WithContext(ctx)).
		Where("id IN ?", ids).
		Find(ctx)
	if err != nil {
		r.log.Errorf("query tenant by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query tenant by ids failed")
	}

	dtos := make([]*userV1.Tenant, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toTenantDTO(&entities[i]))
	}

	return dtos, nil
}

func toTenantDTO(entity *model.SysTenant) *userV1.Tenant {
	dto := &userV1.Tenant{}
	if entity == nil {
		return dto
	}
	id := uint32(entity.ID)
	dto.Id = &id
	dto.Name = &entity.Name
	dto.Code = &entity.Code
	dto.LogoUrl = &entity.LogoURL
	dto.Remark = &entity.Remark
	dto.Industry = &entity.Industry
	dto.AdminUserId = toUint32Ptr(entity.AdminUserID)
	if entity.Status != "" {
		if v, ok := userV1.Tenant_Status_value[entity.Status]; ok {
			status := userV1.Tenant_Status(v)
			dto.Status = &status
		}
	}
	if entity.Type != "" {
		if v, ok := userV1.Tenant_Type_value[entity.Type]; ok {
			tp := userV1.Tenant_Type(v)
			dto.Type = &tp
		}
	}
	if entity.AuditStatus != "" {
		if v, ok := userV1.Tenant_AuditStatus_value[entity.AuditStatus]; ok {
			as := userV1.Tenant_AuditStatus(v)
			dto.AuditStatus = &as
		}
	}
	dto.SubscriptionPlan = &entity.SubscriptionPlan
	if !entity.SubscriptionAt.IsZero() {
		dto.SubscriptionAt = timestamppb.New(entity.SubscriptionAt)
	}
	if !entity.UnsubscribeAt.IsZero() {
		dto.UnsubscribeAt = timestamppb.New(entity.UnsubscribeAt)
	}
	if !entity.ExpiredAt.IsZero() {
		dto.ExpiredAt = timestamppb.New(entity.ExpiredAt)
	}
	if !entity.LastLoginTime.IsZero() {
		dto.LastLoginTime = timestamppb.New(entity.LastLoginTime)
	}
	if entity.LastLoginIP != "" {
		dto.LastLoginIp = &entity.LastLoginIP
	}
	dto.CreatedBy = toUint32Ptr(entity.CreatedBy)
	dto.UpdatedBy = toUint32Ptr(entity.UpdatedBy)
	if !entity.CreatedAt.IsZero() {
		dto.CreatedAt = timestamppb.New(entity.CreatedAt)
	}
	if !entity.UpdatedAt.IsZero() {
		dto.UpdatedAt = timestamppb.New(entity.UpdatedAt)
	}
	return dto
}

func toTenantModel(dto *userV1.Tenant) *model.SysTenant {
	if dto == nil {
		return nil
	}
	entity := &model.SysTenant{
		ID:               int64(dto.GetId()),
		Name:             dto.GetName(),
		Code:             dto.GetCode(),
		LogoURL:          dto.GetLogoUrl(),
		Remark:           dto.GetRemark(),
		Industry:         dto.GetIndustry(),
		AdminUserID:      int64(dto.GetAdminUserId()),
		Status:           dto.GetStatus().String(),
		Type:             dto.GetType().String(),
		AuditStatus:      dto.GetAuditStatus().String(),
		SubscriptionPlan: dto.GetSubscriptionPlan(),
		CreatedBy:        int64(dto.GetCreatedBy()),
		UpdatedBy:        int64(dto.GetUpdatedBy()),
	}
	if dto.SubscriptionAt != nil {
		entity.SubscriptionAt = dto.SubscriptionAt.AsTime()
	}
	if dto.UnsubscribeAt != nil {
		entity.UnsubscribeAt = dto.UnsubscribeAt.AsTime()
	}
	if dto.ExpiredAt != nil {
		entity.ExpiredAt = dto.ExpiredAt.AsTime()
	}
	if dto.LastLoginTime != nil {
		entity.LastLoginTime = dto.LastLoginTime.AsTime()
	}
	if dto.LastLoginIp != nil {
		entity.LastLoginIP = dto.GetLastLoginIp()
	}
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.CreatedAt.AsTime()
	}
	if dto.UpdatedAt != nil {
		entity.UpdatedAt = dto.UpdatedAt.AsTime()
	}
	return entity
}

var tenantColumns = map[string]string{
	"id":                "id",
	"created_at":        "created_at",
	"updated_at":        "updated_at",
	"name":              "name",
	"code":              "code",
	"logo_url":          "logo_url",
	"industry":          "industry",
	"admin_user_id":     "admin_user_id",
	"status":            "status",
	"type":              "type",
	"audit_status":      "audit_status",
	"subscription_at":   "subscription_at",
	"unsubscribe_at":    "unsubscribe_at",
	"subscription_plan": "subscription_plan",
	"expired_at":        "expired_at",
	"last_login_time":   "last_login_time",
	"last_login_ip":     "last_login_ip",
	"created_by":        "created_by",
	"updated_by":        "updated_by",
}

func scopeOrderTenant(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := tenantColumns[key]; ok {
				cols = append(cols, clause.OrderByColumn{
					Column: clause.Column{Name: col},
					Desc:   desc,
				})
			}
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskTenant(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := tenantColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}
