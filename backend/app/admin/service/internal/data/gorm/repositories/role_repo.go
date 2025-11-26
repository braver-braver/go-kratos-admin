package repositories

import (
	"context"
	"time"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gorm"
)

type RoleRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewRoleRepo(db *gorm.DB, logger log.Logger) *RoleRepo {
	return &RoleRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "role/repo/admin-service")),
	}
}

func (r *RoleRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Role{})

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

func (r *RoleRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListRoleResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	var roles []models.Role
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Role{})

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

	if err := query.Find(&roles).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*userV1.Role, 0, len(roles))
	for _, role := range roles {
		dto := r.toDTO(&role)
		dtos = append(dtos, dto)
	}

	return &userV1.ListRoleResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *RoleRepo) Get(ctx context.Context, roleId uint32) (*userV1.Role, error) {
	if roleId == 0 {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, roleId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorRoleNotFound("role not found")
		}
		r.log.Errorf("query role failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&role), nil
}

func (r *RoleRepo) Create(ctx context.Context, req *userV1.CreateRoleRequest) (*userV1.Role, error) {
	if req == nil || req.Data == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	role := r.fromCreateRequest(req)

	// 设置创建时间
	if req.Data.CreatedAt == nil {
		role.CreatedAt = time.Now()
	} else {
		createdAt := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		role.CreatedAt = *createdAt
	}

	if err := r.db.WithContext(ctx).Create(&role).Error; err != nil {
		r.log.Errorf("create role failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(role), nil
}

func (r *RoleRepo) Update(ctx context.Context, req *userV1.UpdateRoleRequest) error {
	if req == nil || req.Data == nil {
		return userV1.ErrorBadRequest("invalid parameter")
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

	if err := r.db.WithContext(ctx).Model(&models.Role{}).Where(
		"id = ?", req.Data.GetId(),
	).Updates(updateData).Error; err != nil {
		r.log.Errorf("update role failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *RoleRepo) Delete(ctx context.Context, roleId uint32) error {
	if err := r.db.WithContext(ctx).Delete(&models.Role{}, roleId).Error; err != nil {
		r.log.Errorf("delete role failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *RoleRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Role{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check role exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *RoleRepo) GetRoleByRoleCode(ctx context.Context, roleCode string) (*userV1.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("code = ?", roleCode).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorRoleNotFound("role not found")
		}
		r.log.Errorf("query role by code failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&role), nil
}

func (r *RoleRepo) GetRolesByIds(ctx context.Context, ids []uint32) ([]*userV1.Role, error) {
	if len(ids) == 0 {
		return []*userV1.Role{}, nil
	}

	var roles []models.Role
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&roles).Error; err != nil {
		r.log.Errorf("query roles by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query roles by ids failed")
	}

	dtos := make([]*userV1.Role, 0, len(roles))
	for _, role := range roles {
		dto := r.toDTO(&role)
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

func (r *RoleRepo) GetRoleByRoleName(ctx context.Context, roleName string) (*userV1.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorRoleNotFound("role not found")
		}
		r.log.Errorf("query role by name failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&role), nil
}

// 辅助方法

func (r *RoleRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})
	return conditions, nil
}

func (r *RoleRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	return nil
}

func (r *RoleRepo) buildUpdateData(data *userV1.Role, paths []string) map[string]interface{} {
	updateData := make(map[string]interface{})
	// Implementation for field mask updates
	return updateData
}

func (r *RoleRepo) buildUpdateDataFromRequest(data *userV1.Role) map[string]interface{} {
	updateData := make(map[string]interface{})
	if data.Name != nil {
		updateData["name"] = *data.Name
	}
	if data.Code != nil {
		updateData["code"] = *data.Code
	}
	if data.Status != nil {
		updateData["status"] = int32(*data.Status)
	}
	if data.Remark != nil {
		updateData["remark"] = *data.Remark
	}
	if data.UpdatedAt != nil {
		updateData["updated_at"] = timeutil.TimestamppbToTime(data.UpdatedAt)
	}
	if data.UpdatedBy != nil {
		updateData["update_by"] = *data.UpdatedBy
	}
	return updateData
}

func (r *RoleRepo) fromCreateRequest(req *userV1.CreateRoleRequest) *models.Role {
	role := &models.Role{
		Name:   req.Data.Name,
		Code:   req.Data.Code,
		Remark: req.Data.Remark,
	}

	if req.Data.Status != nil {
		status := int32(*req.Data.Status)
		role.Status = &status
	}
	if req.Data.CreatedBy != nil {
		role.CreatedBy = req.Data.CreatedBy
	}
	if req.Data.TenantId != nil {
		role.TenantID = req.Data.TenantId
	}

	return role
}

func (r *RoleRepo) toDTO(role *models.Role) *userV1.Role {
	dto := &userV1.Role{
		Id: &role.ID,
	}

	if role.Name != nil {
		dto.Name = role.Name
	}
	if role.Code != nil {
		dto.Code = role.Code
	}
	if role.Status != nil {
		status := userV1.Role_Status(*role.Status)
		dto.Status = &status
	}
	if role.Remark != nil {
		dto.Remark = role.Remark
	}
	if role.TenantID != nil {
		dto.TenantId = role.TenantID
	}
	if role.CreatedBy != nil {
		dto.CreatedBy = role.CreatedBy
	}
	if role.UpdatedBy != nil {
		dto.UpdatedBy = role.UpdatedBy
	}

	dto.CreatedAt = timeutil.TimeToTimestamppb(&role.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(&role.UpdatedAt)

	return dto
}
