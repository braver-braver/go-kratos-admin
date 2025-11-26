package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gorm"
)

type UserRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewUserRepo(db *gorm.DB, logger log.Logger) *UserRepo {
	return &UserRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "user/repo/admin-service")),
	}
}

// Count 统计用户数量
func (r *UserRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.User{})

	// 应用查询条件
	if err := r.applyConditions(query, conditions); err != nil {
		return 0, err
	}

	if err := query.Count(&count).Error; err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, userV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

// List 获取用户列表
func (r *UserRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListUserResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	var users []models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	// 构建查询条件
	conditions, err := r.buildConditions(req)
	if err != nil {
		return nil, err
	}

	// 应用查询条件
	if err := r.applyConditions(query, conditions); err != nil {
		return nil, err
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query count failed")
	}

	// 分页查询
	if !req.GetNoPaging() {
		offset := (req.GetPage() - 1) * req.GetPageSize()
		query = query.Offset(int(offset)).Limit(int(req.GetPageSize()))
	}

	// 排序
	if len(req.GetOrderBy()) > 0 {
		query = query.Order(req.GetOrderBy()[0])
	} else {
		query = query.Order("created_at DESC")
	}

	// 字段掩码处理
	if req.GetFieldMask() != nil && len(req.GetFieldMask().GetPaths()) > 0 {
		query = query.Select(req.GetFieldMask().GetPaths())
	}

	if err := query.Find(&users).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*userV1.User, 0, len(users))
	for _, user := range users {
		dto := r.toDTO(&user)
		dtos = append(dtos, dto)
	}

	return &userV1.ListUserResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

