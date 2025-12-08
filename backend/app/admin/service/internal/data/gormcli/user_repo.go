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

type UserRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewUserRepo(db *gorm.DB, logger log.Logger) *UserRepo {
	repo := &UserRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "user/gormcli")),
	}
	return repo
}

func (r *UserRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysUser](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, userV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *UserRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListUserResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderUser(req.GetOrderBy()),
		scopeFieldMaskUser(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysUser](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
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

	dtos := make([]*userV1.User, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, r.toUserDTO(&entities[i]))
	}

	return &userV1.ListUserResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *UserRepo) Get(ctx context.Context, req *userV1.GetUserRequest) (*userV1.User, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	g := gorm.G[model.SysUser](r.db.WithContext(ctx)).Scopes(scopeFieldMaskUser(req.GetViewMask().GetPaths()))

	var (
		entity model.SysUser
		err    error
	)
	switch req.GetQueryBy().(type) {
	case *userV1.GetUserRequest_Id:
		entity, err = g.Where("id = ?", req.GetId()).Take(ctx)
	case *userV1.GetUserRequest_Username:
		entity, err = g.Where("username = ?", req.GetUsername()).Take(ctx)
	default:
		return nil, userV1.ErrorBadRequest("invalid query parameter")
	}
	if err != nil {
		return r.handleGetErr(err)
	}
	return r.toUserDTO(&entity), nil
}

func (r *UserRepo) Create(ctx context.Context, req *userV1.CreateUserRequest) (*userV1.User, error) {
	if req == nil || req.Data == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}
	entity := r.toUserModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if err := gorm.G[model.SysUser](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("insert data failed")
	}
	return r.toUserDTO(entity), nil
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

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch p {
		case "nickname":
			updates["nickname"] = req.Data.GetNickname()
		case "realname":
			updates["realname"] = req.Data.GetRealname()
		case "avatar":
			updates["avatar"] = req.Data.GetAvatar()
		case "email":
			updates["email"] = req.Data.GetEmail()
		case "mobile":
			updates["mobile"] = req.Data.GetMobile()
		case "telephone":
			updates["telephone"] = req.Data.GetTelephone()
		case "region":
			updates["region"] = req.Data.GetRegion()
		case "address":
			updates["address"] = req.Data.GetAddress()
		case "description":
			updates["description"] = req.Data.GetDescription()
		case "remark":
			updates["remark"] = req.Data.GetRemark()
		case "last_login_time":
			updates["last_login_time"] = req.Data.GetLastLoginTime().AsTime()
		case "last_login_ip":
			updates["last_login_ip"] = req.Data.GetLastLoginIp()
		case "status":
			updates["status"] = req.Data.GetStatus().String()
		case "gender":
			updates["gender"] = req.Data.GetGender().String()
		case "authority":
			updates["authority"] = req.Data.GetAuthority().String()
		case "org_id":
			updates["org_id"] = req.Data.GetOrgId()
		case "department_id":
			updates["department_id"] = req.Data.GetDepartmentId()
		case "position_id":
			updates["position_id"] = req.Data.GetPositionId()
		case "work_id":
			updates["work_id"] = req.Data.GetWorkId()
		case "role_ids":
			updates["role_ids"] = req.Data.GetRoleIds()
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
	db := r.db.WithContext(ctx).Model(&model.SysUser{}).Where("id = ?", req.Data.GetId())
	if err := db.Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, userId uint32) error {
	if err := r.db.WithContext(ctx).Where("id = ?", userId).Delete(&model.SysUser{}).Error; err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

// GetUserByUserName 根据用户名获取用户
func (r *UserRepo) GetUserByUserName(ctx context.Context, userName string) (*userV1.User, error) {
	if userName == "" {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.SysUser](r.db.WithContext(ctx)).
		Where("username = ?", userName).
		Take(ctx)
	if err != nil {
		return r.handleGetErr(err)
	}
	return r.toUserDTO(&entity), nil
}

// GetUsersByIds 根据ID列表获取用户列表
func (r *UserRepo) GetUsersByIds(ctx context.Context, ids []uint32) ([]*userV1.User, error) {
	if len(ids) == 0 {
		return []*userV1.User{}, nil
	}

	entities, err := gorm.G[model.SysUser](r.db.WithContext(ctx)).
		Where("id IN ?", ids).
		Find(ctx)
	if err != nil {
		r.log.Errorf("query user by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query user by ids failed")
	}

	dtos := make([]*userV1.User, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, r.toUserDTO(&entities[i]))
	}

	return dtos, nil
}

func (r *UserRepo) UserExists(ctx context.Context, req *userV1.UserExistsRequest) (*userV1.UserExistsResponse, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table(model.TableNameSysUser).Where(
		"username = ?", req.GetUsername(),
	).Count(&count).Error; err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return &userV1.UserExistsResponse{Exist: false}, userV1.ErrorInternalServerError("query exist failed")
	}
	return &userV1.UserExistsResponse{Exist: count > 0}, nil
}

