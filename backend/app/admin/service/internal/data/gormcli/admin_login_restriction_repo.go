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

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type AdminLoginRestrictionRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewAdminLoginRestrictionRepo(db *gorm.DB, logger log.Logger) *AdminLoginRestrictionRepo {
	return &AdminLoginRestrictionRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "admin-login-restriction/gormcli")),
	}
}

func (r *AdminLoginRestrictionRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysAdminLoginRestriction](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *AdminLoginRestrictionRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListAdminLoginRestrictionResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderAdminLoginRestriction(req.GetOrderBy()),
		scopeFieldMaskAdminLoginRestriction(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)
	filterScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	filterScopes = removeNilScopes(filterScopes)

	g := gorm.G[model.SysAdminLoginRestriction](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(filterScopes...)
	total, err := r.Count(ctx, filterScopes...)
	if err != nil {
		return nil, err
	}

	entities, err := g.Find(ctx)
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query list failed")
	}

	items := make([]*adminV1.AdminLoginRestriction, 0, len(entities))
	for i := range entities {
		items = append(items, toAdminLoginRestrictionDTO(&entities[i]))
	}

	return &adminV1.ListAdminLoginRestrictionResponse{
		Total: uint32(total),
		Items: items,
	}, nil
}

func (r *AdminLoginRestrictionRepo) Get(ctx context.Context, req *adminV1.GetAdminLoginRestrictionRequest) (*adminV1.AdminLoginRestriction, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysAdminLoginRestriction](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskAdminLoginRestriction(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, adminV1.ErrorNotFound("admin login restriction not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}
	return toAdminLoginRestrictionDTO(&entity), nil
}

func (r *AdminLoginRestrictionRepo) Create(ctx context.Context, req *adminV1.CreateAdminLoginRestrictionRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	entity := toAdminLoginRestrictionModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}
	if err := gorm.G[model.SysAdminLoginRestriction](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *AdminLoginRestrictionRepo) Update(ctx context.Context, req *adminV1.UpdateAdminLoginRestrictionRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &adminV1.CreateAdminLoginRestrictionRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch p {
		case "target_id":
			updates["target_id"] = req.Data.GetTargetId()
		case "type":
			updates["type"] = req.Data.GetType().String()
		case "method":
			updates["method"] = req.Data.GetMethod().String()
		case "value":
			updates["value"] = req.Data.GetValue()
		case "reason":
			updates["reason"] = req.Data.GetReason()
		case "updated_by":
			updates["updated_by"] = req.Data.GetUpdatedBy()
		case "updated_at":
			if req.Data.GetUpdatedAt() != nil {
				updates["updated_at"] = req.Data.GetUpdatedAt().AsTime()
			}
		}
	}
	if len(updates) == 0 {
		return nil
	}
	if _, ok := updates["updated_at"]; !ok {
		updates["updated_at"] = time.Now()
	}

	if err := r.db.WithContext(ctx).Model(&model.SysAdminLoginRestriction{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *AdminLoginRestrictionRepo) Delete(ctx context.Context, req *adminV1.DeleteAdminLoginRestrictionRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).Where("id = ?", req.GetId()).Delete(&model.SysAdminLoginRestriction{}).Error; err != nil {
		r.log.Errorf("delete data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete data failed")
	}
	return nil
}

func (r *AdminLoginRestrictionRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysAdminLoginRestriction](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

var adminLoginRestrictionColumns = map[string]string{
	"id":         "id",
	"target_id":  "target_id",
	"type":       "type",
	"method":     "method",
	"value":      "value",
	"reason":     "reason",
	"created_by": "created_by",
	"updated_by": "updated_by",
	"deleted_by": "deleted_by",
	"created_at": "created_at",
	"updated_at": "updated_at",
	"deleted_at": "deleted_at",
}

func scopeOrderAdminLoginRestriction(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := adminLoginRestrictionColumns[key]; ok {
				cols = append(cols, clause.OrderByColumn{Column: clause.Column{Name: col}, Desc: desc})
			}
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskAdminLoginRestriction(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := adminLoginRestrictionColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func toAdminLoginRestrictionDTO(entity *model.SysAdminLoginRestriction) *adminV1.AdminLoginRestriction {
	dto := &adminV1.AdminLoginRestriction{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	dto.TargetId = toUint32Ptr(entity.TargetID)
	dto.Type = parseAdminLoginRestrictionType(entity.Type)
	dto.Method = parseAdminLoginRestrictionMethod(entity.Method)
	dto.Value = &entity.Value
	dto.Reason = &entity.Reason
	dto.CreatedBy = toUint32Ptr(entity.CreatedBy)
	dto.UpdatedBy = toUint32Ptr(entity.UpdatedBy)
	dto.DeletedBy = toUint32Ptr(entity.DeletedBy)
	if !entity.CreatedAt.IsZero() {
		dto.CreatedAt = timestamppb.New(entity.CreatedAt)
	}
	if !entity.UpdatedAt.IsZero() {
		dto.UpdatedAt = timestamppb.New(entity.UpdatedAt)
	}
	if entity.DeletedAt.Valid {
		dto.DeletedAt = timestamppb.New(entity.DeletedAt.Time)
	}
	return dto
}

func toAdminLoginRestrictionModel(dto *adminV1.AdminLoginRestriction) *model.SysAdminLoginRestriction {
	if dto == nil {
		return nil
	}
	entity := &model.SysAdminLoginRestriction{
		ID:        int64(dto.GetId()),
		TargetID:  int64(dto.GetTargetId()),
		Value:     dto.GetValue(),
		Reason:    dto.GetReason(),
		CreatedBy: int64(dto.GetCreatedBy()),
		UpdatedBy: int64(dto.GetUpdatedBy()),
		DeletedBy: int64(dto.GetDeletedBy()),
	}
	if dto.Type != nil {
		entity.Type = dto.GetType().String()
	}
	if dto.Method != nil {
		entity.Method = dto.GetMethod().String()
	}
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.CreatedAt.AsTime()
	}
	if dto.UpdatedAt != nil {
		entity.UpdatedAt = dto.UpdatedAt.AsTime()
	}
	if dto.DeletedAt != nil {
		entity.DeletedAt = gorm.DeletedAt{Time: dto.DeletedAt.AsTime(), Valid: true}
	}
	return entity
}

func parseAdminLoginRestrictionType(v string) *adminV1.AdminLoginRestriction_Type {
	if v == "" {
		return nil
	}
	if val, ok := adminV1.AdminLoginRestriction_Type_value[v]; ok {
		t := adminV1.AdminLoginRestriction_Type(val)
		return &t
	}
	return nil
}

func parseAdminLoginRestrictionMethod(v string) *adminV1.AdminLoginRestriction_Method {
	if v == "" {
		return nil
	}
	if val, ok := adminV1.AdminLoginRestriction_Method_value[v]; ok {
		t := adminV1.AdminLoginRestriction_Method(val)
		return &t
	}
	return nil
}
