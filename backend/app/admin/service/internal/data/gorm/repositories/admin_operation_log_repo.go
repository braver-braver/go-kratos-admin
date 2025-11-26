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

type AdminOperationLogRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewAdminOperationLogRepo(db *gorm.DB, logger log.Logger) *AdminOperationLogRepo {
	return &AdminOperationLogRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "admin-operation-log/repo/admin-service")),
	}
}

// Count 统计管理员操作日志数量
func (r *AdminOperationLogRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.AdminOperationLog{})

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

// List 获取管理员操作日志列表
func (r *AdminOperationLogRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListAdminOperationLogResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	var logs []models.AdminOperationLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AdminOperationLog{})

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

	if err := query.Find(&logs).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*adminV1.AdminOperationLog, 0, len(logs))
	for _, log := range logs {
		dto := r.toDTO(&log)
		dtos = append(dtos, dto)
	}

	return &adminV1.ListAdminOperationLogResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

// Get 根据ID获取管理员操作日志
func (r *AdminOperationLogRepo) Get(ctx context.Context, logId uint32) (*adminV1.AdminOperationLog, error) {
	if logId == 0 {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	var log models.AdminOperationLog
	if err := r.db.WithContext(ctx).First(&log, logId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, adminV1.ErrorNotFound("admin operation log not found")
		}
		r.log.Errorf("query log failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&log), nil
}

// Create 创建管理员操作日志
func (r *AdminOperationLogRepo) Create(ctx context.Context, req *adminV1.CreateAdminOperationLogRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	log := r.fromCreateRequest(req)

	// 设置创建时间
	if req.Data.CreatedAt == nil {
		log.CreatedAt = time.Now()
	} else {
		createdAt := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		log.CreatedAt = *createdAt
	}

	if err := r.db.WithContext(ctx).Create(&log).Error; err != nil {
		r.log.Errorf("create admin operation log failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

// IsExist 检查管理员操作日志是否存在
func (r *AdminOperationLogRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.AdminOperationLog{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check log exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

// 辅助方法

// buildConditions 构建查询条件
func (r *AdminOperationLogRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
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
func (r *AdminOperationLogRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		switch key {
		case "username":
			query = query.Where("username LIKE ?", "%"+value.(string)+"%")
		case "user_id":
			query = query.Where("user_id = ?", value)
		case "method":
			query = query.Where("method = ?", value)
		case "status_code":
			query = query.Where("status_code = ?", value)
		default:
			query = query.Where(key+" = ?", value)
		}
	}
	return nil
}

// fromCreateRequest 从创建请求构建模型
func (r *AdminOperationLogRepo) fromCreateRequest(req *adminV1.CreateAdminOperationLogRequest) *models.AdminOperationLog {
	log := &models.AdminOperationLog{}

	if req.Data.RequestId != nil {
		log.RequestId = req.Data.RequestId
	}
	if req.Data.Method != nil {
		log.Method = req.Data.Method
	}
	if req.Data.Operation != nil {
		log.Operation = req.Data.Operation
	}
	if req.Data.Path != nil {
		log.Path = req.Data.Path
	}
	if req.Data.Referer != nil {
		log.Referer = req.Data.Referer
	}
	if req.Data.RequestUri != nil {
		log.RequestUri = req.Data.RequestUri
	}
	if req.Data.RequestBody != nil {
		log.RequestBody = req.Data.RequestBody
	}
	if req.Data.RequestHeader != nil {
		log.RequestHeader = req.Data.RequestHeader
	}
	if req.Data.Response != nil {
		log.Response = req.Data.Response
	}
	if req.Data.CostTime != nil {
		totalSeconds := float64(req.Data.CostTime.Seconds) + float64(req.Data.CostTime.Nanos)/1e9
		log.CostTime = &totalSeconds
	}
	if req.Data.UserId != nil {
		log.UserID = req.Data.UserId
	}
	if req.Data.Username != nil {
		log.Username = req.Data.Username
	}
	if req.Data.ClientIp != nil {
		log.ClientIp = req.Data.ClientIp
	}
	if req.Data.UserAgent != nil {
		log.UserAgent = req.Data.UserAgent
	}
	if req.Data.BrowserName != nil {
		log.BrowserName = req.Data.BrowserName
	}
	if req.Data.BrowserVersion != nil {
		log.BrowserVersion = req.Data.BrowserVersion
	}
	if req.Data.ClientId != nil {
		log.ClientID = req.Data.ClientId
	}
	if req.Data.ClientName != nil {
		log.ClientName = req.Data.ClientName
	}
	if req.Data.OsName != nil {
		log.OsName = req.Data.OsName
	}
	if req.Data.OsVersion != nil {
		log.OsVersion = req.Data.OsVersion
	}
	if req.Data.StatusCode != nil {
		log.StatusCode = req.Data.StatusCode
	}
	if req.Data.Success != nil {
		log.Success = req.Data.Success
	}
	if req.Data.Reason != nil {
		log.Reason = req.Data.Reason
	}
	if req.Data.Location != nil {
		log.Location = req.Data.Location
	}

	return log
}

// toDTO 将模型转换为 DTO
func (r *AdminOperationLogRepo) toDTO(log *models.AdminOperationLog) *adminV1.AdminOperationLog {
	dto := &adminV1.AdminOperationLog{
		Id: &log.ID,
	}

	if log.RequestId != nil {
		dto.RequestId = log.RequestId
	}
	if log.Method != nil {
		dto.Method = log.Method
	}
	if log.Operation != nil {
		dto.Operation = log.Operation
	}
	if log.Path != nil {
		dto.Path = log.Path
	}
	if log.Referer != nil {
		dto.Referer = log.Referer
	}
	if log.RequestUri != nil {
		dto.RequestUri = log.RequestUri
	}
	if log.RequestBody != nil {
		dto.RequestBody = log.RequestBody
	}
	if log.RequestHeader != nil {
		dto.RequestHeader = log.RequestHeader
	}
	if log.Response != nil {
		dto.Response = log.Response
	}
	if log.CostTime != nil {
		dto.CostTime = timeutil.SecondToDurationpb(log.CostTime)
	}
	if log.UserID != nil {
		dto.UserId = log.UserID
	}
	if log.Username != nil {
		dto.Username = log.Username
	}
	if log.ClientIp != nil {
		dto.ClientIp = log.ClientIp
	}
	if log.UserAgent != nil {
		dto.UserAgent = log.UserAgent
	}
	if log.BrowserName != nil {
		dto.BrowserName = log.BrowserName
	}
	if log.BrowserVersion != nil {
		dto.BrowserVersion = log.BrowserVersion
	}
	if log.ClientID != nil {
		dto.ClientId = log.ClientID
	}
	if log.ClientName != nil {
		dto.ClientName = log.ClientName
	}
	if log.OsName != nil {
		dto.OsName = log.OsName
	}
	if log.OsVersion != nil {
		dto.OsVersion = log.OsVersion
	}
	if log.StatusCode != nil {
		dto.StatusCode = log.StatusCode
	}
	if log.Success != nil {
		dto.Success = log.Success
	}
	if log.Reason != nil {
		dto.Reason = log.Reason
	}
	if log.Location != nil {
		dto.Location = log.Location
	}

	// 设置时间字段
	dto.CreatedAt = timeutil.TimeToTimestamppb(&log.CreatedAt)

	return dto
}
