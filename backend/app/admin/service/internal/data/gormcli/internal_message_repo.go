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

type InternalMessageRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewInternalMessageRepo(db *gorm.DB, logger log.Logger) *InternalMessageRepo {
	return &InternalMessageRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "internal-message/gormcli")),
	}
}

func (r *InternalMessageRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.InternalMessage](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, internalMessageV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *InternalMessageRepo) List(ctx context.Context, req *pagination.PagingRequest) (*internalMessageV1.ListInternalMessageResponse, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderInternalMessage(req.GetOrderBy()),
		scopeFieldMaskInternalMessage(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.InternalMessage](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
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

	dtos := make([]*internalMessageV1.InternalMessage, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toInternalMessageDTO(&entities[i]))
	}

	return &internalMessageV1.ListInternalMessageResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *InternalMessageRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.InternalMessage](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, internalMessageV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *InternalMessageRepo) Get(ctx context.Context, req *internalMessageV1.GetInternalMessageRequest) (*internalMessageV1.InternalMessage, error) {
	if req == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.InternalMessage](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskInternalMessage(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, internalMessageV1.ErrorNotFound("message not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("query data failed")
	}

	return toInternalMessageDTO(&entity), nil
}

func (r *InternalMessageRepo) Create(ctx context.Context, req *internalMessageV1.CreateInternalMessageRequest) (*internalMessageV1.InternalMessage, error) {
	if req == nil || req.Data == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	entity := toInternalMessageModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.InternalMessage](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return nil, internalMessageV1.ErrorInternalServerError("insert data failed")
	}
	return toInternalMessageDTO(entity), nil
}

func (r *InternalMessageRepo) Update(ctx context.Context, req *internalMessageV1.UpdateInternalMessageRequest) error {
	if req == nil || req.Data == nil {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &internalMessageV1.CreateInternalMessageRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			_, err = r.Create(ctx, createReq)
			return err
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch strings.ToLower(p) {
		case "title":
			updates["title"] = req.Data.GetTitle()
		case "content":
			updates["content"] = req.Data.GetContent()
		case "sender_id":
			updates["sender_id"] = req.Data.GetSenderId()
		case "category_id":
			updates["category_id"] = req.Data.GetCategoryId()
		case "status":
			updates["status"] = req.Data.GetStatus().String()
		case "type":
			updates["type"] = req.Data.GetType().String()
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
		Model(&model.InternalMessage{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *InternalMessageRepo) Delete(ctx context.Context, id uint32) error {
	if id == 0 {
		return internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.InternalMessage{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return internalMessageV1.ErrorNotFound("internal message not found")
		}
		r.log.Errorf("delete one data failed: %s", err.Error())
		return internalMessageV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func toInternalMessageDTO(entity *model.InternalMessage) *internalMessageV1.InternalMessage {
	dto := &internalMessageV1.InternalMessage{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	if entity.Title != "" {
		dto.Title = &entity.Title
	}
	if entity.Content != "" {
		dto.Content = &entity.Content
	}
	dto.SenderId = toUint32Ptr(entity.SenderID)
	dto.CategoryId = toUint32Ptr(entity.CategoryID)
	if entity.Status != "" {
		if v, ok := internalMessageV1.InternalMessage_Status_value[entity.Status]; ok {
			val := internalMessageV1.InternalMessage_Status(v)
			dto.Status = &val
		}
	}
	if entity.Type != "" {
		if v, ok := internalMessageV1.InternalMessage_Type_value[entity.Type]; ok {
			val := internalMessageV1.InternalMessage_Type(v)
			dto.Type = &val
		}
	}
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

func toInternalMessageModel(dto *internalMessageV1.InternalMessage) *model.InternalMessage {
	entity := &model.InternalMessage{}
	if dto == nil {
		return entity
	}
	if dto.Id != nil {
		entity.ID = int64(dto.GetId())
	}
	if dto.Title != nil {
		entity.Title = dto.GetTitle()
	}
	if dto.Content != nil {
		entity.Content = dto.GetContent()
	}
	entity.SenderID = int64(dto.GetSenderId())
	entity.CategoryID = int64(dto.GetCategoryId())
	if dto.Status != nil {
		entity.Status = dto.GetStatus().String()
	}
	if dto.Type != nil {
		entity.Type = dto.GetType().String()
	}
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

var internalMessageColumns = map[string]string{
	"id":          "id",
	"title":       "title",
	"content":     "content",
	"sender_id":   "sender_id",
	"category_id": "category_id",
	"status":      "status",
	"type":        "type",
	"created_by":  "created_by",
	"updated_by":  "updated_by",
	"created_at":  "created_at",
	"updated_at":  "updated_at",
}

func scopeOrderInternalMessage(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			col, ok := internalMessageColumns[strings.ToLower(key)]
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

func scopeFieldMaskInternalMessage(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := internalMessageColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}
