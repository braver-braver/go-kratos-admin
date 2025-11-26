package repositories

import (
	"context"
	"time"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gorm"
)

type ApiResourceRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewApiResourceRepo(db *gorm.DB, logger log.Logger) *ApiResourceRepo {
	return &ApiResourceRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "api-resource/repo/admin-service")),
	}
}

func (r *ApiResourceRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.ApiResource{})

	// 应用查询条件
	if err := r.applyConditions(query, conditions); err != nil {
		return 0, err
	}

	if err := query.Count(&count).Error; err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *ApiResourceRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListApiResourceResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	var resources []models.ApiResource
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ApiResource{})

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
		return nil, adminV1.ErrorInternalServerError("query count failed")
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

	if err := query.Find(&resources).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*adminV1.ApiResource, 0, len(resources))
	for _, resource := range resources {
		dto := r.toDTO(&resource)
		dtos = append(dtos, dto)
	}

	return &adminV1.ListApiResourceResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *ApiResourceRepo) Get(ctx context.Context, resourceId uint32) (*adminV1.ApiResource, error) {
	if resourceId == 0 {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	var resource models.ApiResource
	if err := r.db.WithContext(ctx).First(&resource, resourceId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, adminV1.ErrorNotFound("api resource not found")
		}
		r.log.Errorf("query resource failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&resource), nil
}

func (r *ApiResourceRepo) Create(ctx context.Context, req *adminV1.CreateApiResourceRequest) (*adminV1.ApiResource, error) {
	if req == nil || req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	resource := r.fromCreateRequest(req)

	// 设置创建时间
	if req.Data.CreatedAt == nil {
		resource.CreatedAt = time.Now()
	} else {
		createdAt := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		resource.CreatedAt = *createdAt
	}

	if err := r.db.WithContext(ctx).Create(&resource).Error; err != nil {
		r.log.Errorf("create api resource failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(resource), nil
}

func (r *ApiResourceRepo) Update(ctx context.Context, req *adminV1.UpdateApiResourceRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	// 处理字段掩码
	updateData := make(map[string]interface{})
	if req.UpdateMask != nil {
		req.UpdateMask.Normalize()
		if !req.UpdateMask.IsValid(req.Data) {
			r.log.Errorf("invalid field mask [%v]", req.UpdateMask)
			return adminV1.ErrorBadRequest("invalid field mask")
		}

		// 根据字段掩码构建更新数据
		updateData = r.buildUpdateData(req.Data, req.UpdateMask.GetPaths())
	} else {
		// 更新所有字段
		updateData = r.buildUpdateDataFromRequest(req.Data)
	}

	// 设置更新时间
	updateData["updated_at"] = time.Now()

	if err := r.db.WithContext(ctx).Model(&models.ApiResource{}).Where("id = ?", req.Data.GetId()).Updates(updateData).Error; err != nil {
		r.log.Errorf("update api resource failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *ApiResourceRepo) Delete(ctx context.Context, resourceId uint32) error {
	if err := r.db.WithContext(ctx).Delete(&models.ApiResource{}, resourceId).Error; err != nil {
		r.log.Errorf("delete api resource failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *ApiResourceRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.ApiResource{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check resource exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *ApiResourceRepo) GetResourceByPathAndMethod(ctx context.Context, path string, method string) (*adminV1.ApiResource, error) {
	var resource models.ApiResource
	if err := r.db.WithContext(ctx).Where("path = ? AND method = ?", path, method).First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, adminV1.ErrorNotFound("api resource not found")
		}
		r.log.Errorf("query resource by path and method failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&resource), nil
}

// 辅助方法
func (r *ApiResourceRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})
	return conditions, nil
}

func (r *ApiResourceRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	return nil
}

func (r *ApiResourceRepo) buildUpdateData(data *adminV1.ApiResource, paths []string) map[string]interface{} {
	updateData := make(map[string]interface{})
	// Implementation for field mask updates
	return updateData
}

func (r *ApiResourceRepo) buildUpdateDataFromRequest(data *adminV1.ApiResource) map[string]interface{} {
	updateData := make(map[string]interface{})
	if data.Operation != nil {
		updateData["operation"] = *data.Operation
	}
	if data.Path != nil {
		updateData["path"] = *data.Path
	}
	if data.Method != nil {
		updateData["method"] = *data.Method
	}
	if data.Scope != nil {
		updateData["scope"] = int32(*data.Scope)
	}
	if data.Description != nil {
		updateData["description"] = *data.Description
	}
	if data.UpdatedAt != nil {
		updateData["updated_at"] = timeutil.TimestamppbToTime(data.UpdatedAt)
	}
	if data.UpdatedBy != nil {
		updateData["update_by"] = *data.UpdatedBy
	}
	return updateData
}

func (r *ApiResourceRepo) fromCreateRequest(req *adminV1.CreateApiResourceRequest) *models.ApiResource {
	resource := &models.ApiResource{
		Operation: req.Data.Operation,
		Path:      req.Data.Path,
		Method:    req.Data.Method,
		Desc:      req.Data.Description,
	}

	if req.Data.Scope != nil {
		scope := int32(*req.Data.Scope)
		resource.Scope = &scope
	}
	if req.Data.CreatedBy != nil {
		resource.CreatedBy = req.Data.CreatedBy
	}

	return resource
}

func (r *ApiResourceRepo) toDTO(resource *models.ApiResource) *adminV1.ApiResource {
	dto := &adminV1.ApiResource{
		Id: &resource.ID,
	}

	if resource.Operation != nil {
		dto.Operation = resource.Operation
	}
	if resource.Path != nil {
		dto.Path = resource.Path
	}
	if resource.Method != nil {
		dto.Method = resource.Method
	}
	if resource.Scope != nil {
		scope := adminV1.ApiResource_Scope(*resource.Scope)
		dto.Scope = &scope
	}
	if resource.Desc != nil {
		dto.Description = resource.Desc
	}
	if resource.CreatedBy != nil {
		dto.CreatedBy = resource.CreatedBy
	}
	if resource.UpdatedBy != nil {
		dto.UpdatedBy = resource.UpdatedBy
	}

	dto.CreatedAt = timeutil.TimeToTimestamppb(&resource.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(&resource.UpdatedAt)

	return dto
}
