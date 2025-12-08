package gormcli

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	internalMessageV1 "kratos-admin/api/gen/go/internal_message/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type InternalMessageCategoryRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewInternalMessageCategoryRepo(db *gorm.DB, logger log.Logger) *InternalMessageCategoryRepo {
	return &InternalMessageCategoryRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "internal-message-category/gormcli")),
	}
}

func (r *InternalMessageCategoryRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.InternalMessageCategory](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, internalMessageV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *InternalMessageCategoryRepo) List(ctx context.Context, req *pagination.PagingRequest) (*internalMessageV1.ListInternalMessageCategoryResponse, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderInternalMessageCategory(req.GetOrderBy()),
		scopeFieldMaskInternalMessageCategory(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.InternalMessageCategory](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
	total, err := r.Count(ctx, fScopes...)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query count failed")
	}

	entities, err := g.Find(ctx)
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query list failed")
	}

	sort.SliceStable(entities, func(i, j int) bool {
		return entities[i].ParentID < entities[j].ParentID
	})

	dtos := make([]*internalMessageV1.InternalMessageCategory, 0, len(entities))
	for i := range entities {
		if entities[i].ParentID == 0 {
			dto := toInternalMessageCategoryDTO(&entities[i])
			dtos = append(dtos, dto)
		}
	}
	for i := range entities {
		if entities[i].ParentID != 0 {
			dto := toInternalMessageCategoryDTO(&entities[i])
			if travelInternalMessageCategoryChild(dtos, dto) {
				continue
			}
			dtos = append(dtos, dto)
		}
	}

	return &internalMessageV1.ListInternalMessageCategoryResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func travelInternalMessageCategoryChild(nodes []*internalMessageV1.InternalMessageCategory, node *internalMessageV1.InternalMessageCategory) bool {
	if nodes == nil {
		return false
	}
	if node.ParentId == nil {
		return false
	}
	for _, n := range nodes {
		if n.GetId() == node.GetParentId() {
			n.Children = append(n.Children, node)
			return true
		}
		if travelInternalMessageCategoryChild(n.Children, node) {
			return true
		}
	}
	return false
}

func (r *InternalMessageCategoryRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.InternalMessageCategory](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, internalMessageV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *InternalMessageCategoryRepo) Get(ctx context.Context, req *internalMessageV1.GetInternalMessageCategoryRequest) (*internalMessageV1.InternalMessageCategory, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.InternalMessageCategory](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskInternalMessageCategory(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, internalMessageV1.ErrorNotFound("message category not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query data failed")
	}

	return toInternalMessageCategoryDTO(&entity), nil
}

func (r *InternalMessageCategoryRepo) GetCategoriesByIds(ctx context.Context, ids []uint32) ([]*internalMessageV1.InternalMessageCategory, error) {
	if len(ids) == 0 {
		return []*internalMessageV1.InternalMessageCategory{}, nil
	}

	entities, err := gorm.G[model.InternalMessageCategory](r.db.WithContext(ctx)).
		Where("id IN ?", ids).
		Find(ctx)
	if err != nil {
		r.log.Errorf("query internal message category by ids failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query internal message category by ids failed")
	}

	dtos := make([]*internalMessageV1.InternalMessageCategory, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toInternalMessageCategoryDTO(&entities[i]))
	}
	return dtos, nil
}

