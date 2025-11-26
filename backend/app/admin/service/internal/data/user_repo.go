package data

import (
	"context"
	"encoding/json"
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

type UserRepo struct {
	data *Data
	log  *log.Helper
	db   *gorm.DB
	q    *query.Query
}

func NewUserRepo(logger log.Logger, data *Data) *UserRepo {
	if data == nil {
		panic("data must not be nil")
	}

	if data.db == nil {
		panic("gorm db must not be nil")
	}

	return &UserRepo{
		log:  log.NewHelper(log.With(logger, "module", "user/repo/admin-service")),
		data: data,
		db:   data.db,
		q:    query.Use(data.db),
	}
}

func (r *UserRepo) Count(ctx context.Context, conds ...gen.Condition) (int, error) {
	builder := r.q.User.WithContext(ctx)
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

func (r *UserRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListUserResponse, error) {
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

	items, err := r.entitiesToDTOs(entities)
	if err != nil {
		r.log.Errorf("convert user entities failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("convert data failed")
	}

	return &userV1.ListUserResponse{
		Total: uint32(total),
		Items: items,
	}, nil
}

func (r *UserRepo) buildFilteredQuery(ctx context.Context, req *pagination.PagingRequest) (query.IUserDo, error) {
	builder := r.q.User.WithContext(ctx)

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

func (r *UserRepo) applyFilter(builder query.IUserDo, filterJSON string, isOr bool) (query.IUserDo, error) {
	if filterJSON == "" {
		return builder, nil
	}

	maps, err := datautil.ParseFilterJSON(filterJSON)
	if err != nil {
		return nil, err
	}

	for _, m := range maps {
		for key, value := range m {
			cond, convErr := buildUserCondition(key, value)
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

func (r *UserRepo) applyOrder(builder query.IUserDo, req *pagination.PagingRequest) query.IUserDo {
	orderBys := req.GetOrderBy()
	if len(orderBys) == 0 {
		return builder.Order(query.User.CreatedAt.Desc())
	}

	for _, item := range orderBys {
		field := item
		desc := false
		if strings.HasPrefix(item, "-") {
			desc = true
			field = item[1:]
		}
		if expr, ok := query.User.GetFieldByName(field); ok {
			if desc {
				builder = builder.Order(expr.Desc())
			} else {
				builder = builder.Order(expr.Asc())
			}
		}
	}

	return builder
}

func (r *UserRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	if id == 0 {
		return false, nil
	}

	count, err := r.q.User.WithContext(ctx).
		Where(query.User.ID.Eq(int32(id))).
		Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}

	return count > 0, nil
}

func (r *UserRepo) Get(ctx context.Context, userId uint32) (*userV1.User, error) {
	if userId == 0 {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.q.User.WithContext(ctx).
		Where(query.User.ID.Eq(int32(userId))).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorUserNotFound("user not found")
		}
		r.log.Errorf("query user failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(entity)
}

func (r *UserRepo) GetUserByUserName(ctx context.Context, userName string) (*userV1.User, error) {
	if userName == "" {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.q.User.WithContext(ctx).
		Where(query.User.Username.Eq(userName)).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorUserNotFound("user not found")
		}
		r.log.Errorf("query user by username failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(entity)
}

func (r *UserRepo) GetUsersByIds(ctx context.Context, ids []uint32) ([]*userV1.User, error) {
	if len(ids) == 0 {
		return []*userV1.User{}, nil
	}

	intIDs := make([]int32, 0, len(ids))
	for _, id := range ids {
		intIDs = append(intIDs, int32(id))
	}

	entities, err := r.q.User.WithContext(ctx).
		Where(query.User.ID.In(intIDs...)).
		Find()
	if err != nil {
		r.log.Errorf("query users by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query users by ids failed")
	}

	return r.entitiesToDTOs(entities)
}

func (r *UserRepo) Create(ctx context.Context, req *userV1.CreateUserRequest) (*userV1.User, error) {
	if req == nil || req.Data == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.fromCreateRequest(req.Data)
	if err != nil {
		r.log.Errorf("build user entity failed: %s", err.Error())
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	if err = r.q.User.WithContext(ctx).Create(entity); err != nil {
		r.log.Errorf("insert user failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(entity)
}

func (r *UserRepo) Update(ctx context.Context, req *userV1.UpdateUserRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &userV1.CreateUserRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			_, err = r.Create(ctx, createReq)
			return err
		}
	}

	if req.UpdateMask != nil {
		for i := 0; i < len(req.UpdateMask.Paths); i++ {
			if req.UpdateMask.Paths[i] == "password" {
				req.UpdateMask.Paths = append(req.UpdateMask.Paths[:i], req.UpdateMask.Paths[i+1:]...)
				i--
			}
		}

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
			if column := userFieldToColumn(path); column != "" {
				updateMap[column] = gorm.Expr("NULL")
			}
		}
	}

	if _, ok := updateMap["updated_at"]; !ok {
		updateMap["updated_at"] = time.Now()
	}

	_, err = r.q.User.WithContext(ctx).
		Where(query.User.ID.Eq(int32(req.GetData().GetId()))).
		Updates(updateMap)
	if err != nil {
		r.log.Errorf("update user failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, userId uint32) error {
	if userId == 0 {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	_, err := r.q.User.WithContext(ctx).
		Where(query.User.ID.Eq(int32(userId))).
		Delete()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return userV1.ErrorNotFound("user not found")
		}
		r.log.Errorf("delete user failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *UserRepo) UserExists(ctx context.Context, req *userV1.UserExistsRequest) (*userV1.UserExistsResponse, error) {
	exist, err := r.q.User.WithContext(ctx).
		Where(query.User.Username.Eq(req.GetUsername())).
		Count()
	if err != nil {
		r.log.Errorf("query user exist failed: %s", err.Error())
		return &userV1.UserExistsResponse{Exist: false}, userV1.ErrorInternalServerError("query exist failed")
	}

	return &userV1.UserExistsResponse{Exist: exist > 0}, nil
}

func (r *UserRepo) toDTO(entity *model.User) (*userV1.User, error) {
	if entity == nil {
		return nil, nil
	}

	dto := &userV1.User{}

	if err := copier.CopyWithOption(
		dto, entity, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
			FieldNameMapping: []copier.FieldNameMapping{
				{
					SrcType: model.User{},
					DstType: userV1.User{},
					Mapping: map[string]string{
						"CreatedAt": "CreatedAt",
						"UpdatedAt": "UpdatedAt",
					},
				},
			},
		},
	); err != nil {
		return nil, err
	}

	id := uint32(entity.ID)
	dto.Id = &id

	if entity.TenantID != nil {
		tenant := uint32(*entity.TenantID)
		dto.TenantId = &tenant
	}
	if entity.OrgID != nil {
		org := uint32(*entity.OrgID)
		dto.OrgId = &org
	}
	if entity.PositionID != nil {
		pos := uint32(*entity.PositionID)
		dto.PositionId = &pos
	}
	if entity.WorkID != nil {
		work := uint32(*entity.WorkID)
		dto.WorkId = &work
	}

	if entity.CreateBy != nil {
		createdBy := uint32(*entity.CreateBy)
		dto.CreatedBy = &createdBy
	}
	if entity.UpdateBy != nil {
		updatedBy := uint32(*entity.UpdateBy)
		dto.UpdatedBy = &updatedBy
	}

	if entity.Status != nil {
		if status := userStatusStringToProto(*entity.Status); status != nil {
			dto.Status = status
		}
	}

	if entity.Gender != nil {
		if gender := userGenderStringToProto(*entity.Gender); gender != nil {
			dto.Gender = gender
		}
	}

	if entity.Authority != nil {
		if authority := userAuthorityStringToProto(*entity.Authority); authority != nil {
			dto.Authority = authority
		}
	}

	if entity.Role != nil {
		roleIDs, err := datautil.DecodeUint32Slice(*entity.Role)
		if err != nil {
			return nil, err
		}
		dto.RoleIds = roleIDs
	}

	dto.CreatedAt = timeutil.TimeToTimestamppb(entity.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(entity.UpdatedAt)
	dto.DeletedAt = timeutil.TimeToTimestamppb(entity.DeletedAt)

	return dto, nil
}

func (r *UserRepo) entitiesToDTOs(entities []*model.User) ([]*userV1.User, error) {
	dtos := make([]*userV1.User, 0, len(entities))
	for _, entity := range entities {
		dto, err := r.toDTO(entity)
		if err != nil {
			return nil, err
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

func (r *UserRepo) fromCreateRequest(data *userV1.User) (*model.User, error) {
	now := time.Now()

	entity := &model.User{
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	if data.Id != nil {
		entity.ID = int32(data.GetId())
	}
	if data.Username != nil {
		entity.Username = datautil.CloneString(data.GetUsername())
	}
	if data.Nickname != nil {
		entity.Nickname = datautil.CloneString(data.GetNickname())
	}
	if data.Realname != nil {
		entity.Realname = datautil.CloneString(data.GetRealname())
	}
	if data.Avatar != nil {
		entity.Avatar = datautil.CloneString(data.GetAvatar())
	}
	if data.Email != nil {
		entity.Email = datautil.CloneString(data.GetEmail())
	}
	if data.Mobile != nil {
		entity.Mobile = datautil.CloneString(data.GetMobile())
	}
	if data.Telephone != nil {
		entity.Telephone = datautil.CloneString(data.GetTelephone())
	}
	if data.Region != nil {
		entity.Region = datautil.CloneString(data.GetRegion())
	}
	if data.Address != nil {
		entity.Address = datautil.CloneString(data.GetAddress())
	}
	if data.Description != nil {
		entity.Description = datautil.CloneString(data.GetDescription())
	}
	if data.Remark != nil {
		entity.Remark = datautil.CloneString(data.GetRemark())
	}

	if data.Status != nil {
		if status := userStatusProtoToString(*data.Status); status != nil {
			entity.Status = status
		}
	}
	if data.Gender != nil {
		if gender := userGenderProtoToString(*data.Gender); gender != nil {
			entity.Gender = gender
		}
	}
	if data.Authority != nil {
		if authority := userAuthorityProtoToString(*data.Authority); authority != nil {
			entity.Authority = authority
		}
	}

	if len(data.RoleIds) > 0 {
		encoded, err := datautil.EncodeUint32Slice(data.RoleIds)
		if err != nil {
			return nil, err
		}
		entity.Role = encoded
	}

	if data.LastLoginTime != nil {
		entity.LastLoginTime = timeutil.TimestamppbToTime(data.LastLoginTime)
	}
	if data.LastLoginIp != nil {
		entity.LastLoginIP = datautil.CloneString(data.GetLastLoginIp())
	}

	if data.CreatedAt != nil {
		entity.CreatedAt = timeutil.TimestamppbToTime(data.CreatedAt)
	}
	if data.UpdatedAt != nil {
		entity.UpdatedAt = timeutil.TimestamppbToTime(data.UpdatedAt)
	}

	if data.CreatedBy != nil {
		createdBy := int64(data.GetCreatedBy())
		entity.CreateBy = &createdBy
	}
	if data.UpdatedBy != nil {
		updatedBy := int64(data.GetUpdatedBy())
		entity.UpdateBy = &updatedBy
	}

	if data.TenantId != nil {
		tenant := int64(data.GetTenantId())
		entity.TenantID = &tenant
	}
	if data.OrgId != nil {
		org := int64(data.GetOrgId())
		entity.OrgID = &org
	}
	if data.PositionId != nil {
		pos := int64(data.GetPositionId())
		entity.PositionID = &pos
	}
	if data.WorkId != nil {
		work := int64(data.GetWorkId())
		entity.WorkID = &work
	}

	return entity, nil
}

func (r *UserRepo) buildUpdateMap(data *userV1.User) (map[string]interface{}, error) {
	updateMap := make(map[string]interface{})

	if data.Username != nil {
		updateMap["username"] = data.GetUsername()
	}
	if data.Nickname != nil {
		updateMap["nickname"] = data.GetNickname()
	}
	if data.Realname != nil {
		updateMap["realname"] = data.GetRealname()
	}
	if data.Avatar != nil {
		updateMap["avatar"] = data.GetAvatar()
	}
	if data.Email != nil {
		updateMap["email"] = data.GetEmail()
	}
	if data.Mobile != nil {
		updateMap["mobile"] = data.GetMobile()
	}
	if data.Telephone != nil {
		updateMap["telephone"] = data.GetTelephone()
	}
	if data.Address != nil {
		updateMap["address"] = data.GetAddress()
	}
	if data.Region != nil {
		updateMap["region"] = data.GetRegion()
	}
	if data.Description != nil {
		updateMap["description"] = data.GetDescription()
	}
	if data.Remark != nil {
		updateMap["remark"] = data.GetRemark()
	}

	if data.Status != nil {
		if status := userStatusProtoToString(*data.Status); status != nil {
			updateMap["status"] = *status
		}
	}
	if data.Gender != nil {
		if gender := userGenderProtoToString(*data.Gender); gender != nil {
			updateMap["gender"] = *gender
		}
	}
	if data.Authority != nil {
		if authority := userAuthorityProtoToString(*data.Authority); authority != nil {
			updateMap["authority"] = *authority
		}
	}

	if data.LastLoginTime != nil {
		if ts := timeutil.TimestamppbToTime(data.LastLoginTime); ts != nil {
			updateMap["last_login_time"] = *ts
		}
	}
	if data.LastLoginIp != nil {
		updateMap["last_login_ip"] = data.GetLastLoginIp()
	}

	if data.UpdatedBy != nil {
		updateMap["update_by"] = int64(data.GetUpdatedBy())
	}
	if data.UpdatedAt != nil {
		if ts := timeutil.TimestamppbToTime(data.UpdatedAt); ts != nil {
			updateMap["updated_at"] = *ts
		}
	}

	if data.TenantId != nil {
		updateMap["tenant_id"] = int64(data.GetTenantId())
	}
	if data.OrgId != nil {
		updateMap["org_id"] = int64(data.GetOrgId())
	}
	if data.PositionId != nil {
		updateMap["position_id"] = int64(data.GetPositionId())
	}
	if data.WorkId != nil {
		updateMap["work_id"] = int64(data.GetWorkId())
	}

	if len(data.RoleIds) > 0 {
		encoded, err := datautil.EncodeUint32Slice(data.RoleIds)
		if err != nil {
			return nil, err
		}
		updateMap["roles"] = *encoded
	}

	return updateMap, nil
}

func (r *UserRepo) toColumnExprs(paths []string) []field.Expr {
	fields := make([]field.Expr, 0, len(paths))
	for _, path := range paths {
		if expr, ok := query.User.GetFieldByName(path); ok {
			fields = append(fields, expr)
		}
	}
	return fields
}

func userFieldToColumn(field string) string {
	switch field {
	case "username":
		return "username"
	case "nickname":
		return "nickname"
	case "realname":
		return "realname"
	case "avatar":
		return "avatar"
	case "email":
		return "email"
	case "mobile":
		return "mobile"
	case "telephone":
		return "telephone"
	case "address":
		return "address"
	case "region":
		return "region"
	case "description":
		return "description"
	case "remark":
		return "remark"
	case "status":
		return "status"
	case "gender":
		return "gender"
	case "authority":
		return "authority"
	case "last_login_time":
		return "last_login_time"
	case "last_login_ip":
		return "last_login_ip"
	case "tenant_id":
		return "tenant_id"
	case "org_id":
		return "org_id"
	case "position_id":
		return "position_id"
	case "work_id":
		return "work_id"
	case "role_ids":
		return "roles"
	case "updated_by":
		return "update_by"
	case "updated_at":
		return "updated_at"
	default:
		return ""
	}
}

func buildUserCondition(key string, value interface{}) (gen.Condition, error) {
	fieldName, operator := datautil.ParseFilterKey(key)

	switch fieldName {
	case "id":
		switch operator {
		case "in":
			ints, err := datautil.ToInt32Slice(value)
			if err != nil {
				return nil, err
			}
			return query.User.ID.In(ints...), nil
		default:
			intVal, err := datautil.ToInt32(value)
			if err != nil {
				return nil, err
			}
			return query.User.ID.Eq(intVal), nil
		}
	case "username":
		strVal := fmt.Sprint(value)
		if operator == "contains" || operator == "icontains" {
			return query.User.Username.Like("%" + strVal + "%"), nil
		}
		return query.User.Username.Eq(strVal), nil
	case "nickname":
		strVal := fmt.Sprint(value)
		if operator == "contains" || operator == "icontains" {
			return query.User.Nickname.Like("%" + strVal + "%"), nil
		}
		return query.User.Nickname.Eq(strVal), nil
	case "realname":
		strVal := fmt.Sprint(value)
		if operator == "contains" || operator == "icontains" {
			return query.User.Realname.Like("%" + strVal + "%"), nil
		}
		return query.User.Realname.Eq(strVal), nil
	case "email":
		strVal := fmt.Sprint(value)
		if operator == "contains" || operator == "icontains" {
			return query.User.Email.Like("%" + strVal + "%"), nil
		}
		return query.User.Email.Eq(strVal), nil
	case "mobile":
		strVal := fmt.Sprint(value)
		return query.User.Mobile.Eq(strVal), nil
	case "status":
		strVal := fmt.Sprint(value)
		return query.User.Status.Eq(strVal), nil
	case "authority":
		strVal := fmt.Sprint(value)
		return query.User.Authority.Eq(strVal), nil
	case "tenant_id":
		intVal, err := datautil.ToInt64(value)
		if err != nil {
			return nil, err
		}
		return query.User.TenantID.Eq(intVal), nil
	case "org_id":
		intVal, err := datautil.ToInt64(value)
		if err != nil {
			return nil, err
		}
		return query.User.OrgID.Eq(intVal), nil
	case "position_id":
		intVal, err := datautil.ToInt64(value)
		if err != nil {
			return nil, err
		}
		return query.User.PositionID.Eq(intVal), nil
	case "created_at":
		switch operator {
		case "gte":
			t, err := parseTime(value)
			if err != nil {
				return nil, err
			}
			return query.User.CreatedAt.Gte(t), nil
		case "lte":
			t, err := parseTime(value)
			if err != nil {
				return nil, err
			}
			return query.User.CreatedAt.Lte(t), nil
		}
	case "updated_at":
		switch operator {
		case "gte":
			t, err := parseTime(value)
			if err != nil {
				return nil, err
			}
			return query.User.UpdatedAt.Gte(t), nil
		case "lte":
			t, err := parseTime(value)
			if err != nil {
				return nil, err
			}
			return query.User.UpdatedAt.Lte(t), nil
		}
	}

	return nil, nil
}

func parseTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return time.Time{}, fmt.Errorf("empty time string")
		}
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			return ts, nil
		}
		if ts, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return ts, nil
		}
		return time.Time{}, fmt.Errorf("invalid time format %s", v)
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(int64(f), 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time type %T", value)
	}
}

func userStatusProtoToString(status userV1.User_Status) *string {
	name := userV1.User_Status_name[int32(status)]
	if name == "" {
		return nil
	}
	return &name
}

func userStatusStringToProto(status string) *userV1.User_Status {
	if value, ok := userV1.User_Status_value[strings.ToUpper(status)]; ok {
		enum := userV1.User_Status(value)
		return &enum
	}
	return nil
}

func userGenderProtoToString(gender userV1.User_Gender) *string {
	name := userV1.User_Gender_name[int32(gender)]
	if name == "" {
		return nil
	}
	return &name
}

func userGenderStringToProto(gender string) *userV1.User_Gender {
	if value, ok := userV1.User_Gender_value[strings.ToUpper(gender)]; ok {
		enum := userV1.User_Gender(value)
		return &enum
	}
	return nil
}

func userAuthorityProtoToString(authority userV1.User_Authority) *string {
	name := userV1.User_Authority_name[int32(authority)]
	if name == "" {
		return nil
	}
	return &name
}

func userAuthorityStringToProto(authority string) *userV1.User_Authority {
	if value, ok := userV1.User_Authority_value[strings.ToUpper(authority)]; ok {
		enum := userV1.User_Authority(value)
		return &enum
	}
	return nil
}
