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

type OrganizationRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewOrganizationRepo(db *gorm.DB, logger log.Logger) *OrganizationRepo {
	return &OrganizationRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "organization/gormcli")),
	}
}

func (r *OrganizationRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysOrganization](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, userV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *OrganizationRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListOrganizationResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderOrganization(req.GetOrderBy()),
		scopeFieldMaskOrganization(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysOrganization](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
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

	sort.SliceStable(entities, func(i, j int) bool {
		return entities[i].SortOrder < entities[j].SortOrder
	})

	dtos := make([]*userV1.Organization, 0, len(entities))
	for i := range entities {
		if entities[i].ParentID == 0 {
			dtos = append(dtos, toOrganizationDTO(&entities[i]))
		}
	}
	for i := range entities {
		if entities[i].ParentID != 0 {
			dto := toOrganizationDTO(&entities[i])
			if addOrganizationChild(dtos, dto) {
				continue
			}
			dtos = append(dtos, dto)
		}
	}

	return &userV1.ListOrganizationResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *OrganizationRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysOrganization](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *OrganizationRepo) Get(ctx context.Context, req *userV1.GetOrganizationRequest) (*userV1.Organization, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.SysOrganization](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskOrganization(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userV1.ErrorOrganizationNotFound("organization not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return toOrganizationDTO(&entity), nil
}

func (r *OrganizationRepo) GetOrganizationsByIds(ctx context.Context, ids []uint32) ([]*userV1.Organization, error) {
	if len(ids) == 0 {
		return []*userV1.Organization{}, nil
	}

	entities, err := gorm.G[model.SysOrganization](r.db.WithContext(ctx)).
		Where("id IN ?", ids).
		Find(ctx)
	if err != nil {
		r.log.Errorf("query organization by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query organization by ids failed")
	}

	dtos := make([]*userV1.Organization, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toOrganizationDTO(&entities[i]))
	}
	return dtos, nil
}

func (r *OrganizationRepo) Create(ctx context.Context, req *userV1.CreateOrganizationRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	entity := toOrganizationModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.SysOrganization](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *OrganizationRepo) Update(ctx context.Context, req *userV1.UpdateOrganizationRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &userV1.CreateOrganizationRequest{Data: req.Data}
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
		case "organization_type":
			updates["organization_type"] = req.Data.GetOrganizationType().String()
		case "credit_code":
			updates["credit_code"] = req.Data.GetCreditCode()
		case "address":
			updates["address"] = req.Data.GetAddress()
		case "business_scope":
			updates["business_scope"] = req.Data.GetBusinessScope()
		case "is_legal_entity":
			updates["is_legal_entity"] = req.Data.GetIsLegalEntity()
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

	if err := r.db.WithContext(ctx).Model(&model.SysOrganization{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *OrganizationRepo) Delete(ctx context.Context, req *userV1.DeleteOrganizationRequest) error {
	if req == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).Where("id = ?", req.GetId()).Delete(&model.SysOrganization{}).Error; err != nil {
		r.log.Errorf("delete organization failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

var organizationColumns = map[string]string{
	"id":                "id",
	"name":              "name",
	"description":       "description",
	"organization_type": "organization_type",
	"credit_code":       "credit_code",
	"address":           "address",
	"business_scope":    "business_scope",
	"is_legal_entity":   "is_legal_entity",
	"manager_id":        "manager_id",
	"tenant_id":         "tenant_id",
	"sort_order":        "sort_order",
	"status":            "status",
	"remark":            "remark",
	"parent_id":         "parent_id",
	"created_by":        "created_by",
	"updated_by":        "updated_by",
	"deleted_by":        "deleted_by",
	"created_at":        "created_at",
	"updated_at":        "updated_at",
	"deleted_at":        "deleted_at",
}

func scopeOrderOrganization(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := organizationColumns[key]; ok {
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

func scopeFieldMaskOrganization(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := organizationColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func toOrganizationDTO(entity *model.SysOrganization) *userV1.Organization {
	dto := &userV1.Organization{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	dto.Name = &entity.Name
	if entity.OrganizationType != "" {
		if v, ok := userV1.Organization_Type_value[entity.OrganizationType]; ok {
			tp := userV1.Organization_Type(v)
			dto.OrganizationType = &tp
		}
	}
	dto.CreditCode = &entity.CreditCode
	dto.Address = &entity.Address
	dto.BusinessScope = &entity.BusinessScope
	dto.IsLegalEntity = &entity.IsLegalEntity
	dto.ManagerId = toUint32Ptr(entity.ManagerID)
	dto.TenantId = toUint32Ptr(entity.TenantID)
	dto.SortOrder = &entity.SortOrder
	if entity.Status != "" {
		if v, ok := userV1.Organization_Status_value[entity.Status]; ok {
			st := userV1.Organization_Status(v)
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

func toOrganizationModel(dto *userV1.Organization) *model.SysOrganization {
	if dto == nil {
		return nil
	}
	entity := &model.SysOrganization{
		ID:               int64(dto.GetId()),
		Name:             dto.GetName(),
		OrganizationType: dto.GetOrganizationType().String(),
		CreditCode:       dto.GetCreditCode(),
		Address:          dto.GetAddress(),
		BusinessScope:    dto.GetBusinessScope(),
		IsLegalEntity:    dto.GetIsLegalEntity(),
		ManagerID:        int64(dto.GetManagerId()),
		TenantID:         int64(dto.GetTenantId()),
		SortOrder:        dto.GetSortOrder(),
		Status:           dto.GetStatus().String(),
		Remark:           dto.GetRemark(),
		ParentID:         int64(dto.GetParentId()),
		CreatedBy:        int64(dto.GetCreatedBy()),
		UpdatedBy:        int64(dto.GetUpdatedBy()),
		DeletedBy:        int64(dto.GetDeletedBy()),
	}
	if dto.OrganizationType != nil {
		entity.OrganizationType = dto.GetOrganizationType().String()
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

func addOrganizationChild(nodes []*userV1.Organization, node *userV1.Organization) bool {
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
		if addOrganizationChild(n.Children, node) {
			return true
		}
	}
	return false
}
