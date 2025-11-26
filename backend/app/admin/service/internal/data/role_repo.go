package data

import (
	"context"
	"fmt"
	"strings"
	"time"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/model"
	"kratos-admin/app/admin/service/internal/data/gorm/query"
	datautil "kratos-admin/pkg/datautil"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
	"github.com/tx7do/go-utils/fieldmaskutil"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

type RoleRepo struct {
	data *Data
	log  *log.Helper
	db   *gorm.DB
	q    *query.Query
}

func NewRoleRepo(data *Data, logger log.Logger) *RoleRepo {
	if data == nil {
		panic("data must not be nil")
	}

	if data.db == nil {
		panic("gorm db must not be nil")
	}

	repo := &RoleRepo{
		log:  log.NewHelper(log.With(logger, "module", "role/repo/admin-service")),
		data: data,
		db:   data.db,
		q:    query.Use(data.db),
	}

	return repo
}

func (r *RoleRepo) Count(ctx context.Context, conds ...gen.Condition) (int, error) {
	builder := r.q.SysRole.WithContext(ctx)
	if len(conds) > 0 {
		builder = builder.Where(conds...)
	}

	count, err := builder.Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, userV1.ErrorInternalServerError("query count failed")
	}

	return int(count), nil
}

func (r *RoleRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListRoleResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	countQuery, err := r.buildFilteredQuery(ctx, req)
	if err != nil {
		return nil, err
	}

	total, err := countQuery.Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query count failed")
	}

	dataQuery, err := r.buildFilteredQuery(ctx, req)
	if err != nil {
		return nil, err
	}

	dataQuery = r.applyOrder(dataQuery, req)

	if !req.GetNoPaging() {
		pageSize := int(req.GetPageSize())
		if pageSize <= 0 {
			pageSize = 10
		}
		offset := int(req.GetPage()-1) * pageSize
		if offset < 0 {
			offset = 0
		}
		dataQuery = dataQuery.Offset(offset).Limit(pageSize)
	}

	if req.GetFieldMask() != nil && len(req.GetFieldMask().GetPaths()) > 0 {
		fields := r.toColumnExprs(req.GetFieldMask().GetPaths())
		if len(fields) > 0 {
			dataQuery = dataQuery.Select(fields...)
		}
	}

	entities, err := dataQuery.Find()
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query list failed")
	}

	dtos := make([]*userV1.Role, 0, len(entities))
	for _, entity := range entities {
		dto, convErr := r.toDTO(entity)
		if convErr != nil {
			r.log.Errorf("convert entity to dto failed: %s", convErr.Error())
			return nil, userV1.ErrorInternalServerError("convert data failed")
		}
		dtos = append(dtos, dto)
	}

	return &userV1.ListRoleResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *RoleRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	if id == 0 {
		return false, nil
	}

	builder := r.q.SysRole.WithContext(ctx).Where(query.SysRole.ID.Eq(int32(id)))

	count, err := builder.Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}

	return count > 0, nil
}

