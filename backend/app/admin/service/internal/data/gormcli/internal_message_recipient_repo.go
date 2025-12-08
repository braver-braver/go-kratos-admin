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

	internalMessageV1 "kratos-admin/api/gen/go/internal_message/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type InternalMessageRecipientRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewInternalMessageRecipientRepo(db *gorm.DB, logger log.Logger) *InternalMessageRecipientRepo {
	return &InternalMessageRecipientRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "internal-message-recipient/gormcli")),
	}
}

func (r *InternalMessageRecipientRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.InternalMessageRecipient](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, internalMessageV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *InternalMessageRecipientRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.InternalMessageRecipient](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, internalMessageV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *InternalMessageRecipientRepo) List(ctx context.Context, req *pagination.PagingRequest) (*internalMessageV1.ListUserInboxResponse, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderInternalMessageRecipient(req.GetOrderBy()),
		scopeFieldMaskInternalMessageRecipient(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.InternalMessageRecipient](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
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

	dtos := make([]*internalMessageV1.InternalMessageRecipient, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toInternalMessageRecipientDTO(&entities[i]))
	}

	return &internalMessageV1.ListUserInboxResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *InternalMessageRecipientRepo) Get(ctx context.Context, id uint32) (*internalMessageV1.InternalMessageRecipient, error) {
	if id == 0 {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.InternalMessageRecipient](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, internalMessageV1.ErrorNotFound("message not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query data failed")
	}

	return toInternalMessageRecipientDTO(&entity), nil
}

func (r *InternalMessageRecipientRepo) Create(ctx context.Context, req *internalMessageV1.InternalMessageRecipient) (*internalMessageV1.InternalMessageRecipient, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	entity := toInternalMessageRecipientModel(req)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.InternalMessageRecipient](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("insert data failed")
	}

	return toInternalMessageRecipientDTO(entity), nil
}