func (r *UserRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysUser](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *UserRepo) toUserDTO(entity *model.SysUser) *userV1.User {
	dto := &userV1.User{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	dto.Username = &entity.Username
	dto.Nickname = &entity.Nickname
	dto.Realname = &entity.Realname
	dto.TenantId = toUint32Ptr(entity.TenantID)
	dto.Avatar = &entity.Avatar
	dto.Email = &entity.Email
	dto.Mobile = &entity.Mobile
	dto.Telephone = &entity.Telephone
	dto.Region = &entity.Region
	dto.Address = &entity.Address
	dto.Description = &entity.Description
	dto.Remark = &entity.Remark
	if !entity.LastLoginTime.IsZero() {
		dto.LastLoginTime = timestamppb.New(entity.LastLoginTime)
	}
	dto.LastLoginIp = &entity.LastLoginIP
	if entity.Status != "" {
		if v, ok := userV1.User_Status_value[entity.Status]; ok {
			status := userV1.User_Status(v)
			dto.Status = &status
		}
	}
	if entity.Gender != "" {
		if v, ok := userV1.User_Gender_value[entity.Gender]; ok {
			g := userV1.User_Gender(v)
			dto.Gender = &g
		}
	}
	if entity.Authority != "" {
		if v, ok := userV1.User_Authority_value[entity.Authority]; ok {
			a := userV1.User_Authority(v)
			dto.Authority = &a
		}
	}
	dto.OrgId = toUint32Ptr(entity.OrgID)
	dto.DepartmentId = toUint32Ptr(entity.DepartmentID)
	dto.PositionId = toUint32Ptr(entity.PositionID)
	dto.WorkId = toUint32Ptr(entity.WorkID)
	dto.RoleIds, _ = entityhelper.ParseUint32SliceFromJSONArrayString(entity.RoleIds)
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

func (r *UserRepo) toUserModel(dto *userV1.User) *model.SysUser {
	if dto == nil {
		return nil
	}
	entity := &model.SysUser{
		ID:           int64(dto.GetId()),
		TenantID:     int64(dto.GetTenantId()),
		Username:     dto.GetUsername(),
		Nickname:     dto.GetNickname(),
		Realname:     dto.GetRealname(),
		Avatar:       dto.GetAvatar(),
		Email:        dto.GetEmail(),
		Mobile:       dto.GetMobile(),
		Telephone:    dto.GetTelephone(),
		Region:       dto.GetRegion(),
		Address:      dto.GetAddress(),
		Description:  dto.GetDescription(),
		Remark:       dto.GetRemark(),
		Status:       "",
		Gender:       "",
		Authority:    "",
		OrgID:        int64(dto.GetOrgId()),
		DepartmentID: int64(dto.GetDepartmentId()),
		PositionID:   int64(dto.GetPositionId()),
		WorkID:       int64(dto.GetWorkId()),
		RoleIds:      "",
		CreatedBy:    int64(dto.GetCreatedBy()),
		UpdatedBy:    int64(dto.GetUpdatedBy()),
	}
	if dto.Status != nil {
		entity.Status = dto.GetStatus().String()
	}
	if dto.Gender != nil {
		entity.Gender = dto.GetGender().String()
	}
	if dto.Authority != nil {
		entity.Authority = dto.GetAuthority().String()
	}
	if dto.LastLoginTime != nil {
		entity.LastLoginTime = dto.LastLoginTime.AsTime()
	}
	entity.LastLoginIP = dto.GetLastLoginIp()
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.CreatedAt.AsTime()
	}
	if dto.UpdatedAt != nil {
		entity.UpdatedAt = dto.UpdatedAt.AsTime()
	}
	return entity
}

var userColumns = map[string]string{
	"id":              "id",
	"username":        "username",
	"nickname":        "nickname",
	"realname":        "realname",
	"avatar":          "avatar",
	"email":           "email",
	"mobile":          "mobile",
	"telephone":       "telephone",
	"region":          "region",
	"address":         "address",
	"description":     "description",
	"remark":          "remark",
	"last_login_time": "last_login_time",
	"last_login_ip":   "last_login_ip",
	"status":          "status",
	"gender":          "gender",
	"authority":       "authority",
	"org_id":          "org_id",
	"department_id":   "department_id",
	"position_id":     "position_id",
	"work_id":         "work_id",
	"role_ids":        "role_ids",
	"created_by":      "created_by",
	"updated_by":      "updated_by",
	"created_at":      "created_at",
	"updated_at":      "updated_at",
}

func scopeOrderUser(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := userColumns[key]; ok {
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

func scopeFieldMaskUser(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := userColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func (r *UserRepo) handleGetErr(err error) (*userV1.User, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, userV1.ErrorUserNotFound("user not found")
	}
	r.log.Errorf("query one data failed: %s", err.Error())
	return nil, userV1.ErrorInternalServerError("query data failed")
}
