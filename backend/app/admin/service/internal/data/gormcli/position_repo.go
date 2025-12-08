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

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type PositionRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewPositionRepo(db *gorm.DB, logger log.Logger) *PositionRepo {
	return &PositionRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "position/gormcli")),
	}
}

func (r *PositionRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysPosition](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, userV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *PositionRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListPositionResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderPosition(req.GetOrderBy()),
		scopeFieldMaskPosition(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysPosition](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
	total, err := r.Count(ctx, fScopes...)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query count failed")
	}

	entities, err := g.Find(ctx)
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query list failed")
	}

	dtos := make([]*userV1.Position, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toPositionDTO(&entities[i]))
	}

	return &userV1.ListPositionResponse{Total: uint32(total), Items: dtos}, nil
}

func (r *PositionRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysPosition](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *PositionRepo) Get(ctx context.Context, req *userV1.GetPositionRequest) (*userV1.Position, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.SysPosition](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskPosition(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userV1.ErrorPositionNotFound("position not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return toPositionDTO(&entity), nil
}

func (r *PositionRepo) GetPositionsByIds(ctx context.Context, ids []uint32) ([]*userV1.Position, error) {
	if len(ids) == 0 {
		return []*userV1.Position{}, nil
	}

	entities, err := gorm.G[model.SysPosition](r.db.WithContext(ctx)).
		Where("id IN ?", ids).
		Find(ctx)
	if err != nil {
		r.log.Errorf("query position by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query position by ids failed")
	}

	dtos := make([]*userV1.Position, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toPositionDTO(&entities[i]))
	}
	return dtos, nil
}

func (r *PositionRepo) Create(ctx context.Context, req *userV1.CreatePositionRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	entity := toPositionModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.SysPosition](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *PositionRepo) Update(ctx context.Context, req *userV1.UpdatePositionRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &userV1.CreatePositionRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch p {
		case "name":
			updates["name"] = req.Data.GetName()
		case "description":
			updates["description"] = req.Data.GetDescription()
		case "organization_id":
			updates["organization_id"] = req.Data.GetOrganizationId()
		case "department_id":
			updates["department_id"] = req.Data.GetDepartmentId()
		case "parent_id":
			updates["parent_id"] = req.Data.GetParentId()
		case "sort_order":
			updates["sort_order"] = req.Data.GetSortOrder()
		case "status":
			updates["status"] = req.Data.GetStatus().String()
		case "remark":
			updates["remark"] = req.Data.GetRemark()
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

	if err := r.db.WithContext(ctx).Model(&model.SysPosition{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *PositionRepo) Delete(ctx context.Context, req *userV1.DeletePositionRequest) error {
	if req == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).Where("id = ?", req.GetId()).Delete(&model.SysPosition{}).Error; err != nil {
		r.log.Errorf("delete position failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

var positionColumns = map[string]string{
	"id":              "id",
	"name":            "name",
	"description":     "description",
	"organization_id": "organization_id",
	"department_id":   "department_id",
	"parent_id":       "parent_id",
	"sort_order":      "sort_order",
	"status":          "status",
	"remark":          "remark",
	"created_by":      "created_by",
	"updated_by":      "updated_by",
	"deleted_by":      "deleted_by",
	"created_at":      "created_at",
	"updated_at":      "updated_at",
	"deleted_at":      "deleted_at",
}

func scopeOrderPosition(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := positionColumns[key]; ok {
				cols = append(cols, clause.OrderByColumn{Column: clause.Column{Name: col}, Desc: desc})
			}
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskPosition(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := positionColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func toPositionDTO(entity *model.SysPosition) *userV1.Position {
	dto := &userV1.Position{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	dto.Name = &entity.Name
	dto.Description = &entity.Description
	dto.OrganizationId = toUint32Ptr(entity.OrganizationID)
	dto.DepartmentId = toUint32Ptr(entity.DepartmentID)
	dto.ParentId = toUint32Ptr(entity.ParentID)
	dto.SortOrder = &entity.SortOrder
	if entity.Status != "" {
		if v, ok := userV1.Position_Status_value[entity.Status]; ok {
			st := userV1.Position_Status(v)
			dto.Status = &st
		}
	}
	dto.Remark = &entity.Remark
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

func toPositionModel(dto *userV1.Position) *model.SysPosition {
	if dto == nil {
		return nil
	}
	entity := &model.SysPosition{
		ID:             int64(dto.GetId()),
		Name:           dto.GetName(),
		Description:    dto.GetDescription(),
		OrganizationID: int64(dto.GetOrganizationId()),
		DepartmentID:   int64(dto.GetDepartmentId()),
		ParentID:       int64(dto.GetParentId()),
		SortOrder:      dto.GetSortOrder(),
		Status:         dto.GetStatus().String(),
		Remark:         dto.GetRemark(),
		CreatedBy:      int64(dto.GetCreatedBy()),
		UpdatedBy:      int64(dto.GetUpdatedBy()),
		DeletedBy:      int64(dto.GetDeletedBy()),
	}
	if dto.Status != nil {
		entity.Status = dto.GetStatus().String()
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