// Get 根据ID获取用户
func (r *UserRepo) Get(ctx context.Context, userId uint32) (*userV1.User, error) {
	if userId == 0 {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	var user models.User
	if err := r.db.WithContext(ctx).First(&user, userId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorUserNotFound("user not found")
		}
		r.log.Errorf("query user failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&user), nil
}

// GetUserByUserName 根据用户名获取用户
func (r *UserRepo) GetUserByUserName(ctx context.Context, userName string) (*userV1.User, error) {
	if userName == "" {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	var user models.User
	if err := r.db.WithContext(ctx).Where("username = ?", userName).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorUserNotFound("user not found")
		}
		r.log.Errorf("query user by username failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&user), nil
}

// Create 创建用户
func (r *UserRepo) Create(ctx context.Context, req *userV1.CreateUserRequest) (*userV1.User, error) {
	if req == nil || req.Data == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	user := r.fromCreateRequest(req)

	// 设置创建时间
	if req.Data.CreatedAt == nil {
		now := time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now
	} else {
		createTime := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		user.CreatedAt = *createTime
		user.UpdatedAt = *createTime
	}

	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		r.log.Errorf("create user failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(user), nil
}

// Update 更新用户
func (r *UserRepo) Update(ctx context.Context, req *userV1.UpdateUserRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
	}

	// 如果不存在则创建
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

	// 处理字段掩码
	updateData := make(map[string]interface{})
	if req.UpdateMask != nil {
		req.UpdateMask.Normalize()
		if !req.UpdateMask.IsValid(req.Data) {
			r.log.Errorf("invalid field mask [%v]", req.UpdateMask)
			return userV1.ErrorBadRequest("invalid field mask")
		}

		// 根据字段掩码构建更新数据
		updateData = r.buildUpdateData(req.Data, req.UpdateMask.GetPaths())
	} else {
		// 更新所有字段
		updateData = r.buildUpdateDataFromRequest(req.Data)
	}

	// 设置更新时间
	updateData["updated_at"] = time.Now()

	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", req.Data.GetId()).Updates(updateData).Error; err != nil {
		r.log.Errorf("update user failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

// Delete 删除用户
func (r *UserRepo) Delete(ctx context.Context, userId uint32) error {
	if err := r.db.WithContext(ctx).Delete(&models.User{}, userId).Error; err != nil {
		r.log.Errorf("delete user failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

// IsExist 检查用户是否存在
func (r *UserRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check user exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

// UserExists 检查用户名是否存在
func (r *UserRepo) UserExists(ctx context.Context, req *userV1.UserExistsRequest) (*userV1.UserExistsResponse, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", req.GetUsername()).Count(&count).Error; err != nil {
		r.log.Errorf("check username exist failed: %s", err.Error())
		return &userV1.UserExistsResponse{Exist: false}, userV1.ErrorInternalServerError("query exist failed")
	}

	return &userV1.UserExistsResponse{
		Exist: count > 0,
	}, nil
}

// GetUsersByIds 根据ID列表获取用户列表
func (r *UserRepo) GetUsersByIds(ctx context.Context, ids []uint32) ([]*userV1.User, error) {
	if len(ids) == 0 {
		return []*userV1.User{}, nil
	}

	var users []models.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		r.log.Errorf("query users by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query users by ids failed")
	}

	dtos := make([]*userV1.User, 0, len(users))
	for _, user := range users {
		dto := r.toDTO(&user)
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

// 辅助方法

// buildConditions 构建查询条件
func (r *UserRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})

	// 处理查询字符串
	if req.GetQuery() != "" {
		// 这里需要解析查询字符串，例如：username:admin,status:1
		// 简化实现，实际需要更复杂的解析逻辑
		queryParts := strings.Split(req.GetQuery(), ",")
		for _, part := range queryParts {
			if kv := strings.Split(part, ":"); len(kv) == 2 {
				conditions[kv[0]] = kv[1]
			}
		}
	}

	// 处理 OR 查询
	if req.GetOrQuery() != "" {
		// 实现 OR 查询逻辑
	}

	return conditions, nil
}

// applyConditions 应用查询条件
func (r *UserRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		switch key {
		case "username":
			query = query.Where("username LIKE ?", "%"+fmt.Sprintf("%v", value)+"%")
		case "status":
			query = query.Where("status = ?", value)
		case "tenant_id":
			query = query.Where("tenant_id = ?", value)
		case "org_id":
			query = query.Where("org_id = ?", value)
		default:
			query = query.Where(key+" = ?", value)
		}
	}
	return nil
}

// buildUpdateData 根据字段掩码构建更新数据
func (r *UserRepo) buildUpdateData(data *userV1.User, paths []string) map[string]interface{} {
	updateData := make(map[string]interface{})

	for _, path := range paths {
		switch path {
		case "nickname":
			if data.Nickname != nil {
				updateData["nickname"] = *data.Nickname
			}
		case "realname":
			if data.Realname != nil {
				updateData["realname"] = *data.Realname
			}
		case "email":
			if data.Email != nil {
				updateData["email"] = *data.Email
			}
		case "mobile":
			if data.Mobile != nil {
				updateData["mobile"] = *data.Mobile
			}
		case "status":
			if data.Status != nil {
				updateData["status"] = *data.Status
			}
		case "authority":
			if data.Authority != nil {
				updateData["authority"] = data.Authority.String()
			}
		case "gender":
			if data.Gender != nil {
				updateData["gender"] = data.Gender.String()
			}
		case "roles":
			if data.Roles != nil {
				// 将角色数组转换为 JSON 字符串
				rolesJSON := r.rolesToJSON(data.Roles)
				updateData["roles"] = rolesJSON
			}
			// 添加更多字段处理...
		}
	}

	return updateData
}

// buildUpdateDataFromRequest 从请求构建完整的更新数据
func (r *UserRepo) buildUpdateDataFromRequest(data *userV1.User) map[string]interface{} {
	updateData := make(map[string]interface{})

	if data.Nickname != nil {
		updateData["nickname"] = *data.Nickname
	}
	if data.Realname != nil {
		updateData["realname"] = *data.Realname
	}
	if data.Email != nil {
		updateData["email"] = *data.Email
	}
	if data.Mobile != nil {
		updateData["mobile"] = *data.Mobile
	}
	if data.Status != nil {
		updateData["status"] = *data.Status
	}
	if data.Authority != nil {
		updateData["authority"] = data.Authority.String()
	}
	if data.Gender != nil {
		updateData["gender"] = data.Gender.String()
	}
	if data.Roles != nil {
		updateData["roles"] = r.rolesToJSON(data.Roles)
	}

	return updateData
}

// fromCreateRequest 从创建请求构建模型
func (r *UserRepo) fromCreateRequest(req *userV1.CreateUserRequest) *models.User {
	user := &models.User{}

	if req.Data.Username != nil {
		user.Username = req.Data.Username
	}
	if req.Data.Nickname != nil {
		user.Nickname = req.Data.Nickname
	}
	if req.Data.Realname != nil {
		user.Realname = req.Data.Realname
	}
	if req.Data.Email != nil {
		user.Email = req.Data.Email
	}
	if req.Data.Mobile != nil {
		user.Mobile = req.Data.Mobile
	}
	if req.Data.Telephone != nil {
		user.Telephone = req.Data.Telephone
	}
	if req.Data.Avatar != nil {
		user.Avatar = req.Data.Avatar
	}
	if req.Data.Address != nil {
		user.Address = req.Data.Address
	}
	if req.Data.Region != nil {
		user.Region = req.Data.Region
	}
	if req.Data.Description != nil {
		user.Description = req.Data.Description
	}
	if req.Data.Gender != nil {
		gender := req.Data.Gender.String()
		user.Gender = &gender
	}
	if req.Data.Authority != nil {
		authority := req.Data.Authority.String()
		user.Authority = &authority
	}
	if req.Data.LastLoginTime != nil {
		user.LastLoginTime = timeutil.TimestamppbToTime(req.Data.LastLoginTime)
	}
	if req.Data.LastLoginIp != nil {
		user.LastLoginIP = req.Data.LastLoginIp
	}
	if req.Data.OrgId != nil {
		user.OrgID = req.Data.OrgId
	}
	if req.Data.PositionId != nil {
		user.PositionID = req.Data.PositionId
	}
	if req.Data.WorkId != nil {
		workID := uint32(*req.Data.WorkId)
		user.WorkID = &workID
	}
	if req.Data.TenantId != nil {
		user.TenantID = req.Data.TenantId
	}
	if req.Data.CreatedBy != nil {
		user.CreatedBy = req.Data.CreatedBy
	}
	if req.Data.UpdatedBy != nil {
		user.UpdatedBy = req.Data.UpdatedBy
	}
	if req.Data.Remark != nil {
		user.Remark = req.Data.Remark
	}
	if req.Data.Status != nil {
		status := int32(*req.Data.Status)
		user.Status = &status
	}
	if req.Data.Roles != nil {
		user.Roles = r.rolesToJSON(req.Data.Roles)
	}

	return user
}

// toDTO 将模型转换为 DTO
func (r *UserRepo) toDTO(user *models.User) *userV1.User {
	dto := &userV1.User{
		Id: &user.ID,
	}

	if user.Username != nil {
		dto.Username = user.Username
	}
	if user.Nickname != nil {
		dto.Nickname = user.Nickname
	}
	if user.Realname != nil {
		dto.Realname = user.Realname
	}
	if user.Email != nil {
		dto.Email = user.Email
	}
	if user.Mobile != nil {
		dto.Mobile = user.Mobile
	}
	if user.Telephone != nil {
		dto.Telephone = user.Telephone
	}
	if user.Avatar != nil {
		dto.Avatar = user.Avatar
	}
	if user.Address != nil {
		dto.Address = user.Address
	}
	if user.Region != nil {
		dto.Region = user.Region
	}
	if user.Description != nil {
		dto.Description = user.Description
	}
	if user.Gender != nil {
		dto.Gender = r.stringToGenderEnum(*user.Gender)
	}
	if user.Authority != nil {
		dto.Authority = r.stringToAuthorityEnum(*user.Authority)
	}
	if user.LastLoginTime != nil {
		dto.LastLoginTime = timeutil.TimeToTimestamppb(user.LastLoginTime)
	}
	if user.LastLoginIP != nil {
		dto.LastLoginIp = user.LastLoginIP
	}
	if user.OrgID != nil {
		dto.OrgId = user.OrgID
	}
	if user.PositionID != nil {
		dto.PositionId = user.PositionID
	}
	if user.WorkID != nil {
		workID := uint32(*user.WorkID)
		dto.WorkId = &workID
	}
	if user.TenantID != nil {
		dto.TenantId = user.TenantID
	}
	if user.CreatedBy != nil {
		dto.CreatedBy = user.CreatedBy
	}
	if user.UpdatedBy != nil {
		dto.UpdatedBy = user.UpdatedBy
	}
	if user.Remark != nil {
		dto.Remark = user.Remark
	}
	if user.Status != nil {
		status := userV1.User_Status(*user.Status)
		dto.Status = &status
	}
	if user.Roles != nil {
		dto.Roles = r.jsonToRoles(*user.Roles)
	}

	// 设置时间字段
	dto.CreatedAt = timeutil.TimeToTimestamppb(&user.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(&user.UpdatedAt)

	return dto
}

// 枚举转换辅助方法
func (r *UserRepo) stringToGenderEnum(s string) *userV1.User_Gender {
	switch s {
	case "SECRET":
		gender := userV1.User_SECRET
		return &gender
	case "MALE":
		gender := userV1.User_MALE
		return &gender
	case "FEMALE":
		gender := userV1.User_FEMALE
		return &gender
	default:
		return nil
	}
}

func (r *UserRepo) stringToAuthorityEnum(s string) *userV1.User_Authority {
	switch s {
	case "SYS_ADMIN":
		authority := userV1.User_SYS_ADMIN
		return &authority
	case "TENANT_ADMIN":
		authority := userV1.User_TENANT_ADMIN
		return &authority
	case "CUSTOMER_USER":
		authority := userV1.User_CUSTOMER_USER
		return &authority
	case "GUEST":
		authority := userV1.User_GUEST
		return &authority
	default:
		return nil
	}
}

// JSON 处理辅助方法
func (r *UserRepo) rolesToJSON(roles []string) *string {
	if len(roles) == 0 {
		return nil
	}
	// 这里需要使用 JSON 序列化
	// 简化实现
	rolesStr := strings.Join(roles, ",")
	return &rolesStr
}

func (r *UserRepo) jsonToRoles(rolesJSON string) []string {
	if rolesJSON == "" {
		return nil
	}
	// 这里需要使用 JSON 反序列化
	// 简化实现
	return strings.Split(rolesJSON, ",")
}
