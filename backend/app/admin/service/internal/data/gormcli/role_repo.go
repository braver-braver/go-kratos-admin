package gormcli

import (
	"context"
	"errors"
	entityhelper "kratos-admin/pkg/utils/entity"
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

type RoleRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewRoleRepo(db *gorm.DB, logger log.Logger) *RoleRepo {
	return &RoleRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "role/gormcli")),
	}
}

func (r *RoleRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysRole](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, userV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *RoleRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListRoleResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderRole(req.GetOrderBy()),
		scopeFieldMaskRole(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysRole](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
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

	dtos := make([]*userV1.Role, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toRoleDTO(&entities[i]))
	}

	return &userV1.ListRoleResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *RoleRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysRole](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *RoleRepo) Get(ctx context.Context, req *userV1.GetRoleRequest) (*userV1.Role, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysRole](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskRole(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userV1.ErrorRoleNotFound("role not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}
	return toRoleDTO(&entity), nil
}

func (r *RoleRepo) GetRoleByCode(ctx context.Context, code string) (*userV1.Role, error) {
	if code == "" {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysRole](r.db.WithContext(ctx)).
		Where("code = ?", code).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userV1.ErrorRoleNotFound("role not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}
	return toRoleDTO(&entity), nil
}

func (r *RoleRepo) GetRolesByRoleCodes(ctx context.Context, codes []string) ([]*userV1.Role, error) {
	if len(codes) == 0 {
		return []*userV1.Role{}, nil
	}
	entities, err := gorm.G[model.SysRole](r.db.WithContext(ctx)).
		Where("code IN ?", codes).
		Find(ctx)
	if err != nil {
		r.log.Errorf("query roles by codes failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query roles by codes failed")
	}
	dtos := make([]*userV1.Role, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toRoleDTO(&entities[i]))
	}
	return dtos, nil
}

func (r *RoleRepo) GetRolesByRoleIds(ctx context.Context, ids []uint32) ([]*userV1.Role, error) {
	if len(ids) == 0 {
		return []*userV1.Role{}, nil
	}
	id64 := make([]int64, 0, len(ids))
	for _, v := range ids {
		id64 = append(id64, int64(v))
	}
	entities, err := gorm.G[model.SysRole](r.db.WithContext(ctx)).
		Where("id IN ?", id64).
		Find(ctx)
	if err != nil {
		r.log.Errorf("query roles by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query roles by ids failed")
	}
	dtos := make([]*userV1.Role, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toRoleDTO(&entities[i]))
	}
	return dtos, nil
}

// GetRoleCodesByRoleIds returns role codes for given role IDs.
func (r *RoleRepo) GetRoleCodesByRoleIds(ctx context.Context, ids []uint32) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}

	var codes []string
	if err := r.db.WithContext(ctx).
		Table(model.TableNameSysRole).
		Select("code").
		Where("id IN ?", ids).
		Pluck("code", &codes).Error; err != nil {
		r.log.Errorf("query role codes by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query role codes by ids failed")
	}

	return codes, nil
}

