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

type DictEntryRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewDictEntryRepo(db *gorm.DB, logger log.Logger) *DictEntryRepo {
	return &DictEntryRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "dict-entry/repo/admin-service")),
	}
}

func (r *DictEntryRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.DictEntry{})

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

func (r *DictEntryRepo) List(ctx context.Context, req *pagination.PagingRequest) (*dictV1.ListDictEntryResponse, error) {
	if req == nil {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}

	var dictEntries []models.DictEntry
	var total int64

	query := r.db.WithContext(ctx).Model(&models.DictEntry{})

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

	if err := query.Find(&dictEntries).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*dictV1.DictEntry, 0, len(dictEntries))
	for _, dictEntry := range dictEntries {
		dto := r.toDTO(&dictEntry)
		dtos = append(dtos, dto)
	}

	return &dictV1.ListDictEntryResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *DictEntryRepo) Get(ctx context.Context, entryId uint32) (*dictV1.DictEntry, error) {
	if entryId == 0 {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}

	var dictEntry models.DictEntry
	if err := r.db.WithContext(ctx).First(&dictEntry, entryId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, dictV1.ErrorNotFound("dict entry not found")
		}
		r.log.Errorf("query dict entry failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&dictEntry), nil
}

func (r *DictEntryRepo) Create(ctx context.Context, req *dictV1.CreateDictEntryRequest) (*dictV1.DictEntry, error) {
	if req == nil || req.Data == nil {
		return nil, dictV1.ErrorBadRequest("invalid parameter")
	}

	dictEntry := r.fromCreateRequest(req)

	// 设置创建时间
	if req.Data.CreatedAt == nil {
		dictEntry.CreatedAt = time.Now()
	} else {
		createdAt := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		dictEntry.CreatedAt = *createdAt
	}

	if err := r.db.WithContext(ctx).Create(&dictEntry).Error; err != nil {
		r.log.Errorf("create dict entry failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(dictEntry), nil
}

func (r *DictEntryRepo) Update(ctx context.Context, req *dictV1.UpdateDictEntryRequest) error {
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

	if err := r.db.WithContext(ctx).Model(&models.DictEntry{}).Where("id = ?", req.Data.GetId()).Updates(updateData).Error; err != nil {
		r.log.Errorf("update dict entry failed: %s", err.Error())
		return dictV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *DictEntryRepo) Delete(ctx context.Context, entryId uint32) error {
	if err := r.db.WithContext(ctx).Delete(&models.DictEntry{}, entryId).Error; err != nil {
		r.log.Errorf("delete dict entry failed: %s", err.Error())
		return dictV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *DictEntryRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.DictEntry{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check dict entry exist failed: %s", err.Error())
		return false, dictV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *DictEntryRepo) GetDictEntryByDictTypeCode(ctx context.Context, dictTypeCode string) ([]*dictV1.DictEntry, error) {
	var dictEntries []models.DictEntry
	if err := r.db.WithContext(ctx).Where("type_code = ?", dictTypeCode).Find(&dictEntries).Error; err != nil {
		r.log.Errorf("query dict entries by type code failed: %s", err.Error())
		return nil, dictV1.ErrorInternalServerError("query data failed")
	}

	dtos := make([]*dictV1.DictEntry, 0, len(dictEntries))
	for _, dictEntry := range dictEntries {
		dto := r.toDTO(&dictEntry)
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

// 辅助方法
func (r *DictEntryRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})
	return conditions, nil
}

func (r *DictEntryRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	return nil
}

func (r *DictEntryRepo) buildUpdateData(data *dictV1.DictEntry, paths []string) map[string]interface{} {
	updateData := make(map[string]interface{})
	// Implementation for field mask updates
	return updateData
}

func (r *DictEntryRepo) buildUpdateDataFromRequest(data *dictV1.DictEntry) map[string]interface{} {
	updateData := make(map[string]interface{})
	if data.EntryLabel != nil {
		updateData["entry_label"] = *data.EntryLabel
	}
	if data.EntryValue != nil {
		updateData["entry_value"] = *data.EntryValue
	}
	if data.NumericValue != nil {
		updateData["numeric_value"] = *data.NumericValue
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
	if data.LanguageCode != nil {
		updateData["language_code"] = *data.LanguageCode
	}
	if data.UpdatedAt != nil {
		updateData["updated_at"] = timeutil.TimestamppbToTime(data.UpdatedAt)
	}
	if data.UpdatedBy != nil {
		updateData["update_by"] = *data.UpdatedBy
	}
	return updateData
}

func (r *DictEntryRepo) fromCreateRequest(req *dictV1.CreateDictEntryRequest) *models.DictEntry {
	dictEntry := &models.DictEntry{
		EntryLabel:   req.Data.EntryLabel,
		EntryValue:   req.Data.EntryValue,
		NumericValue: req.Data.NumericValue,
		LanguageCode: req.Data.LanguageCode,
		Description:  req.Data.Description,
	}

	if req.Data.TypeId != nil {
		dictEntry.TypeID = req.Data.TypeId
	}
	if req.Data.IsEnabled != nil {
		dictEntry.IsEnabled = req.Data.IsEnabled
	}
	if req.Data.SortOrder != nil {
		dictEntry.SortOrder = req.Data.SortOrder
	}
	if req.Data.CreatedBy != nil {
		dictEntry.CreatedBy = req.Data.CreatedBy
	}

	return dictEntry
}

func (r *DictEntryRepo) toDTO(dictEntry *models.DictEntry) *dictV1.DictEntry {
	dto := &dictV1.DictEntry{
		Id: &dictEntry.ID,
	}

	if dictEntry.TypeID != nil {
		dto.TypeId = dictEntry.TypeID
	}
	if dictEntry.EntryLabel != nil {
		dto.EntryLabel = dictEntry.EntryLabel
	}
	if dictEntry.EntryValue != nil {
		dto.EntryValue = dictEntry.EntryValue
	}
	if dictEntry.NumericValue != nil {
		dto.NumericValue = dictEntry.NumericValue
	}
	if dictEntry.LanguageCode != nil {
		dto.LanguageCode = dictEntry.LanguageCode
	}
	if dictEntry.IsEnabled != nil {
		dto.IsEnabled = dictEntry.IsEnabled
	}
	if dictEntry.SortOrder != nil {
		dto.SortOrder = dictEntry.SortOrder
	}
	if dictEntry.Description != nil {
		dto.Description = dictEntry.Description
	}
	if dictEntry.CreatedBy != nil {
		dto.CreatedBy = dictEntry.CreatedBy
	}
	if dictEntry.UpdatedBy != nil {
		dto.UpdatedBy = dictEntry.UpdatedBy
	}

	dto.CreatedAt = timeutil.TimeToTimestamppb(&dictEntry.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(&dictEntry.UpdatedAt)

	return dto
}
