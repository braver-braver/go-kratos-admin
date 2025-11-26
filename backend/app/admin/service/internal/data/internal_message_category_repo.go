package data

import (
	"context"
	"time"

	internalMessageV1 "kratos-admin/api/gen/go/internal_message/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/model"
	"kratos-admin/app/admin/service/internal/data/gorm/query"
	"kratos-admin/pkg/datautil"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// InternalMessageCategoryRepo GORM implementation.
type InternalMessageCategoryRepo struct {
	log *log.Helper
	q   *query.Query
}

func NewInternalMessageCategoryRepo(data *Data, logger log.Logger) *InternalMessageCategoryRepo {
	return &InternalMessageCategoryRepo{
		log: log.NewHelper(log.With(logger, "module", "internal-message-category/repo/admin-service")),
		q:   query.Use(data.db),
	}
}

func (r *InternalMessageCategoryRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.q.NotificationMessageCategory.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, internalMessageV1.ErrorInternalServerError("query count failed")
	}
	return int(count), nil
}

func (r *InternalMessageCategoryRepo) List(ctx context.Context, req *pagination.PagingRequest) (*internalMessageV1.ListInternalMessageCategoryResponse, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.q.NotificationMessageCategory.WithContext(ctx).Order(r.q.NotificationMessageCategory.SortID.Desc())

	if !req.GetNoPaging() {
		ps := int(req.GetPageSize())
		if ps <= 0 {
			ps = 10
		}
		offset := int(req.GetPage()-1) * ps
		if offset < 0 {
			offset = 0
		}
		builder = builder.Offset(offset).Limit(ps)
	}

	entities, err := builder.Find()
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query list failed")
	}

	total, err := r.q.NotificationMessageCategory.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query count failed")
	}

	items := make([]*internalMessageV1.InternalMessageCategory, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}

	return &internalMessageV1.ListInternalMessageCategoryResponse{
		Total: uint32(total),
		Items: items,
	}, nil
}

func (r *InternalMessageCategoryRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	count, err := r.q.NotificationMessageCategory.WithContext(ctx).
		Where(r.q.NotificationMessageCategory.ID.Eq(int32(id))).
		Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, internalMessageV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *InternalMessageCategoryRepo) Get(ctx context.Context, req *internalMessageV1.GetInternalMessageCategoryRequest) (*internalMessageV1.InternalMessageCategory, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := r.q.NotificationMessageCategory.WithContext(ctx).
		Where(r.q.NotificationMessageCategory.ID.Eq(int32(req.GetId()))).
		First()
	if err != nil {
		return nil, internalMessageV1.ErrorNotFound("category not found")
	}
	return r.toDTO(entity), nil
}

func (r *InternalMessageCategoryRepo) GetCategoriesByIds(ctx context.Context, ids []uint32) ([]*internalMessageV1.InternalMessageCategory, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	intIDs := make([]int32, 0, len(ids))
	for _, id := range ids {
		intIDs = append(intIDs, int32(id))
	}
	entities, err := r.q.NotificationMessageCategory.WithContext(ctx).
		Where(r.q.NotificationMessageCategory.ID.In(intIDs...)).
		Find()
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query data failed")
	}
	items := make([]*internalMessageV1.InternalMessageCategory, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}
	return items, nil
}

func (r *InternalMessageCategoryRepo) Create(ctx context.Context, req *internalMessageV1.CreateInternalMessageCategoryRequest) error {
	if req == nil || req.Data == nil {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}
	now := time.Now()
	entity := &model.NotificationMessageCategory{
		CreatedAt: &now,
		UpdatedAt: &now,
		CreateBy:  cloneInt64FromUint32(req.Data.CreatedBy),
		UpdateBy:  cloneInt64FromUint32(req.Data.UpdatedBy),
		Name:      cloneStringPtr(req.Data.Name),
		Code:      cloneStringPtr(req.Data.Code),
		SortID:    req.Data.SortOrder,
		Enable:    cloneBoolPtr(req.Data.IsEnabled),
		ParentID:  cloneInt32FromUint32(req.Data.ParentId),
	}
	return r.q.NotificationMessageCategory.WithContext(ctx).Create(entity)
}

func (r *InternalMessageCategoryRepo) Update(ctx context.Context, req *internalMessageV1.UpdateInternalMessageCategoryRequest) error {
	if req == nil || req.Data == nil {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}
	update := map[string]any{
		"updated_at": time.Now(),
	}
	if req.Data.Name != nil {
		update["name"] = req.Data.GetName()
	}
	if req.Data.Code != nil {
		update["code"] = req.Data.GetCode()
	}
	if req.Data.SortOrder != nil {
		update["sort_id"] = req.Data.GetSortOrder()
	}
	if req.Data.IsEnabled != nil {
		update["enable"] = req.Data.GetIsEnabled()
	}
	if req.Data.ParentId != nil {
		update["parent_id"] = int32(req.Data.GetParentId())
	}
	if req.Data.UpdatedBy != nil {
		update["update_by"] = req.Data.GetUpdatedBy()
	}

	_, err := r.q.NotificationMessageCategory.WithContext(ctx).
		Where(r.q.NotificationMessageCategory.ID.Eq(int32(req.Data.GetId()))).
		Updates(update)
	return err
}

func (r *InternalMessageCategoryRepo) Delete(ctx context.Context, id uint32) error {
	_, err := r.q.NotificationMessageCategory.WithContext(ctx).
		Where(r.q.NotificationMessageCategory.ID.Eq(int32(id))).
		Delete()
	return err
}

func (r *InternalMessageCategoryRepo) toDTO(entity *model.NotificationMessageCategory) *internalMessageV1.InternalMessageCategory {
	if entity == nil {
		return nil
	}
	return &internalMessageV1.InternalMessageCategory{
		Id:        datautil.CloneUint32(uint32(entity.ID)),
		Name:      cloneStringPtr(entity.Name),
		Code:      cloneStringPtr(entity.Code),
		SortOrder: entity.SortID,
		IsEnabled: cloneBoolPtr(entity.Enable),
		ParentId:  cloneUint32FromInt32(entity.ParentID),
		CreatedBy: datautil.CloneUint32(toUint32(entity.CreateBy)),
		UpdatedBy: datautil.CloneUint32(toUint32(entity.UpdateBy)),
		CreatedAt: timeutil.TimeToTimestamppb(entity.CreatedAt),
		UpdatedAt: timeutil.TimeToTimestamppb(entity.UpdatedAt),
		DeletedAt: timeutil.TimeToTimestamppb(entity.DeletedAt),
	}
}

func cloneInt32Ptr(v *int32) *int32 {
	if v == nil {
		return nil
	}
	val := *v
	return &val
}

func cloneBoolPtr(v *bool) *bool {
	if v == nil {
		return nil
	}
	val := *v
	return &val
}

func cloneUint32FromInt32(v *int32) *uint32 {
	if v == nil {
		return nil
	}
	val := uint32(*v)
	return &val
}

func cloneInt32FromUint32(v *uint32) *int32 {
	if v == nil {
		return nil
	}
	val := int32(*v)
	return &val
}
