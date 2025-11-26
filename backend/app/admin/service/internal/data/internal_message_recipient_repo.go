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

// InternalMessageRecipientRepo GORM implementation.
type InternalMessageRecipientRepo struct {
	log *log.Helper
	q   *query.Query
}

func NewInternalMessageRecipientRepo(data *Data, logger log.Logger) *InternalMessageRecipientRepo {
	return &InternalMessageRecipientRepo{
		log: log.NewHelper(log.With(logger, "module", "internal-message-recipient/repo/admin-service")),
		q:   query.Use(data.db),
	}
}

func (r *InternalMessageRecipientRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.q.NotificationMessageRecipient.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, internalMessageV1.ErrorInternalServerError("query count failed")
	}
	return int(count), nil
}

func (r *InternalMessageRecipientRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	count, err := r.q.NotificationMessageRecipient.WithContext(ctx).
		Where(r.q.NotificationMessageRecipient.ID.Eq(int32(id))).
		Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, internalMessageV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *InternalMessageRecipientRepo) List(ctx context.Context, req *pagination.PagingRequest) (*internalMessageV1.ListUserInboxResponse, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.q.NotificationMessageRecipient.WithContext(ctx)
	builder = builder.Order(r.q.NotificationMessageRecipient.CreatedAt.Desc())

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

	total, err := r.q.NotificationMessageRecipient.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query count failed")
	}

	items := make([]*internalMessageV1.InternalMessageRecipient, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}

	return &internalMessageV1.ListUserInboxResponse{
		Total: uint32(total),
		Items: items,
	}, nil
}

func (r *InternalMessageRecipientRepo) Get(ctx context.Context, id uint32) (*internalMessageV1.InternalMessageRecipient, error) {
	entity, err := r.q.NotificationMessageRecipient.WithContext(ctx).
		Where(r.q.NotificationMessageRecipient.ID.Eq(int32(id))).
		First()
	if err != nil {
		return nil, internalMessageV1.ErrorNotFound("message not found")
	}
	return r.toDTO(entity), nil
}

func (r *InternalMessageRecipientRepo) Create(ctx context.Context, req *internalMessageV1.InternalMessageRecipient) (*internalMessageV1.InternalMessageRecipient, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	now := time.Now()
	entity := &model.NotificationMessageRecipient{
		CreatedAt:   &now,
		UpdatedAt:   &now,
		MessageID:   cloneInt64FromUint32(req.MessageId),
		RecipientID: cloneInt64FromUint32(req.RecipientUserId),
		Status:      cloneStringPtr(enumString(req.Status)),
	}

	if err := r.q.NotificationMessageRecipient.WithContext(ctx).Create(entity); err != nil {
		r.log.Errorf("insert data failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("insert data failed")
	}
	return r.toDTO(entity), nil
}

func (r *InternalMessageRecipientRepo) Update(ctx context.Context, req *internalMessageV1.UpdateInternalMessageRecipientRequest) error {
	if req == nil || req.Data == nil {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	update := map[string]any{
		"updated_at": time.Now(),
	}
	if req.Data.Status != nil {
		update["status"] = req.Data.GetStatus().String()
	}

	_, err := r.q.NotificationMessageRecipient.WithContext(ctx).
		Where(r.q.NotificationMessageRecipient.ID.Eq(int32(req.Data.GetId()))).
		Updates(update)
	return err
}

func (r *InternalMessageRecipientRepo) Delete(ctx context.Context, id uint32) error {
	_, err := r.q.NotificationMessageRecipient.WithContext(ctx).
		Where(r.q.NotificationMessageRecipient.ID.Eq(int32(id))).
		Delete()
	return err
}

func (r *InternalMessageRecipientRepo) DeleteNotificationFromInbox(ctx context.Context, req *internalMessageV1.DeleteNotificationFromInboxRequest) error {
	_, err := r.q.NotificationMessageRecipient.WithContext(ctx).
		Where(
			r.q.NotificationMessageRecipient.RecipientID.Eq(int64(req.GetUserId())),
			r.q.NotificationMessageRecipient.ID.In(int32Slice(req.GetRecipientIds())...),
		).
		Delete()
	return err
}

func (r *InternalMessageRecipientRepo) MarkNotificationAsRead(ctx context.Context, req *internalMessageV1.MarkNotificationAsReadRequest) error {
	_, err := r.q.NotificationMessageRecipient.WithContext(ctx).
		Where(
			r.q.NotificationMessageRecipient.ID.In(int32Slice(req.GetRecipientIds())...),
			r.q.NotificationMessageRecipient.RecipientID.Eq(int64(req.GetUserId())),
		).
		Update(r.q.NotificationMessageRecipient.Status, internalMessageV1.InternalMessageRecipient_READ.String())
	return err
}

func (r *InternalMessageRecipientRepo) MarkNotificationsStatus(ctx context.Context, req *internalMessageV1.MarkNotificationsStatusRequest) error {
	_, err := r.q.NotificationMessageRecipient.WithContext(ctx).
		Where(
			r.q.NotificationMessageRecipient.ID.In(int32Slice(req.GetRecipientIds())...),
			r.q.NotificationMessageRecipient.RecipientID.Eq(int64(req.GetUserId())),
		).
		Update(r.q.NotificationMessageRecipient.Status, req.GetNewStatus().String())
	return err
}

func (r *InternalMessageRecipientRepo) RevokeMessage(ctx context.Context, req *internalMessageV1.RevokeMessageRequest) error {
	_, err := r.q.NotificationMessageRecipient.WithContext(ctx).
		Where(
			r.q.NotificationMessageRecipient.MessageID.Eq(int64(req.GetMessageId())),
			r.q.NotificationMessageRecipient.RecipientID.Eq(int64(req.GetUserId())),
		).
		Delete()
	return err
}

func (r *InternalMessageRecipientRepo) toDTO(entity *model.NotificationMessageRecipient) *internalMessageV1.InternalMessageRecipient {
	if entity == nil {
		return nil
	}
	return &internalMessageV1.InternalMessageRecipient{
		Id:              datautil.CloneUint32(uint32(entity.ID)),
		MessageId:       datautil.CloneUint32(toUint32(entity.MessageID)),
		RecipientUserId: datautil.CloneUint32(toUint32(entity.RecipientID)),
		Status:          stringToRecipientStatus(entity.Status),
		CreatedAt:       timeutil.TimeToTimestamppb(entity.CreatedAt),
		UpdatedAt:       timeutil.TimeToTimestamppb(entity.UpdatedAt),
		DeletedAt:       timeutil.TimeToTimestamppb(entity.DeletedAt),
	}
}

func stringToRecipientStatus(s *string) *internalMessageV1.InternalMessageRecipient_Status {
	if s == nil {
		return nil
	}
	if v, ok := internalMessageV1.InternalMessageRecipient_Status_value[*s]; ok {
		val := internalMessageV1.InternalMessageRecipient_Status(v)
		return &val
	}
	return nil
}

func int32Slice(v []uint32) []int32 {
	out := make([]int32, 0, len(v))
	for _, i := range v {
		out = append(out, int32(i))
	}
	return out
}

func enumString(v *internalMessageV1.InternalMessageRecipient_Status) *string {
	if v == nil {
		return nil
	}
	s := v.String()
	return &s
}