func (r *InternalMessageRecipientRepo) Update(ctx context.Context, req *internalMessageV1.UpdateInternalMessageRecipientRequest) error {
	if req == nil || req.Data == nil {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			req.Data.CreatedBy = req.Data.UpdatedBy
			req.Data.UpdatedBy = nil
			_, err = r.Create(ctx, req.Data)
			return err
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch strings.ToLower(p) {
		case "message_id":
			updates["message_id"] = req.Data.GetMessageId()
		case "recipient_user_id":
			updates["recipient_user_id"] = req.Data.GetRecipientUserId()
		case "status":
			updates["status"] = req.Data.GetStatus().String()
		case "received_at":
			if req.Data.GetReceivedAt() != nil {
				updates["received_at"] = req.Data.GetReceivedAt().AsTime()
			}
		case "read_at":
			if req.Data.GetReadAt() != nil {
				updates["read_at"] = req.Data.GetReadAt().AsTime()
			}
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
		Model(&model.InternalMessageRecipient{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *InternalMessageRecipientRepo) Delete(ctx context.Context, id uint32) error {
	if id == 0 {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.InternalMessageRecipient{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return internalMessageV1.ErrorNotFound("internal message recipient not found")
		}
		r.log.Errorf("delete one data failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *InternalMessageRecipientRepo) MarkNotificationAsRead(ctx context.Context, req *internalMessageV1.MarkNotificationAsReadRequest) error {
	if len(req.GetRecipientIds()) == 0 || req.GetUserId() == 0 {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.InternalMessageRecipient{}).
		Where("id IN ? AND recipient_user_id = ? AND status <> ?", req.GetRecipientIds(), req.GetUserId(), internalMessageV1.InternalMessageRecipient_READ.String()).
		Updates(map[string]any{
			"status":     internalMessageV1.InternalMessageRecipient_READ.String(),
			"read_at":    now,
			"updated_at": now,
		})
	if result.Error != nil {
		r.log.Errorf("mark notification as read failed: %s", result.Error.Error())
		return internalMessageV1.ErrorInternalServerError("mark notification as read failed")
	}
	return nil
}

func (r *InternalMessageRecipientRepo) MarkNotificationsStatus(ctx context.Context, req *internalMessageV1.MarkNotificationsStatusRequest) error {
	if len(req.GetRecipientIds()) == 0 || req.GetUserId() == 0 {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	now := time.Now()
	var readAt *time.Time
	var receivedAt *time.Time
	switch req.GetNewStatus() {
	case internalMessageV1.InternalMessageRecipient_READ:
		readAt = &now
	case internalMessageV1.InternalMessageRecipient_RECEIVED:
		receivedAt = &now
	}

	updates := map[string]any{
		"status":     req.GetNewStatus().String(),
		"updated_at": now,
	}
	if readAt != nil {
		updates["read_at"] = *readAt
	}
	if receivedAt != nil {
		updates["received_at"] = *receivedAt
	}

	result := r.db.WithContext(ctx).
		Model(&model.InternalMessageRecipient{}).
		Where("id IN ? AND recipient_user_id = ? AND status <> ?", req.GetRecipientIds(), req.GetUserId(), req.GetNewStatus().String()).
		Updates(updates)
	if result.Error != nil {
		r.log.Errorf("mark notification status failed: %s", result.Error.Error())
		return internalMessageV1.ErrorInternalServerError("mark notification status failed")
	}
	return nil
}

func (r *InternalMessageRecipientRepo) RevokeMessage(ctx context.Context, req *internalMessageV1.RevokeMessageRequest) error {
	if err := r.db.WithContext(ctx).
		Where("message_id = ? AND recipient_user_id = ?", req.GetMessageId(), req.GetUserId()).
		Delete(&model.InternalMessageRecipient{}).Error; err != nil {
		r.log.Errorf("revoke message failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("revoke message failed")
	}
	return nil
}

func (r *InternalMessageRecipientRepo) DeleteNotificationFromInbox(ctx context.Context, req *internalMessageV1.DeleteNotificationFromInboxRequest) error {
	if err := r.db.WithContext(ctx).
		Where("id IN ? AND recipient_user_id = ?", req.GetRecipientIds(), req.GetUserId()).
		Delete(&model.InternalMessageRecipient{}).Error; err != nil {
		r.log.Errorf("delete notification from inbox failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("delete notification from inbox failed")
	}
	return nil
}

func toInternalMessageRecipientDTO(entity *model.InternalMessageRecipient) *internalMessageV1.InternalMessageRecipient {
	dto := &internalMessageV1.InternalMessageRecipient{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	dto.MessageId = toUint32Ptr(entity.MessageID)
	dto.RecipientUserId = toUint32Ptr(entity.RecipientUserID)
	if entity.Status != "" {
		if v, ok := internalMessageV1.InternalMessageRecipient_Status_value[entity.Status]; ok {
			val := internalMessageV1.InternalMessageRecipient_Status(v)
			dto.Status = &val
		}
	}
	if !entity.ReceivedAt.IsZero() {
		dto.ReceivedAt = timestamppb.New(entity.ReceivedAt)
	}
	if !entity.ReadAt.IsZero() {
		dto.ReadAt = timestamppb.New(entity.ReadAt)
	}
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

func toInternalMessageRecipientModel(dto *internalMessageV1.InternalMessageRecipient) *model.InternalMessageRecipient {
	entity := &model.InternalMessageRecipient{}
	if dto == nil {
		return entity
	}
	if dto.Id != nil {
		entity.ID = int64(dto.GetId())
	}
	entity.MessageID = int64(dto.GetMessageId())
	entity.RecipientUserID = int64(dto.GetRecipientUserId())
	if dto.Status != nil {
		entity.Status = dto.GetStatus().String()
	}
	if dto.ReceivedAt != nil {
		entity.ReceivedAt = dto.GetReceivedAt().AsTime()
	}
	if dto.ReadAt != nil {
		entity.ReadAt = dto.GetReadAt().AsTime()
	}
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

var internalMessageRecipientColumns = map[string]string{
	"id":                "id",
	"message_id":        "message_id",
	"recipient_user_id": "recipient_user_id",
	"status":            "status",
	"received_at":       "received_at",
	"read_at":           "read_at",
	"created_at":        "created_at",
	"updated_at":        "updated_at",
}

func scopeOrderInternalMessageRecipient(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			col, ok := internalMessageRecipientColumns[strings.ToLower(key)]
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

func scopeFieldMaskInternalMessageRecipient(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := internalMessageRecipientColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}
