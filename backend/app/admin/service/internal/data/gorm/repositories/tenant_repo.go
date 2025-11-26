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

type TenantRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewTenantRepo(db *gorm.DB, logger log.Logger) *TenantRepo {
	return &TenantRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "tenant/repo/admin-service")),
	}
}

// Count 统计租户数量
func (r *TenantRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Tenant{})

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

// List 获取租户列表
func (r *TenantRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListTenantResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	var tenants []models.Tenant
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Tenant{})

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

	if err := query.Find(&tenants).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*userV1.Tenant, 0, len(tenants))
	for _, tenant := range tenants {
		dto := r.toDTO(&tenant)
		dtos = append(dtos, dto)
	}

	return &userV1.ListTenantResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

// Get 根据ID获取租户
func (r *TenantRepo) Get(ctx context.Context, tenantId uint32) (*userV1.Tenant, error) {
	if tenantId == 0 {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorTenantNotFound("tenant not found")
		}
		r.log.Errorf("query tenant failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&tenant), nil
}

// Create 创建租户
func (r *TenantRepo) Create(ctx context.Context, data *userV1.Tenant) (*userV1.Tenant, error) {
	if data == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	tenant := r.fromRequest(data)

	// 设置创建时间
	if data.CreatedAt == nil {
		now := time.Now()
		tenant.CreatedAt = now
		tenant.UpdatedAt = now
	} else {
		createdAt := timeutil.TimestamppbToTime(data.CreatedAt)
		tenant.CreatedAt = *createdAt
		tenant.UpdatedAt = *createdAt
	}

	if err := r.db.WithContext(ctx).Create(&tenant).Error; err != nil {
		r.log.Errorf("create tenant failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(tenant), nil
}

// Update 更新租户
func (r *TenantRepo) Update(ctx context.Context, req *userV1.UpdateTenantRequest) error {
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
			_, err = r.Create(ctx, req.Data)
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

	if err := r.db.WithContext(ctx).Model(&models.Tenant{}).Where("id = ?", req.Data.GetId()).Updates(updateData).Error; err != nil {
		r.log.Errorf("update tenant failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

// Delete 删除租户
func (r *TenantRepo) Delete(ctx context.Context, tenantId uint32) error {
	if err := r.db.WithContext(ctx).Delete(&models.Tenant{}, tenantId).Error; err != nil {
		r.log.Errorf("delete tenant failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

// IsExist 检查租户是否存在
func (r *TenantRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Tenant{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check tenant exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

// GetTenantByTenantName 根据租户名称获取租户
func (r *TenantRepo) GetTenantByTenantName(ctx context.Context, name string) (*userV1.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorNotFound("tenant not found")
		}
		r.log.Errorf("query tenant by name failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&tenant), nil
}

// GetTenantByTenantCode 根据租户代码获取租户
func (r *TenantRepo) GetTenantByTenantCode(ctx context.Context, code string) (*userV1.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorNotFound("tenant not found")
		}
		r.log.Errorf("query tenant by code failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&tenant), nil
}

// TenantExists 检查租户是否存在
func (r *TenantRepo) TenantExists(ctx context.Context, req *userV1.TenantExistsRequest) (*userV1.TenantExistsResponse, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Tenant{}).Where("code = ?", req.GetCode()).Count(&count).Error; err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query exist failed")
	}

	return &userV1.TenantExistsResponse{
		Exist: count > 0,
	}, nil
}

// GetTenantsByIds 根据ID列表获取租户列表
func (r *TenantRepo) GetTenantsByIds(ctx context.Context, ids []uint32) ([]*userV1.Tenant, error) {
	if len(ids) == 0 {
		return []*userV1.Tenant{}, nil
	}

	var tenants []models.Tenant
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&tenants).Error; err != nil {
		r.log.Errorf("query tenants by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query tenants by ids failed")
	}

	dtos := make([]*userV1.Tenant, 0, len(tenants))
	for _, tenant := range tenants {
		dto := r.toDTO(&tenant)
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

// 辅助方法

// buildConditions 构建查询条件
func (r *TenantRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})

	// 处理查询字符串
	if req.GetQuery() != "" {
		// 实现查询逻辑
	}

	// 处理 OR 查询
	if req.GetOrQuery() != "" {
		// 实现 OR 查询逻辑
	}

	return conditions, nil
}

// applyConditions 应用查询条件
func (r *TenantRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		switch key {
		case "name":
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		case "code":
			query = query.Where("code = ?", value)
		case "status":
			query = query.Where("status = ?", value)
		default:
			query = query.Where(key+" = ?", value)
		}
	}
	return nil
}

// buildUpdateData 根据字段掩码构建更新数据
func (r *TenantRepo) buildUpdateData(data *userV1.Tenant, paths []string) map[string]interface{} {
	updateData := make(map[string]interface{})

	for _, path := range paths {
		switch path {
		case "name":
			if data.Name != nil {
				updateData["name"] = *data.Name
			}
		case "code":
			if data.Code != nil {
				updateData["code"] = *data.Code
			}
		case "status":
			if data.Status != nil {
				updateData["status"] = int32(*data.Status)
			}
		case "remark":
			if data.Remark != nil {
				updateData["remark"] = *data.Remark
			}
			// 添加更多字段处理...
		}
	}

	return updateData
}

// buildUpdateDataFromRequest 从请求构建完整的更新数据
func (r *TenantRepo) buildUpdateDataFromRequest(data *userV1.Tenant) map[string]interface{} {
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
	if data.CreatedBy != nil {
		updateData["create_by"] = *data.CreatedBy
	}
	if data.UpdatedBy != nil {
		updateData["update_by"] = *data.UpdatedBy
	}

	return updateData
}

// fromRequest 从请求构建模型
func (r *TenantRepo) fromRequest(data *userV1.Tenant) *models.Tenant {
	tenant := &models.Tenant{}

	if data.Name != nil {
		tenant.Name = data.Name
	}
	if data.Code != nil {
		tenant.Code = data.Code
	}
	if data.Status != nil {
		status := int32(*data.Status)
		tenant.Status = &status
	}
	if data.Remark != nil {
		tenant.Remark = data.Remark
	}
	if data.CreatedBy != nil {
		tenant.CreatedBy = data.CreatedBy
	}
	if data.UpdatedBy != nil {
		tenant.UpdatedBy = data.UpdatedBy
	}

	return tenant
}

// toDTO 将模型转换为 DTO
func (r *TenantRepo) toDTO(tenant *models.Tenant) *userV1.Tenant {
	dto := &userV1.Tenant{
		Id: &tenant.ID,
	}

	if tenant.Name != nil {
		dto.Name = tenant.Name
	}
	if tenant.Code != nil {
		dto.Code = tenant.Code
	}
	if tenant.Status != nil {
		status := userV1.Tenant_Status(*tenant.Status)
		dto.Status = &status
	}
	if tenant.Remark != nil {
		dto.Remark = tenant.Remark
	}
	if tenant.CreatedBy != nil {
		dto.CreatedBy = tenant.CreatedBy
	}
	if tenant.UpdatedBy != nil {
		dto.UpdatedBy = tenant.UpdatedBy
	}

	// 设置时间字段
	dto.CreatedAt = timeutil.TimeToTimestamppb(&tenant.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(&tenant.UpdatedAt)

	return dto
}
