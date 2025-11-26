package data

import (
	"context"
	"strings"
	"time"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/model"
	"kratos-admin/app/admin/service/internal/data/gorm/query"
	"kratos-admin/pkg/datautil"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

// AdminLoginRestrictionRepo is the GORM-backed repository for admin login restrictions.
type AdminLoginRestrictionRepo struct {
	log *log.Helper
	q   *query.Query
}

func NewAdminLoginRestrictionRepo(data *Data, logger log.Logger) *AdminLoginRestrictionRepo {
	return &AdminLoginRestrictionRepo{
		log: log.NewHelper(log.With(logger, "module", "admin-login-restriction/repo/admin-service")),
		q:   query.Use(data.db),
	}
}

func (r *AdminLoginRestrictionRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.q.AdminLoginRestriction.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}
	return int(count), nil
}

func (r *AdminLoginRestrictionRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListAdminLoginRestrictionResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.q.AdminLoginRestriction.WithContext(ctx)

	builder = builder.Order(r.q.AdminLoginRestriction.CreatedAt.Desc())

	if !req.GetNoPaging() {
		pageSize := int(req.GetPageSize())
		if pageSize <= 0 {
			pageSize = 10
		}
		offset := int(req.GetPage()-1) * pageSize
		if offset < 0 {
			offset = 0
		}
		builder = builder.Offset(offset).Limit(pageSize)
	}

	entities, err := builder.Find()
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query list failed")
	}

	count, err := r.q.AdminLoginRestriction.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query count failed")
	}

	items := make([]*adminV1.AdminLoginRestriction, 0, len(entities))
	for _, entity := range entities {
		items = append(items, r.toDTO(entity))
	}

	return &adminV1.ListAdminLoginRestrictionResponse{
		Total: uint32(count),
		Items: items,
	}, nil
}

func (r *AdminLoginRestrictionRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	count, err := r.q.AdminLoginRestriction.WithContext(ctx).
		Where(r.q.AdminLoginRestriction.ID.Eq(int32(id))).
		Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *AdminLoginRestrictionRepo) Get(ctx context.Context, req *adminV1.GetAdminLoginRestrictionRequest) (*adminV1.AdminLoginRestriction, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.q.AdminLoginRestriction.WithContext(ctx).
		Where(r.q.AdminLoginRestriction.ID.Eq(int32(req.GetId()))).
		First()
	if err != nil {
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorNotFound("admin login restriction not found")
	}

	return r.toDTO(entity), nil
}

func (r *AdminLoginRestrictionRepo) Create(ctx context.Context, req *adminV1.CreateAdminLoginRestrictionRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	entity := &model.AdminLoginRestriction{
		CreatedAt: timeutil.TimestamppbToTime(req.Data.CreatedAt),
		UpdatedAt: timeutil.TimestamppbToTime(req.Data.UpdatedAt),
		CreateBy:  cloneInt64FromUint32(req.Data.CreatedBy),
		UpdateBy:  cloneInt64FromUint32(req.Data.UpdatedBy),
		TargetID:  cloneInt64FromUint32(req.Data.TargetId),
		Value:     cloneStringPtr(req.Data.Value),
		Reason:    cloneStringPtr(req.Data.Reason),
		Type:      adminLoginRestrictionTypeToString(req.Data.Type),
		Method:    adminLoginRestrictionMethodToString(req.Data.Method),
	}

	if req.Data.CreatedAt == nil {
		now := time.Now()
		entity.CreatedAt = &now
		entity.UpdatedAt = &now
	}

	return r.q.AdminLoginRestriction.WithContext(ctx).Create(entity)
}

func (r *AdminLoginRestrictionRepo) Update(ctx context.Context, req *adminV1.UpdateAdminLoginRestrictionRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	update := map[string]any{}
	if req.Data.TargetId != nil {
		update["target_id"] = req.Data.GetTargetId()
	}
	if req.Data.Value != nil {
		update["value"] = req.Data.GetValue()
	}
	if req.Data.Reason != nil {
		update["reason"] = req.Data.GetReason()
	}
	if req.Data.Type != nil {
		update["type"] = adminLoginRestrictionTypeToString(req.Data.Type)
	}
	if req.Data.Method != nil {
		update["method"] = adminLoginRestrictionMethodToString(req.Data.Method)
	}
	if req.Data.UpdatedBy != nil {
		update["update_by"] = req.Data.GetUpdatedBy()
	}
	now := time.Now()
	update["updated_at"] = now

	_, err := r.q.AdminLoginRestriction.WithContext(ctx).
		Where(r.q.AdminLoginRestriction.ID.Eq(int32(req.Data.GetId()))).
		Updates(update)
	if err != nil {
		r.log.Errorf("update data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *AdminLoginRestrictionRepo) Delete(ctx context.Context, req *adminV1.DeleteAdminLoginRestrictionRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	_, err := r.q.AdminLoginRestriction.WithContext(ctx).
		Where(r.q.AdminLoginRestriction.ID.Eq(int32(req.GetId()))).
		Delete()
	if err != nil {
		r.log.Errorf("delete data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete data failed")
	}
	return nil
}

func (r *AdminLoginRestrictionRepo) toDTO(entity *model.AdminLoginRestriction) *adminV1.AdminLoginRestriction {
	if entity == nil {
		return nil
	}

	dto := &adminV1.AdminLoginRestriction{
		Id:        datautil.CloneUint32(uint32(entity.ID)),
		TargetId:  datautil.CloneUint32(toUint32(entity.TargetID)),
		Value:     cloneStringPtr(entity.Value),
		Reason:    cloneStringPtr(entity.Reason),
		Type:      stringToAdminLoginRestrictionType(entity.Type),
		Method:    stringToAdminLoginRestrictionMethod(entity.Method),
		CreatedBy: datautil.CloneUint32(toUint32(entity.CreateBy)),
		UpdatedBy: datautil.CloneUint32(toUint32(entity.UpdateBy)),
		CreatedAt: timeutil.TimeToTimestamppb(entity.CreatedAt),
		UpdatedAt: timeutil.TimeToTimestamppb(entity.UpdatedAt),
		DeletedAt: timeutil.TimeToTimestamppb(entity.DeletedAt),
	}

	return dto
}

func cloneInt64FromUint32(v *uint32) *int64 {
	if v == nil {
		return nil
	}
	val := int64(*v)
	return &val
}

func cloneStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	val := *s
	return &val
}

func adminLoginRestrictionTypeToString(tp *adminV1.AdminLoginRestriction_Type) *string {
	if tp == nil {
		return nil
	}
	s := tp.String()
	return &s
}

func adminLoginRestrictionMethodToString(m *adminV1.AdminLoginRestriction_Method) *string {
	if m == nil {
		return nil
	}
	s := m.String()
	return &s
}

func stringToAdminLoginRestrictionType(s *string) *adminV1.AdminLoginRestriction_Type {
	if s == nil {
		return nil
	}
	if val, ok := adminV1.AdminLoginRestriction_Type_value[strings.ToUpper(*s)]; ok {
		enum := adminV1.AdminLoginRestriction_Type(val)
		return &enum
	}
	return nil
}

func stringToAdminLoginRestrictionMethod(s *string) *adminV1.AdminLoginRestriction_Method {
	if s == nil {
		return nil
	}
	if val, ok := adminV1.AdminLoginRestriction_Method_value[strings.ToUpper(*s)]; ok {
		enum := adminV1.AdminLoginRestriction_Method(val)
		return &enum
	}
	return nil
}

func toUint32(v *int64) uint32 {
	if v == nil {
		return 0
	}
	return uint32(*v)
}
