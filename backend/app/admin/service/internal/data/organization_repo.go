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

// OrganizationRepo is the GORM-backed repository for organizations.
type OrganizationRepo struct {
	log *log.Helper
	q   *query.Query
}

func NewOrganizationRepo(data *Data, logger log.Logger) *OrganizationRepo {
	return &OrganizationRepo{
		log: log.NewHelper(log.With(logger, "module", "organization/repo/admin-service")),
		q:   query.Use(data.db),
	}
}

func (r *OrganizationRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	_ = conditions
	return r.q.Organization.WithContext(ctx).Count()
}

func (r *OrganizationRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListOrganizationResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.q.Organization.WithContext(ctx).Order(r.q.Organization.SortID.Desc())

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

	total, err := r.q.Organization.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query count failed")
	}

	items := make([]*userV1.Organization, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}

	return &userV1.ListOrganizationResponse{
		Total: uint32(total),
		Items: items,
	}, nil
}

func (r *OrganizationRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	count, err := r.q.Organization.WithContext(ctx).
		Where(r.q.Organization.ID.Eq(int32(id))).
		Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *OrganizationRepo) Get(ctx context.Context, req *userV1.GetOrganizationRequest) (*userV1.Organization, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.q.Organization.WithContext(ctx).
		Where(r.q.Organization.ID.Eq(int32(req.GetId()))).
		First()
	if err != nil {
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorNotFound("organization not found")
	}

	return r.toDTO(entity), nil
}

func (r *OrganizationRepo) GetOrganizationsByIds(ctx context.Context, ids []uint32) ([]*userV1.Organization, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	intIDs := make([]int32, 0, len(ids))
	for _, id := range ids {
		intIDs = append(intIDs, int32(id))
	}
	entities, err := r.q.Organization.WithContext(ctx).
		Where(r.q.Organization.ID.In(intIDs...)).
		Find()
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}
	items := make([]*userV1.Organization, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}
	return items, nil
}

func (r *OrganizationRepo) Create(ctx context.Context, req *userV1.CreateOrganizationRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	now := time.Now()
	entity := &model.Organization{
		CreatedAt: &now,
		UpdatedAt: &now,
		Status:    organizationStatusToString(req.Data.Status),
		CreateBy:  cloneInt64FromUint32(req.Data.CreatedBy),
		UpdateBy:  cloneInt64FromUint32(req.Data.UpdatedBy),
		Remark:    cloneStringPtr(req.Data.Remark),
		TenantID:  cloneInt64FromUint32(req.Data.TenantId),
		Name:      cloneStringPtr(req.Data.Name),
		SortID:    req.Data.SortOrder,
		ParentID:  cloneInt32FromUint32(req.Data.ParentId),
	}

	return r.q.Organization.WithContext(ctx).Create(entity)
}

func (r *OrganizationRepo) Update(ctx context.Context, req *userV1.UpdateOrganizationRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	update := map[string]any{}
	if req.Data.Name != nil {
		update["name"] = req.Data.GetName()
	}
	if req.Data.Status != nil {
		update["status"] = req.Data.GetStatus().String()
	}
	if req.Data.Remark != nil {
		update["remark"] = req.Data.GetRemark()
	}
	if req.Data.SortOrder != nil {
		update["sort_id"] = req.Data.GetSortOrder()
	}
	if req.Data.ParentId != nil {
		update["parent_id"] = req.Data.GetParentId()
	}
	if req.Data.UpdatedBy != nil {
		update["update_by"] = req.Data.GetUpdatedBy()
	}
	now := time.Now()
	update["updated_at"] = now

	_, err := r.q.Organization.WithContext(ctx).
		Where(r.q.Organization.ID.Eq(int32(req.Data.GetId()))).
		Updates(update)
	if err != nil {
		r.log.Errorf("update data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *OrganizationRepo) Delete(ctx context.Context, req *userV1.DeleteOrganizationRequest) error {
	if req == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}
	_, err := r.q.Organization.WithContext(ctx).
		Where(r.q.Organization.ID.Eq(int32(req.GetId()))).
		Delete()
	if err != nil {
		r.log.Errorf("delete data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete data failed")
	}
	return nil
}

func (r *OrganizationRepo) toDTO(entity *model.Organization) *userV1.Organization {
	if entity == nil {
		return nil
	}
	dto := &userV1.Organization{
		Id:        datautil.CloneUint32(uint32(entity.ID)),
		Name:      cloneStringPtr(entity.Name),
		SortOrder: cloneInt32Ptr(entity.SortID),
		Remark:    cloneStringPtr(entity.Remark),
		TenantId:  datautil.CloneUint32(toUint32(entity.TenantID)),
		CreatedBy: datautil.CloneUint32(toUint32(entity.CreateBy)),
		UpdatedBy: datautil.CloneUint32(toUint32(entity.UpdateBy)),
		ParentId:  cloneUint32FromInt32(entity.ParentID),
		Status:    stringToOrganizationStatus(entity.Status),
		CreatedAt: timeutil.TimeToTimestamppb(entity.CreatedAt),
		UpdatedAt: timeutil.TimeToTimestamppb(entity.UpdatedAt),
		DeletedAt: timeutil.TimeToTimestamppb(entity.DeletedAt),
	}
	return dto
}

func organizationStatusToString(status *userV1.Organization_Status) *string {
	if status == nil {
		return nil
	}
	s := status.String()
	return &s
}

func stringToOrganizationStatus(s *string) *userV1.Organization_Status {
	if s == nil {
		return nil
	}
	if val, ok := userV1.Organization_Status_value[*s]; ok {
		enum := userV1.Organization_Status(val)
		return &enum
	}
	return nil
}
