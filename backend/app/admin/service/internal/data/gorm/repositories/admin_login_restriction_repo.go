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

type AdminLoginRestrictionRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewAdminLoginRestrictionRepo(db *gorm.DB, logger log.Logger) *AdminLoginRestrictionRepo {
	return &AdminLoginRestrictionRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "admin-login-restriction/repo/admin-service")),
	}
}

func (r *AdminLoginRestrictionRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.AdminLoginRestriction{})

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

func (r *AdminLoginRestrictionRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListAdminLoginRestrictionResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	var restrictions []models.AdminLoginRestriction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AdminLoginRestriction{})

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

	if err := query.Find(&restrictions).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*adminV1.AdminLoginRestriction, 0, len(restrictions))
	for _, restriction := range restrictions {
		dto := r.toDTO(&restriction)
		dtos = append(dtos, dto)
	}

	return &adminV1.ListAdminLoginRestrictionResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *AdminLoginRestrictionRepo) Get(ctx context.Context, restrictionId uint32) (*adminV1.AdminLoginRestriction, error) {
	if restrictionId == 0 {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	var restriction models.AdminLoginRestriction
	if err := r.db.WithContext(ctx).First(&restriction, restrictionId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, adminV1.ErrorNotFound("admin login restriction not found")
		}
		r.log.Errorf("query restriction failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&restriction), nil
}

func (r *AdminLoginRestrictionRepo) Create(ctx context.Context, req *adminV1.CreateAdminLoginRestrictionRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	restriction := r.fromCreateRequest(req)

	// 设置创建时间
	if req.Data.CreatedAt == nil {
		restriction.CreatedAt = time.Now()
	} else {
		createdAt := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		restriction.CreatedAt = *createdAt
	}

	if err := r.db.WithContext(ctx).Create(&restriction).Error; err != nil {
		r.log.Errorf("create admin login restriction failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *AdminLoginRestrictionRepo) Update(ctx context.Context, req *adminV1.UpdateAdminLoginRestrictionRequest) error {
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

	if err := r.db.WithContext(ctx).Model(&models.AdminLoginRestriction{}).Where("id = ?", req.Data.GetId()).Updates(updateData).Error; err != nil {
		r.log.Errorf("update admin login restriction failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *AdminLoginRestrictionRepo) Delete(ctx context.Context, restrictionId uint32) error {
	if err := r.db.WithContext(ctx).Delete(&models.AdminLoginRestriction{}, restrictionId).Error; err != nil {
		r.log.Errorf("delete admin login restriction failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *AdminLoginRestrictionRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.AdminLoginRestriction{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check restriction exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

// 辅助方法

func (r *AdminLoginRestrictionRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})
	return conditions, nil
}

func (r *AdminLoginRestrictionRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	return nil
}

func (r *AdminLoginRestrictionRepo) buildUpdateData(data *adminV1.AdminLoginRestriction, paths []string) map[string]interface{} {
	updateData := make(map[string]interface{})
	// Implementation for field mask updates
	return updateData
}

func (r *AdminLoginRestrictionRepo) buildUpdateDataFromRequest(data *adminV1.AdminLoginRestriction) map[string]interface{} {
	updateData := make(map[string]interface{})
	if data.Reason != nil {
		updateData["reason"] = *data.Reason
	}
	if data.UpdatedAt != nil {
		updateData["updated_at"] = timeutil.TimestamppbToTime(data.UpdatedAt)
	}
	if data.UpdatedBy != nil {
		updateData["update_by"] = *data.UpdatedBy
	}
	return updateData
}

func (r *AdminLoginRestrictionRepo) fromCreateRequest(req *adminV1.CreateAdminLoginRestrictionRequest) *models.AdminLoginRestriction {
	restriction := &models.AdminLoginRestriction{
		TargetID: req.Data.TargetId,
		Value:    req.Data.Value,
		Reason:   req.Data.Reason,
	}

	if req.Data.CreatedBy != nil {
		restriction.CreatedBy = req.Data.CreatedBy
	}

	return restriction
}

func (r *AdminLoginRestrictionRepo) toDTO(restriction *models.AdminLoginRestriction) *adminV1.AdminLoginRestriction {
	dto := &adminV1.AdminLoginRestriction{
		Id: &restriction.ID,
	}

	if restriction.TargetID != nil {
		dto.TargetId = restriction.TargetID
	}
	if restriction.Value != nil {
		dto.Value = restriction.Value
	}
	if restriction.Reason != nil {
		dto.Reason = restriction.Reason
	}
	if restriction.CreatedBy != nil {
		dto.CreatedBy = restriction.CreatedBy
	}

	dto.CreatedAt = timeutil.TimeToTimestamppb(&restriction.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(&restriction.UpdatedAt)

	return dto
}