func (r *InternalMessageCategoryRepo) Create(ctx context.Context, req *internalMessageV1.CreateInternalMessageCategoryRequest) error {
	if req == nil || req.Data == nil {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	entity := toInternalMessageCategoryModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.InternalMessageCategory](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *InternalMessageCategoryRepo) Update(ctx context.Context, req *internalMessageV1.UpdateInternalMessageCategoryRequest) error {
	if req == nil || req.Data == nil {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &internalMessageV1.CreateInternalMessageCategoryRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch strings.ToLower(p) {
		case "name":
			updates["name"] = req.Data.GetName()
		case "code":
			updates["code"] = req.Data.GetCode()
		case "icon_url":
			updates["icon_url"] = req.Data.GetIconUrl()
		case "parent_id":
			updates["parent_id"] = req.Data.GetParentId()
		case "sort_order":
			updates["sort_order"] = req.Data.GetSortOrder()
		case "is_enabled":
			updates["is_enabled"] = req.Data.GetIsEnabled()
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

	if err := r.db.WithContext(ctx).
		Model(&model.InternalMessageCategory{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *InternalMessageCategoryRepo) Delete(ctx context.Context, req *internalMessageV1.DeleteInternalMessageCategoryRequest) error {
	if req == nil {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	ids := []uint32{req.GetId()}
	queue := []uint32{req.GetId()}

	for len(queue) > 0 {
		var current uint32
		current, queue = queue[0], queue[1:]

		var children []uint32
		if err := r.db.WithContext(ctx).
			Table(model.TableNameInternalMessageCategory).
			Where("parent_id = ?", current).
			Pluck("id", &children).Error; err != nil {
			r.log.Errorf("query child internal message categories failed: %s", err.Error())
			return internalMessageV1.ErrorInternalServerError("query child internal message categories failed")
		}
		if len(children) > 0 {
			ids = append(ids, children...)
			queue = append(queue, children...)
		}
	}

	if err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&model.InternalMessageCategory{}).Error; err != nil {
		r.log.Errorf("delete internal message categories failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("delete internal message categories failed")
	}
	return nil
}

func toInternalMessageCategoryDTO(entity *model.InternalMessageCategory) *internalMessageV1.InternalMessageCategory {
	dto := &internalMessageV1.InternalMessageCategory{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	if entity.Name != "" {
		dto.Name = &entity.Name
	}
	if entity.Code != "" {
		dto.Code = &entity.Code
	}
	if entity.IconURL != "" {
		dto.IconUrl = &entity.IconURL
	}
	dto.ParentId = toUint32Ptr(entity.ParentID)
	dto.SortOrder = &entity.SortOrder
	dto.IsEnabled = &entity.IsEnabled
	dto.CreatedBy = toUint32Ptr(entity.CreatedBy)
	dto.UpdatedBy = toUint32Ptr(entity.UpdatedBy)
	dto.DeletedBy = toUint32Ptr(entity.DeletedBy)
	if !entity.CreatedAt.IsZero() {
		dto.CreatedAt = timestamppb.New(entity.CreatedAt)
	}
	if !entity.UpdatedAt.IsZero() {
		dto.UpdatedAt = timestamppb.New(entity.UpdatedAt)
	}
	if entity.DeletedAt.Valid && entity.DeletedAt.Time.Unix() > 0 {
		dto.DeletedAt = timestamppb.New(entity.DeletedAt.Time)
	}
	return dto
}

func toInternalMessageCategoryModel(dto *internalMessageV1.InternalMessageCategory) *model.InternalMessageCategory {
	entity := &model.InternalMessageCategory{}
	if dto == nil {
		return entity
	}
	if dto.Id != nil {
		entity.ID = int64(dto.GetId())
	}
	if dto.Name != nil {
		entity.Name = dto.GetName()
	}
	if dto.Code != nil {
		entity.Code = dto.GetCode()
	}
	if dto.IconUrl != nil {
		entity.IconURL = dto.GetIconUrl()
	}
	entity.ParentID = int64(dto.GetParentId())
	entity.SortOrder = dto.GetSortOrder()
	entity.IsEnabled = dto.GetIsEnabled()
	entity.CreatedBy = int64(dto.GetCreatedBy())
	entity.UpdatedBy = int64(dto.GetUpdatedBy())
	entity.DeletedBy = int64(dto.GetDeletedBy())
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.GetCreatedAt().AsTime()
	}
	if dto.UpdatedAt != nil {
		entity.UpdatedAt = dto.GetUpdatedAt().AsTime()
	}
	if dto.DeletedAt != nil {
		entity.DeletedAt.Time = dto.GetDeletedAt().AsTime()
		entity.DeletedAt.Valid = true
	}
	return entity
}

var internalMessageCategoryColumns = map[string]string{
	"id":         "id",
	"name":       "name",
	"code":       "code",
	"icon_url":   "icon_url",
	"parent_id":  "parent_id",
	"sort_order": "sort_order",
	"is_enabled": "is_enabled",
	"created_by": "created_by",
	"updated_by": "updated_by",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func scopeOrderInternalMessageCategory(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			col, ok := internalMessageCategoryColumns[strings.ToLower(key)]
			if !ok {
				continue
			}
			cols = append(cols, clause.OrderByColumn{
				Column: clause.Column{Name: col},
				Desc:   desc,
			})
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskInternalMessageCategory(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := internalMessageCategoryColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}
