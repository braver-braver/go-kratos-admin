package data

import (
	"context"
	"time"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/model"
	"kratos-admin/app/admin/service/internal/data/gorm/query"
	"kratos-admin/pkg/datautil"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// PositionRepo is the GORM-backed repository for positions.
type PositionRepo struct {
	log *log.Helper
	q   *query.Query
}

func NewPositionRepo(data *Data, logger log.Logger) *PositionRepo {
	return &PositionRepo{
		log: log.NewHelper(log.With(logger, "module", "position/repo/admin-service")),
		q:   query.Use(data.db),
	}
}

func (r *PositionRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.q.Position.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, userV1.ErrorInternalServerError("query count failed")
	}
	return int(count), nil
}

func (r *PositionRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListPositionResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.q.Position.WithContext(ctx).Order(r.q.Position.SortID.Desc())

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
		return nil, userV1.ErrorInternalServerError("query list failed")
	}

	total, err := r.q.Position.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query count failed")
	}

	items := make([]*userV1.Position, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}

	return &userV1.ListPositionResponse{
		Total: uint32(total),
		Items: items,
	}, nil
}

func (r *PositionRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	count, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.ID.Eq(int32(id))).
		Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *PositionRepo) Get(ctx context.Context, req *userV1.GetPositionRequest) (*userV1.Position, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.ID.Eq(int32(req.GetId()))).
		First()
	if err != nil {
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorNotFound("position not found")
	}
	return r.toDTO(entity), nil
}

func (r *PositionRepo) Create(ctx context.Context, req *userV1.CreatePositionRequest) (*userV1.Position, error) {
	if req == nil || req.Data == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	now := time.Now()
	entity := &model.Position{
		CreatedAt: &now,
		UpdatedAt: &now,
		Status:    positionStatusToString(req.Data.Status),
		CreateBy:  cloneInt64FromUint32(req.Data.CreatedBy),
		UpdateBy:  cloneInt64FromUint32(req.Data.UpdatedBy),
		Remark:    cloneStringPtr(req.Data.Remark),
		TenantID:  cloneInt64FromUint32(req.Data.TenantId),
		Name:      req.Data.GetName(),
		Code:      req.Data.GetCode(),
		SortID:    req.Data.GetSortOrder(),
		ParentID:  cloneInt32FromUint32(req.Data.ParentId),
	}

	if err := r.q.Position.WithContext(ctx).Create(entity); err != nil {
		r.log.Errorf("insert data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("insert data failed")
	}
	return r.toDTO(entity), nil
}

func (r *PositionRepo) Update(ctx context.Context, req *userV1.UpdatePositionRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	update := map[string]any{}
	if req.Data.Name != nil {
		update["name"] = req.Data.GetName()
	}
	if req.Data.Code != nil {
		update["code"] = req.Data.GetCode()
	}
	if req.Data.SortOrder != nil {
		update["sort_id"] = req.Data.GetSortOrder()
	}
	if req.Data.ParentId != nil {
		update["parent_id"] = req.Data.GetParentId()
	}
	if req.Data.Status != nil {
		update["status"] = req.Data.GetStatus().String()
	}
	if req.Data.UpdatedBy != nil {
		update["update_by"] = req.Data.GetUpdatedBy()
	}
	update["updated_at"] = time.Now()

	_, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.ID.Eq(int32(req.Data.GetId()))).
		Updates(update)
	if err != nil {
		r.log.Errorf("update data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *PositionRepo) Delete(ctx context.Context, positionId uint32) error {
	_, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.ID.Eq(int32(positionId))).
		Delete()
	if err != nil {
		r.log.Errorf("delete data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete data failed")
	}
	return nil
}

func (r *PositionRepo) GetPositionsByIds(ctx context.Context, ids []uint32) ([]*userV1.Position, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	intIDs := make([]int32, 0, len(ids))
	for _, id := range ids {
		intIDs = append(intIDs, int32(id))
	}

	entities, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.ID.In(intIDs...)).
		Find()
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	items := make([]*userV1.Position, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}
	return items, nil
}

// Alias to match previous ent-style naming.
func (r *PositionRepo) GetPositionByIds(ctx context.Context, ids []uint32) ([]*userV1.Position, error) {
	return r.GetPositionsByIds(ctx, ids)
}

func (r *PositionRepo) toDTO(entity *model.Position) *userV1.Position {
	if entity == nil {
		return nil
	}
	dto := &userV1.Position{
		Id:        datautil.CloneUint32(uint32(entity.ID)),
		Name:      datautil.CloneString(entity.Name),
		Code:      datautil.CloneString(entity.Code),
		SortOrder: cloneInt32Ptr(&entity.SortID),
		ParentId:  cloneUint32FromInt32(entity.ParentID),
		Status:    stringToPositionStatus(entity.Status),
		TenantId:  datautil.CloneUint32(toUint32(entity.TenantID)),
		CreatedBy: datautil.CloneUint32(toUint32(entity.CreateBy)),
		UpdatedBy: datautil.CloneUint32(toUint32(entity.UpdateBy)),
		Remark:    cloneStringPtr(entity.Remark),
		CreatedAt: timeutil.TimeToTimestamppb(entity.CreatedAt),
		UpdatedAt: timeutil.TimeToTimestamppb(entity.UpdatedAt),
		DeletedAt: timeutil.TimeToTimestamppb(entity.DeletedAt),
	}
	return dto
}

func positionStatusToString(status *userV1.Position_Status) *string {
	if status == nil {
		return nil
	}
	s := status.String()
	return &s
}

func stringToPositionStatus(s *string) *userV1.Position_Status {
	if s == nil {
		return nil
	}
	if val, ok := userV1.Position_Status_value[*s]; ok {
		enum := userV1.Position_Status(val)
		return &enum
	}
	return nil
}
