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

type DepartmentRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewDepartmentRepo(db *gorm.DB, logger log.Logger) *DepartmentRepo {
	return &DepartmentRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "department/repo/admin-service")),
	}
}

// Count 统计部门数量
func (r *DepartmentRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Department{})

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

// List 获取部门列表
func (r *DepartmentRepo) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListDepartmentResponse, error) {
	if req == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	var departments []models.Department
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Department{})

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

	if err := query.Find(&departments).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*userV1.Department, 0, len(departments))
	for _, department := range departments {
		dto := r.toDTO(&department)
		dtos = append(dtos, dto)
	}

	return &userV1.ListDepartmentResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

// Get 根据ID获取部门
func (r *DepartmentRepo) Get(ctx context.Context, departmentId uint32) (*userV1.Department, error) {
	if departmentId == 0 {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	var department models.Department
	if err := r.db.WithContext(ctx).First(&department, departmentId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, userV1.ErrorDepartmentNotFound("department not found")
		}
		r.log.Errorf("query department failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&department), nil
}

// Create 创建部门
func (r *DepartmentRepo) Create(ctx context.Context, req *userV1.CreateDepartmentRequest) (*userV1.Department, error) {
	if req == nil || req.Data == nil {
		return nil, userV1.ErrorBadRequest("invalid parameter")
	}

	department := r.fromCreateRequest(req)

	// 设置创建时间
	if req.Data.CreatedAt == nil {
		now := time.Now()
		department.CreatedAt = now
		department.UpdatedAt = now
	} else {
		createdAt := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		department.CreatedAt = *createdAt
		department.UpdatedAt = *createdAt
	}

	if err := r.db.WithContext(ctx).Create(&department).Error; err != nil {
		r.log.Errorf("create department failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(department), nil
}

// Update 更新部门
func (r *DepartmentRepo) Update(ctx context.Context, req *userV1.UpdateDepartmentRequest) error {
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
			createReq := &userV1.CreateDepartmentRequest{Data: req.Data}
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

	if err := r.db.WithContext(ctx).Model(&models.Department{}).Where("id = ?", req.Data.GetId()).Updates(updateData).Error; err != nil {
		r.log.Errorf("update department failed: %s", err.Error())
		return userV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

// Delete 删除部门
func (r *DepartmentRepo) Delete(ctx context.Context, departmentId uint32) error {
	if err := r.db.WithContext(ctx).Delete(&models.Department{}, departmentId).Error; err != nil {
		r.log.Errorf("delete department failed: %s", err.Error())
		return userV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

// IsExist 检查部门是否存在
func (r *DepartmentRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Department{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check department exist failed: %s", err.Error())
		return false, userV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

// GetDepartmentsByIds 根据ID列表获取部门列表
func (r *DepartmentRepo) GetDepartmentsByIds(ctx context.Context, ids []uint32) ([]*userV1.Department, error) {
	if len(ids) == 0 {
		return []*userV1.Department{}, nil
	}

	var departments []models.Department
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&departments).Error; err != nil {
		r.log.Errorf("query departments by ids failed: %s", err.Error())
		return nil, userV1.ErrorInternalServerError("query departments by ids failed")
	}

	dtos := make([]*userV1.Department, 0, len(departments))
	for _, department := range departments {
		dto := r.toDTO(&department)
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

// 辅助方法

// buildConditions 构建查询条件
func (r *DepartmentRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
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
func (r *DepartmentRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		switch key {
		case "name":
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
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
func (r *DepartmentRepo) buildUpdateData(data *userV1.Department, paths []string) map[string]interface{} {
	updateData := make(map[string]interface{})

	for _, path := range paths {
		switch path {
		case "name":
			if data.Name != nil {
				updateData["name"] = *data.Name
			}
		case "parent_id":
			if data.ParentId != nil {
				updateData["parent_id"] = *data.ParentId
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
func (r *DepartmentRepo) buildUpdateDataFromRequest(data *userV1.Department) map[string]interface{} {
	updateData := make(map[string]interface{})

	if data.Name != nil {
		updateData["name"] = *data.Name
	}
	if data.ParentId != nil {
		updateData["parent_id"] = *data.ParentId
	}
	if data.Status != nil {
		updateData["status"] = int32(*data.Status)
	}
	if data.Remark != nil {
		updateData["remark"] = *data.Remark
	}
	if data.SortOrder != nil {
		updateData["sort_order"] = *data.SortOrder
	}
	if data.Description != nil {
		updateData["description"] = *data.Description
	}
	if data.ManagerId != nil {
		updateData["manager_id"] = *data.ManagerId
	}
	if data.OrganizationId != nil {
		updateData["organization_id"] = *data.OrganizationId
	}
	if data.TenantId != nil {
		updateData["tenant_id"] = *data.TenantId
	}
	if data.CreatedBy != nil {
		updateData["create_by"] = *data.CreatedBy
	}
	if data.UpdatedBy != nil {
		updateData["update_by"] = *data.UpdatedBy
	}

	return updateData
}

// fromCreateRequest 从创建请求构建模型
func (r *DepartmentRepo) fromCreateRequest(req *userV1.CreateDepartmentRequest) *models.Department {
	department := &models.Department{}

	if req.Data.Name != nil {
		department.Name = req.Data.Name
	}
	if req.Data.ParentId != nil {
		department.ParentID = req.Data.ParentId
	}
	if req.Data.SortOrder != nil {
		department.SortOrder = req.Data.SortOrder
	}
	if req.Data.Remark != nil {
		department.Remark = req.Data.Remark
	}
	if req.Data.Status != nil {
		status := int32(*req.Data.Status)
		department.Status = &status
	}
	if req.Data.OrganizationId != nil {
		department.OrganizationID = req.Data.OrganizationId
	}
	if req.Data.ManagerId != nil {
		department.ManagerID = req.Data.ManagerId
	}
	if req.Data.Description != nil {
		department.Description = req.Data.Description
	}
	if req.Data.CreatedBy != nil {
		department.CreatedBy = req.Data.CreatedBy
	}
	if req.Data.UpdatedBy != nil {
		department.UpdatedBy = req.Data.UpdatedBy
	}
	if req.Data.TenantId != nil {
		department.TenantID = req.Data.TenantId
	}

	return department
}

// toDTO 将模型转换为 DTO
func (r *DepartmentRepo) toDTO(department *models.Department) *userV1.Department {
	dto := &userV1.Department{
		Id: &department.ID,
	}

	if department.Name != nil {
		dto.Name = department.Name
	}
	if department.ParentID != nil {
		dto.ParentId = department.ParentID
	}
	if department.SortOrder != nil {
		dto.SortOrder = department.SortOrder
	}
	if department.Remark != nil {
		dto.Remark = department.Remark
	}
	if department.Status != nil {
		status := userV1.Department_Status(*department.Status)
		dto.Status = &status
	}
	if department.OrganizationID != nil {
		dto.OrganizationId = department.OrganizationID
	}
	if department.ManagerID != nil {
		dto.ManagerId = department.ManagerID
	}
	if department.Description != nil {
		dto.Description = department.Description
	}
	if department.TenantID != nil {
		dto.TenantId = department.TenantID
	}
	if department.CreatedBy != nil {
		dto.CreatedBy = department.CreatedBy
	}
	if department.UpdatedBy != nil {
		dto.UpdatedBy = department.UpdatedBy
	}

	// 设置时间字段
	dto.CreatedAt = timeutil.TimeToTimestamppb(&department.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(&department.UpdatedAt)

	// 初始化Children字段
	dto.Children = make([]*userV1.Department, 0)

	return dto
}
