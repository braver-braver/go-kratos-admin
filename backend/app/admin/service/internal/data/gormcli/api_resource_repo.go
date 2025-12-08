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

type ApiResourceRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewApiResourceRepo(db *gorm.DB, logger log.Logger) *ApiResourceRepo {
	return &ApiResourceRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "api-resource/gormcli")),
	}
}

func (r *ApiResourceRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysAPIResource](r.db.WithContext(ctx)).Scopes(scopes...)

	total, err := g.Count(ctx, "id")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *ApiResourceRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListApiResourceResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderApi(req.GetOrderBy()),
		scopeFieldMaskApi(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysAPIResource](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
	total, err := r.Count(ctx, fScopes...)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query count failed")
	}

	entities, err := g.Find(ctx)
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query list failed")
	}

	dtos := make([]*adminV1.ApiResource, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toApiResourceDTO(&entities[i]))
	}

	return &adminV1.ListApiResourceResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *ApiResourceRepo) Get(ctx context.Context, req *adminV1.GetApiResourceRequest) (*adminV1.ApiResource, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysAPIResource](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskApi(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, adminV1.ErrorNotFound("api resource not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}
	return toApiResourceDTO(&entity), nil
}

func (r *ApiResourceRepo) GetByEndpoint(ctx context.Context, path, method string) (*adminV1.ApiResource, error) {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(method) == "" {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysAPIResource](r.db.WithContext(ctx)).
		Where("path = ? AND method = ?", path, method).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, adminV1.ErrorNotFound("api resource not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}
	return toApiResourceDTO(&entity), nil
}

func (r *ApiResourceRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysAPIResource](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *ApiResourceRepo) Create(ctx context.Context, req *adminV1.CreateApiResourceRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	entity := toApiResourceModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}
	if err := gorm.G[model.SysAPIResource](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *ApiResourceRepo) Update(ctx context.Context, req *adminV1.UpdateApiResourceRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		//out, err := r.Get(ctx, &adminV1.GetApiResourceRequest{Id: req.GetData().GetId()})
		if err != nil {
			return err
		}
		if !exist {
			createReq := &adminV1.CreateApiResourceRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, path := range req.GetUpdateMask().GetPaths() {
		switch strings.ToLower(path) {
		case "description":
			updates["description"] = req.Data.GetDescription()
		case "module":
			updates["module"] = req.Data.GetModule()
		case "module_description":
			updates["module_description"] = req.Data.GetModuleDescription()
		case "operation":
			updates["operation"] = req.Data.GetOperation()
		case "path":
			updates["path"] = req.Data.GetPath()
		case "method":
			updates["method"] = req.Data.GetMethod()
		case "scope":
			updates["scope"] = req.Data.GetScope().String()
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

	db := r.db.WithContext(ctx).Model(&model.SysAPIResource{}).Where("id = ?", req.Data.GetId())
	if err := db.Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *ApiResourceRepo) Delete(ctx context.Context, req *adminV1.DeleteApiResourceRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).Where("id = ?", req.GetId()).Delete(&model.SysAPIResource{}).Error; err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *ApiResourceRepo) Truncate(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("1=1").Delete(&model.SysAPIResource{}).Error; err != nil {
		r.log.Errorf("truncate api_resources failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("truncate failed")
	}
	return nil
}

func toApiResourceDTO(entity *model.SysAPIResource) *adminV1.ApiResource {
	dto := &adminV1.ApiResource{}
	if entity == nil {
		return dto
	}
	id := uint32(entity.ID)
	dto.Id = &id
	dto.Description = &entity.Description
	dto.Module = &entity.Module
	dto.ModuleDescription = &entity.ModuleDescription
	dto.Operation = &entity.Operation
	dto.Path = &entity.Path
	dto.Method = &entity.Method
	if entity.Scope != "" {
		if v, ok := adminV1.ApiResource_Scope_value[entity.Scope]; ok {
			scope := adminV1.ApiResource_Scope(v)
			dto.Scope = &scope
		}
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

func toApiResourceModel(dto *adminV1.ApiResource) *model.SysAPIResource {
	if dto == nil {
		return nil
	}
	entity := &model.SysAPIResource{
		ID:                int64(dto.GetId()),
		Description:       dto.GetDescription(),
		Module:            dto.GetModule(),
		ModuleDescription: dto.GetModuleDescription(),
		Operation:         dto.GetOperation(),
		Path:              dto.GetPath(),
		Method:            dto.GetMethod(),
		Scope:             dto.GetScope().String(),
		CreatedBy:         int64(dto.GetCreatedBy()),
		UpdatedBy:         int64(dto.GetUpdatedBy()),
	}
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.CreatedAt.AsTime()
	}
	if dto.UpdatedAt != nil {
		entity.UpdatedAt = dto.UpdatedAt.AsTime()
	}
	return entity
}

var apiResourceColumns = map[string]string{
	"id":                 "id",
	"created_at":         "created_at",
	"updated_at":         "updated_at",
	"description":        "description",
	"module":             "module",
	"module_description": "module_description",
	"operation":          "operation",
	"path":               "path",
	"method":             "method",
	"scope":              "scope",
	"created_by":         "created_by",
	"updated_by":         "updated_by",
}

func scopeOrderApi(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			col, ok := apiResourceColumns[key]
			if !ok {
				continue
			}
			cols = append(
				cols, clause.OrderByColumn{
					Column: clause.Column{Name: col},
					Desc:   desc,
				},
			)
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskApi(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := apiResourceColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func toUint32Ptr(v int64) *uint32 {
	if v == 0 {
		return nil
	}
	val := uint32(v)
	return &val
}

func toTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
