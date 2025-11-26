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

// InternalMessageRepo uses GORM to manage notification messages (formerly Ent).
type InternalMessageRepo struct {
	log *log.Helper
	q   *query.Query
}

func NewInternalMessageRepo(data *Data, logger log.Logger) *InternalMessageRepo {
	return &InternalMessageRepo{
		log: log.NewHelper(log.With(logger, "module", "internal-message/repo/admin-service")),
		q:   query.Use(data.db),
	}
}

func (r *InternalMessageRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.q.NotificationMessage.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, internalMessageV1.ErrorInternalServerError("query count failed")
	}
	return int(count), nil
}

func (r *InternalMessageRepo) List(ctx context.Context, req *pagination.PagingRequest) (*internalMessageV1.ListInternalMessageResponse, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.q.NotificationMessage.WithContext(ctx).Order(r.q.NotificationMessage.CreatedAt.Desc())

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

	total, err := r.q.NotificationMessage.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query count failed")
	}

	items := make([]*internalMessageV1.InternalMessage, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}

	return &internalMessageV1.ListInternalMessageResponse{
		Total: uint32(total),
		Items: items,
	}, nil
}

func (r *InternalMessageRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	count, err := r.q.NotificationMessage.WithContext(ctx).
		Where(r.q.NotificationMessage.ID.Eq(int32(id))).
		Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, internalMessageV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *InternalMessageRepo) Get(ctx context.Context, req *internalMessageV1.GetInternalMessageRequest) (*internalMessageV1.InternalMessage, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := r.q.NotificationMessage.WithContext(ctx).
		Where(r.q.NotificationMessage.ID.Eq(int32(req.GetId()))).
		First()
	if err != nil {
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, internalMessageV1.ErrorNotFound("message not found")
	}
	return r.toDTO(entity), nil
}

func (r *InternalMessageRepo) Create(ctx context.Context, req *internalMessageV1.CreateInternalMessageRequest) (*internalMessageV1.InternalMessage, error) {
	if req == nil || req.Data == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	now := time.Now()
	entity := &model.NotificationMessage{
		CreatedAt:  &now,
		UpdatedAt:  &now,
		CreateBy:   cloneInt64FromUint32(req.Data.CreatedBy),
		UpdateBy:   cloneInt64FromUint32(req.Data.UpdatedBy),
		Subject:    cloneStringPtr(req.Data.Title),
		Content:    cloneStringPtr(req.Data.Content),
		CategoryID: cloneInt64FromUint32(req.Data.CategoryId),
		Status:     messageStatusToString(req.Data.Status),
	}

	if err := r.q.NotificationMessage.WithContext(ctx).Create(entity); err != nil {
		r.log.Errorf("insert data failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("insert data failed")
	}
	return r.toDTO(entity), nil
}

func (r *InternalMessageRepo) Update(ctx context.Context, req *internalMessageV1.UpdateInternalMessageRequest) error {
	if req == nil || req.Data == nil {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	update := map[string]any{
		"updated_at": time.Now(),
	}
	if req.Data.Title != nil {
		update["subject"] = req.Data.GetTitle()
	}
	if req.Data.Content != nil {
		update["content"] = req.Data.GetContent()
	}
	if req.Data.CategoryId != nil {
		update["category_id"] = req.Data.GetCategoryId()
	}
	if req.Data.Status != nil {
		update["status"] = req.Data.GetStatus().String()
	}
	if req.Data.UpdatedBy != nil {
		update["update_by"] = req.Data.GetUpdatedBy()
	}

	_, err := r.q.NotificationMessage.WithContext(ctx).
		Where(r.q.NotificationMessage.ID.Eq(int32(req.Data.GetId()))).
		Updates(update)
	if err != nil {
		r.log.Errorf("update data failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *InternalMessageRepo) Delete(ctx context.Context, id uint32) error {
	_, err := r.q.NotificationMessage.WithContext(ctx).
		Where(r.q.NotificationMessage.ID.Eq(int32(id))).
		Delete()
	if err != nil {
		r.log.Errorf("delete data failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("delete data failed")
	}
	return nil
}

func (r *InternalMessageRepo) toDTO(entity *model.NotificationMessage) *internalMessageV1.InternalMessage {
	if entity == nil {
		return nil
	}
	return &internalMessageV1.InternalMessage{
		Id:         datautil.CloneUint32(uint32(entity.ID)),
		Title:      cloneStringPtr(entity.Subject),
		Content:    cloneStringPtr(entity.Content),
		CategoryId: datautil.CloneUint32(toUint32(entity.CategoryID)),
		Status:     stringToInternalMessageStatus(entity.Status),
		CreatedBy:  datautil.CloneUint32(toUint32(entity.CreateBy)),
		UpdatedBy:  datautil.CloneUint32(toUint32(entity.UpdateBy)),
		CreatedAt:  timeutil.TimeToTimestamppb(entity.CreatedAt),
		UpdatedAt:  timeutil.TimeToTimestamppb(entity.UpdatedAt),
		DeletedAt:  timeutil.TimeToTimestamppb(entity.DeletedAt),
	}
}

func messageStatusToString(v *internalMessageV1.InternalMessage_Status) *string {
	if v == nil {
		return nil
	}
	s := v.String()
	return &s
}

func stringToInternalMessageStatus(s *string) *internalMessageV1.InternalMessage_Status {
	if s == nil {
		return nil
	}
	if v, ok := internalMessageV1.InternalMessage_Status_value[*s]; ok {
		val := internalMessageV1.InternalMessage_Status(v)
		return &val
	}
	return nil
}