func (r *RoleRepo) Get(ctx context.Context, id uint32) (*userV1.Role, error) {
	if id == 0 {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.q.SysRole.WithContext(ctx).Where(query.SysRole.ID.Eq(int32(id))).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorRoleNotFound("role not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(entity)
}

func (r *RoleRepo) GetRoleByCode(ctx context.Context, code string) (*userV1.Role, error) {
	if code == "" {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.q.SysRole.WithContext(ctx).Where(query.SysRole.Code.Eq(code)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorRoleNotFound("role not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(entity)
}

func (r *RoleRepo) GetRolesByRoleCodes(ctx context.Context, codes []string) ([]*userV1.Role, error) {
	if len(codes) == 0 {
		return []*userV1.Role{}, nil
	}

	entities, err := r.q.SysRole.WithContext(ctx).Where(query.SysRole.Code.In(codes...)).Find()
	if err != nil {
		r.log.Errorf("query roles by codes failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query roles by codes failed")
	}

	return r.entitiesToDTOs(entities)
}

func (r *RoleRepo) GetRolesByRoleIds(ctx context.Context, ids []uint32) ([]*userV1.Role, error) {
	if len(ids) == 0 {
		return []*userV1.Role{}, nil
	}

	intIDs := make([]int32, 0, len(ids))
	for _, id := range ids {
		intIDs = append(intIDs, int32(id))
	}

	entities, err := r.q.SysRole.WithContext(ctx).Where(query.SysRole.ID.In(intIDs...)).Find()
	if err != nil {
		r.log.Errorf("query roles by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query roles by ids failed")
	}

	return r.entitiesToDTOs(entities)
}

func (r *RoleRepo) GetRoleCodesByRoleIds(ctx context.Context, ids []uint32) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}

	intIDs := make([]int32, 0, len(ids))
	for _, id := range ids {
		intIDs = append(intIDs, int32(id))
	}

	entities, err := r.q.SysRole.WithContext(ctx).
		Select(query.SysRole.Code).
		Where(query.SysRole.ID.In(intIDs...)).
		Find()
	if err != nil {
		r.log.Errorf("query role codes failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query role codes failed")
	}

	codes := make([]string, 0, len(entities))
	for _, entity := range entities {
		if entity.Code != nil {
			codes = append(codes, *entity.Code)
		}
	}

	return codes, nil
}

func (r *RoleRepo) Create(ctx context.Context, req *userV1.CreateRoleRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.fromCreateRequest(req.Data)
	if err != nil {
		r.log.Errorf("build entity from request failed: %s", err.Error())
		return userV1.ErrorBadRequest("invalid parameter")
	}

	if err = r.q.SysRole.WithContext(ctx).Create(entity); err != nil {
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

	if req.UpdateMask != nil {
		req.UpdateMask.Normalize()
		if !req.UpdateMask.IsValid(req.Data) {
			r.log.Errorf("invalid field mask [%v]", req.UpdateMask)
			return userV1.ErrorBadRequest("invalid field mask")
		}
		fieldmaskutil.Filter(req.GetData(), req.UpdateMask.GetPaths())
	}

	updateMap, err := r.buildUpdateMap(req.Data)
	if err != nil {
		r.log.Errorf("build update data failed: %s", err.Error())
		return userV1.ErrorBadRequest("invalid update data")
	}

	if req.UpdateMask != nil {
		nilPaths := fieldmaskutil.NilValuePaths(req.Data, req.GetUpdateMask().GetPaths())
		for _, path := range nilPaths {
			if column := roleFieldToColumn(path); column != "" {
				updateMap[column] = gorm.Expr("NULL")
			}
		}
	}

	if _, ok := updateMap["updated_at"]; !ok {
		updateMap["updated_at"] = time.Now()
	}

	if _, err := r.q.SysRole.WithContext(ctx).Where(query.SysRole.ID.Eq(int32(req.GetData().GetId()))).Updates(updateMap); err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *RoleRepo) Delete(ctx context.Context, req *userV1.DeleteRoleRequest) error {
	if req == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	childIDs, err := r.queryAllChildrenIds(ctx, req.GetId())
	if err != nil {
		r.log.Errorf("query child roles failed: %s", err.Error())
		return userV1.ErrorInternalServerError("query child roles failed")
	}
	childIDs = append(childIDs, req.GetId())

	intIDs := make([]int32, 0, len(childIDs))
	for _, id := range childIDs {
		intIDs = append(intIDs, int32(id))
	}

	if len(intIDs) == 0 {
		return nil
	}

	if _, err = r.q.SysRole.WithContext(ctx).Where(query.SysRole.ID.In(intIDs...)).Delete(); err != nil {
		r.log.Errorf("delete roles failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete roles failed")
	}

	return nil
}

func (r *RoleRepo) buildFilteredQuery(ctx context.Context, req *pagination.PagingRequest) (query.ISysRoleDo, error) {
	builder := r.q.SysRole.WithContext(ctx)

	var err error
	builder, err = r.applyFilter(builder, req.GetQuery(), false)
	if err != nil {
		r.log.Errorf("apply filter failed: %s", err.Error())
		return nil, userV1.ErrorBadRequest("invalid query parameter")
	}

	builder, err = r.applyFilter(builder, req.GetOrQuery(), true)
	if err != nil {
		r.log.Errorf("apply OR filter failed: %s", err.Error())
		return nil, userV1.ErrorBadRequest("invalid query parameter")
	}

	return builder, nil
}

func (r *RoleRepo) applyFilter(builder query.ISysRoleDo, filterJSON string, isOr bool) (query.ISysRoleDo, error) {
	if filterJSON == "" {
		return builder, nil
	}

	maps, err := datautil.ParseFilterJSON(filterJSON)
	if err != nil {
		return nil, err
	}

	for _, m := range maps {
		for key, value := range m {
			cond, convErr := buildRoleCondition(key, value)
			if convErr != nil {
				return nil, convErr
			}
			if cond == nil {
				continue
			}
			if isOr {
				builder = builder.Or(cond)
			} else {
				builder = builder.Where(cond)
			}
		}
	}

	return builder, nil
}

func (r *RoleRepo) applyOrder(builder query.ISysRoleDo, req *pagination.PagingRequest) query.ISysRoleDo {
	orderBys := req.GetOrderBy()
	if len(orderBys) == 0 {
		return builder.Order(query.SysRole.CreatedAt.Desc())
	}

	for _, item := range orderBys {
		field := item
		desc := false
		if strings.HasPrefix(item, "-") {
			desc = true
			field = item[1:]
		}
		if expr, ok := query.SysRole.GetFieldByName(field); ok {
			if desc {
				builder = builder.Order(expr.Desc())
			} else {
				builder = builder.Order(expr.Asc())
			}
		}
	}

	return builder
}

func (r *RoleRepo) toDTO(entity *model.SysRole) (*userV1.Role, error) {
	if entity == nil {
		return nil, nil
	}

	dto := &userV1.Role{}

	if err := copier.CopyWithOption(dto, entity, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
		FieldNameMapping: []copier.FieldNameMapping{
			{
				SrcType: model.SysRole{},
				DstType: userV1.Role{},
				Mapping: map[string]string{
					"SortID": "SortOrder",
				},
			},
		},
	}); err != nil {
		return nil, err
	}

	id := uint32(entity.ID)
	dto.Id = &id

	if entity.ParentID != nil {
		val := uint32(*entity.ParentID)
		dto.ParentId = &val
	}
	if entity.TenantID != nil {
		val := uint32(*entity.TenantID)
		dto.TenantId = &val
	}

	if entity.CreateBy != nil {
		val := uint32(*entity.CreateBy)
		dto.CreatedBy = &val
	}
	if entity.UpdateBy != nil {
		val := uint32(*entity.UpdateBy)
		dto.UpdatedBy = &val
	}

	if entity.Status != nil {
		if status := statusStringToProto(*entity.Status); status != nil {
			dto.Status = status
		}
	}

	if entity.Menu != nil {
		menus, err := datautil.DecodeUint32Slice(*entity.Menu)
		if err != nil {
			return nil, err
		}
		dto.Menus = menus
	}

	if entity.API != nil {
		apis, err := datautil.DecodeUint32Slice(*entity.API)
		if err != nil {
			return nil, err
		}
		dto.Apis = apis
	}

	dto.CreatedAt = timeutil.TimeToTimestamppb(entity.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(entity.UpdatedAt)
	dto.DeletedAt = timeutil.TimeToTimestamppb(entity.DeletedAt)

	return dto, nil
}

func (r *RoleRepo) entitiesToDTOs(entities []*model.SysRole) ([]*userV1.Role, error) {
	dtos := make([]*userV1.Role, 0, len(entities))
	for _, entity := range entities {
		dto, err := r.toDTO(entity)
		if err != nil {
			return nil, err
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

func (r *RoleRepo) fromCreateRequest(data *userV1.Role) (*model.SysRole, error) {
	now := time.Now()

	entity := &model.SysRole{
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	if data.Id != nil {
		entity.ID = int32(data.GetId())
	}
	if data.Name != nil {
		entity.Name = datautil.CloneString(data.GetName())
	}
	if data.Code != nil {
		entity.Code = datautil.CloneString(data.GetCode())
	}
	if data.SortOrder != nil {
		sortID := data.GetSortOrder()
		entity.SortID = &sortID
	}
	if data.ParentId != nil {
		parent := int32(data.GetParentId())
		entity.ParentID = &parent
	}
	if data.TenantId != nil {
		tenant := int64(data.GetTenantId())
		entity.TenantID = &tenant
	}
	if data.Status != nil {
		if status := statusProtoToString(*data.Status); status != nil {
			entity.Status = status
		}
	}
	if data.Remark != nil {
		entity.Remark = datautil.CloneString(data.GetRemark())
	}
	if data.CreatedBy != nil {
		createdBy := int64(data.GetCreatedBy())
		entity.CreateBy = &createdBy
	}
	if data.UpdatedBy != nil {
		updatedBy := int64(data.GetUpdatedBy())
		entity.UpdateBy = &updatedBy
	}

	if len(data.Menus) > 0 {
		if encoded, err := datautil.EncodeUint32Slice(data.Menus); err == nil {
			entity.Menu = encoded
		} else {
			return nil, err
		}
	}

	if len(data.Apis) > 0 {
		if encoded, err := datautil.EncodeUint32Slice(data.Apis); err == nil {
			entity.API = encoded
		} else {
			return nil, err
		}
	}

	if data.CreatedAt != nil {
		entity.CreatedAt = timeutil.TimestamppbToTime(data.CreatedAt)
	}
	if data.UpdatedAt != nil {
		entity.UpdatedAt = timeutil.TimestamppbToTime(data.UpdatedAt)
	}

	return entity, nil
}

func (r *RoleRepo) buildUpdateMap(data *userV1.Role) (map[string]interface{}, error) {
	updateMap := make(map[string]interface{})

	if data.Name != nil {
		updateMap["name"] = data.GetName()
	}
	if data.SortOrder != nil {
		updateMap["sort_id"] = data.GetSortOrder()
	}
	if data.ParentId != nil {
		updateMap["parent_id"] = int32(data.GetParentId())
	}
	if data.Code != nil {
		updateMap["code"] = data.GetCode()
	}
	if data.Remark != nil {
		updateMap["remark"] = data.GetRemark()
	}
	if data.Status != nil {
		if status := statusProtoToString(*data.Status); status != nil {
			updateMap["status"] = *status
		}
	}
	if data.TenantId != nil {
		updateMap["tenant_id"] = int64(data.GetTenantId())
	}
	if len(data.Menus) > 0 {
		encoded, err := datautil.EncodeUint32Slice(data.Menus)
		if err != nil {
			return nil, err
		}
		updateMap["menus"] = *encoded
	}
	if len(data.Apis) > 0 {
		encoded, err := datautil.EncodeUint32Slice(data.Apis)
		if err != nil {
			return nil, err
		}
		updateMap["apis"] = *encoded
	}
	if data.UpdatedBy != nil {
		updateMap["update_by"] = int64(data.GetUpdatedBy())
	}
	if data.UpdatedAt != nil {
		if ts := timeutil.TimestamppbToTime(data.UpdatedAt); ts != nil {
			updateMap["updated_at"] = *ts
		}
	}

	return updateMap, nil
}

func (r *RoleRepo) toColumnExprs(paths []string) []field.Expr {
	fields := make([]field.Expr, 0, len(paths))
	for _, path := range paths {
		if expr, ok := query.SysRole.GetFieldByName(path); ok {
			fields = append(fields, expr)
		}
	}
	return fields
}

func (r *RoleRepo) queryAllChildrenIds(ctx context.Context, parentID uint32) ([]uint32, error) {
	var sqlQuery string

	switch r.db.Dialector.Name() {
	case "postgres":
		sqlQuery = `
			WITH RECURSIVE all_descendants AS (
				SELECT id, parent_id
				FROM sys_roles
				WHERE parent_id = ?
				UNION ALL
				SELECT r.id, r.parent_id
				FROM sys_roles r
				INNER JOIN all_descendants ad ON r.parent_id = ad.id
			)
			SELECT id FROM all_descendants;
		`
	case "mysql":
		sqlQuery = `
			WITH RECURSIVE all_descendants AS (
				SELECT id, parent_id
				FROM sys_roles
				WHERE parent_id = ?
				UNION ALL
				SELECT r.id, r.parent_id
				FROM sys_roles r
				INNER JOIN all_descendants ad ON r.parent_id = ad.id
			)
			SELECT id FROM all_descendants;
		`
	case "sqlite":
		sqlQuery = `
			WITH RECURSIVE all_descendants AS (
				SELECT id, parent_id
				FROM sys_roles
				WHERE parent_id = ?
				UNION ALL
				SELECT r.id, r.parent_id
				FROM sys_roles r
				INNER JOIN all_descendants ad ON r.parent_id = ad.id
			)
			SELECT id FROM all_descendants;
		`
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", r.db.Dialector.Name())
	}

	rows, err := r.db.WithContext(ctx).Raw(sqlQuery, parentID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]uint32, 0)
	for rows.Next() {
		var id uint32
		if scanErr := rows.Scan(&id); scanErr != nil {
			return nil, scanErr
		}
		results = append(results, id)
	}

	return results, nil
}

func buildRoleCondition(key string, value interface{}) (gen.Condition, error) {
	fieldName, operator := datautil.ParseFilterKey(key)

	switch fieldName {
	case "id":
		switch operator {
		case "in":
			ints, err := datautil.ToInt32Slice(value)
			if err != nil {
				return nil, err
			}
			return query.SysRole.ID.In(ints...), nil
		default:
			intVal, err := datautil.ToInt32(value)
			if err != nil {
				return nil, err
			}
			return query.SysRole.ID.Eq(intVal), nil
		}
	case "code":
		strVal := fmt.Sprint(value)
		if operator == "contains" || operator == "icontains" {
			return query.SysRole.Code.Like("%" + strVal + "%"), nil
		}
		return query.SysRole.Code.Eq(strVal), nil
	case "name":
		strVal := fmt.Sprint(value)
		if operator == "contains" || operator == "icontains" {
			return query.SysRole.Name.Like("%" + strVal + "%"), nil
		}
		return query.SysRole.Name.Eq(strVal), nil
	case "status":
		strVal := fmt.Sprint(value)
		return query.SysRole.Status.Eq(strVal), nil
	case "tenant_id":
		intVal, err := datautil.ToInt64(value)
		if err != nil {
			return nil, err
		}
		return query.SysRole.TenantID.Eq(intVal), nil
	case "parent_id":
		intVal, err := datautil.ToInt32(value)
		if err != nil {
			return nil, err
		}
		return query.SysRole.ParentID.Eq(intVal), nil
	default:
		return nil, nil
	}
}

func statusProtoToString(status userV1.Role_Status) *string {
	name := userV1.Role_Status_name[int32(status)]
	if name == "" {
		return nil
	}
	return &name
}

func statusStringToProto(status string) *userV1.Role_Status {
	if value, ok := userV1.Role_Status_value[strings.ToUpper(status)]; ok {
		enum := userV1.Role_Status(value)
		return &enum
	}
	return nil
}

func roleFieldToColumn(field string) string {
	switch field {
	case "name":
		return "name"
	case "code":
		return "code"
	case "status":
		return "status"
	case "remark":
		return "remark"
	case "tenant_id":
		return "tenant_id"
	case "parent_id":
		return "parent_id"
	case "sort_order":
		return "sort_id"
	case "menus":
		return "menus"
	case "apis":
		return "apis"
	case "updated_by":
		return "update_by"
	case "updated_at":
		return "updated_at"
	default:
		return ""
	}
}
