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

	dictV1 "kratos-admin/api/gen/go/dict/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type DictTypeRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewDictTypeRepo(db *gorm.DB, logger log.Logger) *DictTypeRepo {
	return &DictTypeRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "dict-type/gormcli")),
	}
}

func (r *DictTypeRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysDictType](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, dictV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *DictTypeRepo) List(ctx context.Context, req *pagination.PagingRequest) (*dictV1.ListDictTypeResponse, error) {
	if req == nil {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderDictType(req.GetOrderBy()),
		scopeFieldMaskDictType(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysDictType](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
	total, err := r.Count(ctx, fScopes...)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("query count failed")
	}

	entities, err := g.Find(ctx)
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("query list failed")
	}

	dtos := make([]*dictV1.DictType, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toDictTypeDTO(&entities[i]))
	}

	return &dictV1.ListDictTypeResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *DictTypeRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysDictType](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, dictV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *DictTypeRepo) Get(ctx context.Context, req *dictV1.GetDictTypeRequest) (*dictV1.DictType, error) {
	if req == nil {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}

	g := gorm.G[model.SysDictType](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskDictType(req.GetViewMask().GetPaths()))

	var (
		entity model.SysDictType
		err    error
	)

	switch req.GetQueryBy().(type) {
	case *dictV1.GetDictTypeRequest_Id:
		entity, err = g.Where("id = ?", req.GetId()).Take(ctx)
	case *dictV1.GetDictTypeRequest_Code:
		entity, err = g.Where("type_code = ?", req.GetCode()).Take(ctx)
	default:
		return nil, dictV1.ErrorBadRequest("invalid query parameter")
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dictV1.ErrorNotFound("dict not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("query data failed")
	}

	return toDictTypeDTO(&entity), nil
}

func (r *DictTypeRepo) Create(ctx context.Context, req *dictV1.CreateDictTypeRequest) error {
	if req == nil || req.Data == nil {
		return dictV1.ErrorBadRequest("invalid parameter")
	}

	entity := toDictTypeModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.SysDictType](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return dictV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *DictTypeRepo) Update(ctx context.Context, req *dictV1.UpdateDictTypeRequest) error {
	if req == nil || req.Data == nil {
		return dictV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &dictV1.CreateDictTypeRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch p {
		case "type_code":
			updates["type_code"] = req.Data.GetTypeCode()
		case "type_name":
			updates["type_name"] = req.Data.GetTypeName()
		case "is_enabled":
			updates["is_enabled"] = req.Data.GetIsEnabled()
		case "sort_order":
			updates["sort_order"] = req.Data.GetSortOrder()
		case "description":
			updates["description"] = req.Data.GetDescription()
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

	if err := r.db.WithContext(ctx).Model(&model.SysDictType{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return dictV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *DictTypeRepo) Delete(ctx context.Context, id uint32) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SysDictType{}).Error; err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
		return dictV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *DictTypeRepo) BatchDelete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return dictV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.SysDictType{}).Error; err != nil {
		r.log.Errorf("batch delete failed: %s", err.Error())
		return dictV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

var dictTypeColumns = map[string]string{
	"id":          "id",
	"type_code":   "type_code",
	"type_name":   "type_name",
	"is_enabled":  "is_enabled",
	"sort_order":  "sort_order",
	"description": "description",
	"created_by":  "created_by",
	"updated_by":  "updated_by",
	"deleted_by":  "deleted_by",
	"created_at":  "created_at",
	"updated_at":  "updated_at",
	"deleted_at":  "deleted_at",
}

func scopeOrderDictType(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := dictTypeColumns[key]; ok {
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

func scopeFieldMaskDictType(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := dictTypeColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func toDictTypeDTO(entity *model.SysDictType) *dictV1.DictType {
	dto := &dictV1.DictType{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	dto.TypeCode = &entity.TypeCode
	dto.TypeName = &entity.TypeName
	dto.IsEnabled = &entity.IsEnabled
	dto.SortOrder = &entity.SortOrder
	dto.Description = &entity.Description
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

func toDictTypeModel(dto *dictV1.DictType) *model.SysDictType {
	if dto == nil {
		return nil
	}
	entity := &model.SysDictType{
		ID:          int64(dto.GetId()),
		TypeCode:    dto.GetTypeCode(),
		TypeName:    dto.GetTypeName(),
		IsEnabled:   dto.GetIsEnabled(),
		SortOrder:   dto.GetSortOrder(),
		Description: dto.GetDescription(),
		CreatedBy:   int64(dto.GetCreatedBy()),
		UpdatedBy:   int64(dto.GetUpdatedBy()),
		DeletedBy:   int64(dto.GetDeletedBy()),
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