func (r *RoleRepo) Create(ctx context.Context, req *userV1.CreateRoleRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	entity := toRoleModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.SysRole](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *RoleRepo) Update(ctx context.Context, req *userV1.UpdateRoleRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &userV1.CreateRoleRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, path := range req.GetUpdateMask().GetPaths() {
		switch strings.ToLower(path) {
		case "name":
			updates["name"] = req.Data.GetName()
		case "parent_id":
			updates["parent_id"] = req.Data.GetParentId()
		case "sort_order":
			updates["sort_order"] = req.Data.GetSortOrder()
		case "code":
			updates["code"] = req.Data.GetCode()
		case "remark":
			updates["remark"] = req.Data.GetRemark()
		case "status":
			updates["status"] = req.Data.GetStatus().String()
		case "updated_by":
			updates["updated_by"] = req.Data.GetUpdatedBy()
		case "updated_at":
			updates["updated_at"] = toTime(req.Data.GetUpdatedAt())
		case "menus":
			updates["menus"] = req.Data.GetMenus()
		case "apis":
			updates["apis"] = req.Data.GetApis()
		}
	}

	if len(updates) == 0 {
		return nil
	}
	if _, ok := updates["updated_at"]; !ok {
		updates["updated_at"] = time.Now()
	}

	db := r.db.WithContext(ctx).Model(&model.SysRole{}).Where("id = ?", req.Data.GetId())
	if err := db.Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *RoleRepo) Delete(ctx context.Context, req *userV1.DeleteRoleRequest) error {
	if req == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}
	// cascade children by parent_id path
	var ids []int64
	id := int64(req.GetId())
	ids = append(ids, id)
	var children []int64
	if err := r.db.WithContext(ctx).Raw(
		"WITH RECURSIVE cte AS (SELECT id, parent_id FROM sys_roles WHERE id = ? UNION ALL SELECT s.id, s.parent_id FROM sys_roles s INNER JOIN cte ON s.parent_id = cte.id) SELECT id FROM cte",
		id,
	).Scan(&children).Error; err == nil {
		ids = append(ids, children...)
	}
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.SysRole{}).Error; err != nil {
		r.log.Errorf("delete roles failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete roles failed")
	}
	return nil
}

func toRoleDTO(entity *model.SysRole) *userV1.Role {
	dto := &userV1.Role{}
	if entity == nil {
		return dto
	}
	id := uint32(entity.ID)
	dto.Id = &id
	dto.Name = &entity.Name
	dto.ParentId = toUint32Ptr(entity.ParentID)
	dto.SortOrder = &entity.SortOrder
	dto.Code = &entity.Code
	dto.Remark = &entity.Remark
	if entity.Status != "" {
		if v, ok := userV1.Role_Status_value[entity.Status]; ok {
			status := userV1.Role_Status(v)
			dto.Status = &status
		}
	}
	dto.Menus, _ = entityhelper.ParseUint32SliceFromJSONArrayString(entity.Menus)
	dto.Apis, _ = entityhelper.ParseUint32SliceFromJSONArrayString(entity.Apis)

	dto.CreatedBy = toUint32Ptr(entity.CreatedBy)
	dto.UpdatedBy = toUint32Ptr(entity.UpdatedBy)
	if !entity.CreatedAt.IsZero() {
		dto.CreatedAt = timestamppb.New(entity.CreatedAt)
	}
	if !entity.UpdatedAt.IsZero() {
		dto.UpdatedAt = timestamppb.New(entity.UpdatedAt)
	}
	return dto
}

func toRoleModel(dto *userV1.Role) *model.SysRole {
	if dto == nil {
		return nil
	}
	entity := &model.SysRole{
		ID:        int64(dto.GetId()),
		Name:      dto.GetName(),
		ParentID:  int64(dto.GetParentId()),
		SortOrder: dto.GetSortOrder(),
		Code:      dto.GetCode(),
		Remark:    dto.GetRemark(),
		Status:    dto.GetStatus().String(),
		CreatedBy: int64(dto.GetCreatedBy()),
		UpdatedBy: int64(dto.GetUpdatedBy()),
	}
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.CreatedAt.AsTime()
	}
	if dto.UpdatedAt != nil {
		entity.UpdatedAt = dto.UpdatedAt.AsTime()
	}
	return entity
}

var roleColumns = map[string]string{
	"id":         "id",
	"created_at": "created_at",
	"updated_at": "updated_at",
	"name":       "name",
	"parent_id":  "parent_id",
	"sort_order": "sort_order",
	"code":       "code",
	"remark":     "remark",
	"status":     "status",
	"menus":      "menus",
	"apis":       "apis",
	"created_by": "created_by",
	"updated_by": "updated_by",
	"tenant_id":  "tenant_id",
	"deleted_at": "deleted_at",
	"deleted_by": "deleted_by",
	"data_scope": "data_scope",
}

func scopeOrderRole(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := roleColumns[key]; ok {
				cols = append(
					cols, clause.OrderByColumn{
						Column: clause.Column{Name: col},
						Desc:   desc,
					},
				)
			}
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskRole(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := roleColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}
