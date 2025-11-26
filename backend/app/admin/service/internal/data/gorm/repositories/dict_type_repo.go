package repositories

import (
	"context"
	"time"

	dictV1 "kratos-admin/api/gen/go/dict/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gorm"
)

type DictTypeRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewDictTypeRepo(db *gorm.DB, logger log.Logger) *DictTypeRepo {
	return &DictTypeRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "dict-type/repo/admin-service")),
	}
}

func (r *DictTypeRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.DictType{})

	// 应用查询条件
	if err := r.applyConditions(query, conditions); err != nil {
		return 0, err
	}

	if err := query.Count(&count).Error; err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, dictV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *DictTypeRepo) List(ctx context.Context, req *pagination.PagingRequest) (*dictV1.ListDictTypeResponse, error) {
	if req == nil {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}

	var dictTypes []models.DictType
	var total int64

	query := r.db.WithContext(ctx).Model(&models.DictType{})

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
		return nil, dictV1.ErrorInternalServerError("query count failed")
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

	if err := query.Find(&dictTypes).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*dictV1.DictType, 0, len(dictTypes))
	for _, dictType := range dictTypes {
		dto := r.toDTO(&dictType)
		dtos = append(dtos, dto)
	}

	return &dictV1.ListDictTypeResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *DictTypeRepo) Get(ctx context.Context, typeId uint32) (*dictV1.DictType, error) {
	if typeId == 0 {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}

	var dictType models.DictType
	if err := r.db.WithContext(ctx).First(&dictType, typeId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, dictV1.ErrorNotFound("dict type not found")
		}
		r.log.Errorf("query dict type failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&dictType), nil
}

func (r *DictTypeRepo) Create(ctx context.Context, req *dictV1.CreateDictTypeRequest) (*dictV1.DictType, error) {
	if req == nil || req.Data == nil {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}

	dictType := r.fromCreateRequest(req)

	// 设置创建时间
	if req.Data.CreatedAt == nil {
		dictType.CreatedAt = time.Now()
	} else {
		createdAt := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		dictType.CreatedAt = *createdAt
	}

	if err := r.db.WithContext(ctx).Create(&dictType).Error; err != nil {
		r.log.Errorf("create dict type failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(dictType), nil
}

func (r *DictTypeRepo) Update(ctx context.Context, req *dictV1.UpdateDictTypeRequest) error {
	if req == nil || req.Data == nil {
		return dictV1.ErrorBadRequest("invalid parameter")
	}

	// 处理字段掩码
	updateData := make(map[string]interface{})
	if req.UpdateMask != nil {
		req.UpdateMask.Normalize()
		if !req.UpdateMask.IsValid(req.Data) {
			r.log.Errorf("invalid field mask [%v]", req.UpdateMask)
			return dictV1.ErrorBadRequest("invalid field mask")
		}

		// 根据字段掩码构建更新数据
		updateData = r.buildUpdateData(req.Data, req.UpdateMask.GetPaths())
	} else {
		// 更新所有字段
		updateData = r.buildUpdateDataFromRequest(req.Data)
	}

	// 设置更新时间
	updateData["updated_at"] = time.Now()

	if err := r.db.WithContext(ctx).Model(&models.DictType{}).Where("id = ?", req.Data.GetId()).Updates(updateData).Error; err != nil {
		r.log.Errorf("update dict type failed: %s", err.Error())
		return dictV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *DictTypeRepo) Delete(ctx context.Context, typeId uint32) error {
	if err := r.db.WithContext(ctx).Delete(&models.DictType{}, typeId).Error; err != nil {
		r.log.Errorf("delete dict type failed: %s", err.Error())
		return dictV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *DictTypeRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.DictType{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check dict type exist failed: %s", err.Error())
		return false, dictV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *DictTypeRepo) GetDictTypeByDictTypeCode(ctx context.Context, dictTypeCode string) (*dictV1.DictType, error) {
	var dictType models.DictType
	if err := r.db.WithContext(ctx).Where("code = ?", dictTypeCode).First(&dictType).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, dictV1.ErrorNotFound("dict type not found")
		}
		r.log.Errorf("query dict type by code failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&dictType), nil
}

// 辅助方法
func (r *DictTypeRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})
	return conditions, nil
}

func (r *DictTypeRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	return nil
}

func (r *DictTypeRepo) buildUpdateData(data *dictV1.DictType, paths []string) map[string]interface{} {
	updateData := make(map[string]interface{})
	// Implementation for field mask updates
	return updateData
}

func (r *DictTypeRepo) buildUpdateDataFromRequest(data *dictV1.DictType) map[string]interface{} {
	updateData := make(map[string]interface{})
	if data.TypeName != nil {
		updateData["type_name"] = *data.TypeName
	}
	if data.TypeCode != nil {
		updateData["type_code"] = *data.TypeCode
	}
	if data.IsEnabled != nil {
		updateData["is_enabled"] = *data.IsEnabled
	}
	if data.Description != nil {
		updateData["description"] = *data.Description
	}
	if data.SortOrder != nil {
		updateData["sort_order"] = *data.SortOrder
	}
	if data.UpdatedAt != nil {
		updateData["updated_at"] = timeutil.TimestamppbToTime(data.UpdatedAt)
	}
	if data.UpdatedBy != nil {
		updateData["update_by"] = *data.UpdatedBy
	}
	return updateData
}

func (r *DictTypeRepo) fromCreateRequest(req *dictV1.CreateDictTypeRequest) *models.DictType {
	dictType := &models.DictType{
		TypeName:    req.Data.TypeName,
		TypeCode:    req.Data.TypeCode,
		Description: req.Data.Description,
	}

	if req.Data.IsEnabled != nil {
		dictType.IsEnabled = req.Data.IsEnabled
	}
	if req.Data.SortOrder != nil {
		dictType.SortOrder = req.Data.SortOrder
	}
	if req.Data.CreatedBy != nil {
		dictType.CreatedBy = req.Data.CreatedBy
	}

	return dictType
}

func (r *DictTypeRepo) toDTO(dictType *models.DictType) *dictV1.DictType {
	dto := &dictV1.DictType{
		Id: &dictType.ID,
	}

	if dictType.TypeName != nil {
		dto.TypeName = dictType.TypeName
	}
	if dictType.TypeCode != nil {
		dto.TypeCode = dictType.TypeCode
	}
	if dictType.IsEnabled != nil {
		dto.IsEnabled = dictType.IsEnabled
	}
	if dictType.SortOrder != nil {
		dto.SortOrder = dictType.SortOrder
	}
	if dictType.Description != nil {
		dto.Description = dictType.Description
	}
	if dictType.CreatedBy != nil {
		dto.CreatedBy = dictType.CreatedBy
	}
	if dictType.UpdatedBy != nil {
		dto.UpdatedBy = dictType.UpdatedBy
	}

	dto.CreatedAt = timeutil.TimeToTimestamppb(&dictType.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(&dictType.UpdatedAt)

	return dto
}
