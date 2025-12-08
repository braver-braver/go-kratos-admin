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

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type DepartmentRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewDepartmentRepo(db *gorm.DB, logger log.Logger) *DepartmentRepo {
	return &DepartmentRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "department/gormcli")),
	}
}

func (r *DepartmentRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysDepartment](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, userV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *DepartmentRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListDepartmentResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderDepartment(req.GetOrderBy()),
		scopeFieldMaskDepartment(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysDepartment](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
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

	// stable sort by sort_order like ent implementation
	sort.SliceStable(entities, func(i, j int) bool {
		return entities[i].SortOrder < entities[j].SortOrder
	})

	dtos := make([]*userV1.Department, 0, len(entities))
	for i := range entities {
		// build tree
		if entities[i].ParentID == 0 {
			dto := toDepartmentDTO(&entities[i])
			dtos = append(dtos, dto)
		}
	}
	for i := range entities {
		if entities[i].ParentID != 0 {
			dto := toDepartmentDTO(&entities[i])
			if addDepartmentChild(dtos, dto) {
				continue
			}
			dtos = append(dtos, dto)
		}
	}

	return &userV1.ListDepartmentResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *DepartmentRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysDepartment](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *DepartmentRepo) Get(ctx context.Context, req *userV1.GetDepartmentRequest) (*userV1.Department, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.SysDepartment](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskDepartment(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userV1.ErrorDepartmentNotFound("department not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return toDepartmentDTO(&entity), nil
}

func (r *DepartmentRepo) GetDepartmentsByIds(ctx context.Context, ids []uint32) ([]*userV1.Department, error) {
	if len(ids) == 0 {
		return []*userV1.Department{}, nil
	}

	entities, err := gorm.G[model.SysDepartment](r.db.WithContext(ctx)).
		Where("id IN ?", ids).
		Find(ctx)
	if err != nil {
		r.log.Errorf("query department by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query department by ids failed")
	}

	dtos := make([]*userV1.Department, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toDepartmentDTO(&entities[i]))
	}
	return dtos, nil
}

func (r *DepartmentRepo) Create(ctx context.Context, req *userV1.CreateDepartmentRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	entity := toDepartmentModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.SysDepartment](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *DepartmentRepo) Update(ctx context.Context, req *userV1.UpdateDepartmentRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &userV1.CreateDepartmentRequest{Data: req.Data}
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
		case "manager_id":
			updates["manager_id"] = req.Data.GetManagerId()
		case "tenant_id":
			updates["tenant_id"] = req.Data.GetTenantId()
		case "sort_order":
			updates["sort_order"] = req.Data.GetSortOrder()
		case "status":
			updates["status"] = req.Data.GetStatus().String()
		case "remark":
			updates["remark"] = req.Data.GetRemark()
		case "parent_id":
			updates["parent_id"] = req.Data.GetParentId()
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

	if err := r.db.WithContext(ctx).Model(&model.SysDepartment{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *DepartmentRepo) Delete(ctx context.Context, req *userV1.DeleteDepartmentRequest) error {
	if req == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).Where("id = ?", req.GetId()).Delete(&model.SysDepartment{}).Error; err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

var departmentColumns = map[string]string{
	"id":              "id",
	"name":            "name",
	"description":     "description",
	"organization_id": "organization_id",
	"manager_id":      "manager_id",
	"tenant_id":       "tenant_id",
	"sort_order":      "sort_order",
	"status":          "status",
	"remark":          "remark",
	"parent_id":       "parent_id",
	"created_by":      "created_by",
	"updated_by":      "updated_by",
	"deleted_by":      "deleted_by",
	"created_at":      "created_at",
	"updated_at":      "updated_at",
	"deleted_at":      "deleted_at",
}

func scopeOrderDepartment(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := departmentColumns[key]; ok {
				cols = append(cols, clause.OrderByColumn{
					Column: clause.Column{Name: col},
					Desc:   desc,
				})
			}
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskDepartment(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := departmentColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func toDepartmentDTO(entity *model.SysDepartment) *userV1.Department {
	dto := &userV1.Department{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	dto.Name = &entity.Name
	dto.Description = &entity.Description
	dto.OrganizationId = toUint32Ptr(entity.OrganizationID)
	dto.ManagerId = toUint32Ptr(entity.ManagerID)
	dto.TenantId = toUint32Ptr(entity.TenantID)
	dto.SortOrder = &entity.SortOrder
	if entity.Status != "" {
		if v, ok := userV1.Department_Status_value[entity.Status]; ok {
			st := userV1.Department_Status(v)
			dto.Status = &st
		}
	}
	dto.Remark = &entity.Remark
	dto.ParentId = toUint32Ptr(entity.ParentID)
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

func toDepartmentModel(dto *userV1.Department) *model.SysDepartment {
	if dto == nil {
		return nil
	}
	entity := &model.SysDepartment{
		ID:             int64(dto.GetId()),
		Name:           dto.GetName(),
		Description:    dto.GetDescription(),
		OrganizationID: int64(dto.GetOrganizationId()),
		ManagerID:      int64(dto.GetManagerId()),
		TenantID:       int64(dto.GetTenantId()),
		SortOrder:      dto.GetSortOrder(),
		Status:         dto.GetStatus().String(),
		Remark:         dto.GetRemark(),
		ParentID:       int64(dto.GetParentId()),
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

func addDepartmentChild(nodes []*userV1.Department, node *userV1.Department) bool {
	if nodes == nil {
		return false
	}
	if node.ParentId == nil || node.GetParentId() == 0 {
		nodes = append(nodes, node)
		return true
	}
	for _, n := range nodes {
		if n.GetId() == node.GetParentId() {
			n.Children = append(n.Children, node)
			return true
		}
		if addDepartmentChild(n.Children, node) {
			return true
		}
	}
	return false
}
