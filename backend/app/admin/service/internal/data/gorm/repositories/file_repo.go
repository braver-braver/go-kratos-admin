package repositories

import (
	"context"
	"time"

	fileV1 "kratos-admin/api/gen/go/file/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gorm"
)

type FileRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewFileRepo(db *gorm.DB, logger log.Logger) *FileRepo {
	return &FileRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "file/repo/admin-service")),
	}
}

func (r *FileRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.File{})

	if err := query.Count(&count).Error; err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, fileV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *FileRepo) List(ctx context.Context, req *pagination.PagingRequest) (*fileV1.ListFileResponse, error) {
	if req == nil {
		return nil, fileV1.ErrorBadRequest("invalid parameter")
	}

	var files []models.File
	var total int64

	query := r.db.WithContext(ctx).Model(&models.File{})

	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, fileV1.ErrorInternalServerError("query count failed")
	}

	if !req.GetNoPaging() {
		offset := (req.GetPage() - 1) * req.GetPageSize()
		query = query.Offset(int(offset)).Limit(int(req.GetPageSize()))
	}

	if len(req.GetOrderBy()) > 0 {
		query = query.Order(req.GetOrderBy()[0])
	} else {
		query = query.Order("created_at DESC")
	}

	if req.GetFieldMask() != nil && len(req.GetFieldMask().GetPaths()) > 0 {
		query = query.Select(req.GetFieldMask().GetPaths())
	}

	if err := query.Find(&files).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, fileV1.ErrorInternalServerError("query list failed")
	}

	dtos := make([]*fileV1.File, 0, len(files))
	for _, file := range files {
		dto := r.toDTO(&file)
		dtos = append(dtos, dto)
	}

	return &fileV1.ListFileResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *FileRepo) Get(ctx context.Context, fileId uint32) (*fileV1.File, error) {
	if fileId == 0 {
		return nil, fileV1.ErrorBadRequest("invalid parameter")
	}

	var file models.File
	if err := r.db.WithContext(ctx).First(&file, fileId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fileV1.ErrorNotFound("file not found")
		}
		r.log.Errorf("query file failed: %s", err.Error())
		return nil, fileV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&file), nil
}

func (r *FileRepo) Create(ctx context.Context, req *fileV1.CreateFileRequest) (*fileV1.File, error) {
	if req == nil || req.Data == nil {
		return nil, fileV1.ErrorBadRequest("invalid parameter")
	}

	file := r.fromCreateRequest(req)

	if req.Data.CreatedAt == nil {
		file.CreatedAt = time.Now()
	} else {
		createdAt := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		file.CreatedAt = *createdAt
	}

	if err := r.db.WithContext(ctx).Create(&file).Error; err != nil {
		r.log.Errorf("create file failed: %s", err.Error())
		return nil, fileV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(&file), nil
}

func (r *FileRepo) Update(ctx context.Context, req *fileV1.UpdateFileRequest) error {
	if req == nil || req.Data == nil {
		return fileV1.ErrorBadRequest("invalid parameter")
	}

	updateData := r.buildUpdateDataFromRequest(req.Data)

	updateData["updated_at"] = time.Now()

	if err := r.db.WithContext(ctx).Model(&models.File{}).Where("id = ?", req.Data.GetId()).Updates(updateData).Error; err != nil {
		r.log.Errorf("update file failed: %s", err.Error())
		return fileV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

func (r *FileRepo) Delete(ctx context.Context, fileId uint32) error {
	if err := r.db.WithContext(ctx).Delete(&models.File{}, fileId).Error; err != nil {
		r.log.Errorf("delete file failed: %s", err.Error())
		return fileV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *FileRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.File{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check file exist failed: %s", err.Error())
		return false, fileV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *FileRepo) buildUpdateDataFromRequest(data *fileV1.File) map[string]interface{} {
	updateData := make(map[string]interface{})

	updateData["name"] = data.GetFileName()
	updateData["path"] = data.GetFileDirectory()
	if data.Size != nil && data.GetSize() > 0 {
		updateData["size"] = toInt64PtrValue(data.GetSize())
	}
	updateData["type"] = data.Extension
	// proto lacks explicit mime/remark fields; best-effort placeholders
	updateData["mime_type"] = nil
	updateData["remark"] = data.FileGuid
	updateData["create_by"] = data.CreatedBy
	updateData["update_by"] = data.UpdatedBy

	return updateData
}

func (r *FileRepo) fromCreateRequest(req *fileV1.CreateFileRequest) models.File {
	return models.File{
		Name:      req.Data.FileName,
		Path:      req.Data.FileDirectory,
		Size:      toInt64PtrValue(req.Data.GetSize()),
		Type:      req.Data.Extension,
		MimeType:  nil,
		Remark:    req.Data.FileGuid,
		CreatedBy: req.Data.CreatedBy,
		UpdatedBy: req.Data.UpdatedBy,
	}
}

func (r *FileRepo) toDTO(file *models.File) *fileV1.File {
	if file == nil {
		return nil
	}

	id := uint32(file.ID)
	size := toUint64Ptr(file.Size)
	return &fileV1.File{
		Id:            &id,
		FileName:      file.Name,
		FileDirectory: file.Path,
		Size:          size,
		Extension:     file.Type,
		CreatedBy:     file.CreatedBy,
		UpdatedBy:     file.UpdatedBy,
		CreatedAt:     timeutil.TimeToTimestamppb(&file.CreatedAt),
		UpdatedAt:     timeutil.TimeToTimestamppb(&file.UpdatedAt),
	}
}

func toInt64PtrValue(u uint64) *int64 {
	if u == 0 {
		return nil
	}
	v := int64(u)
	return &v
}

func toUint64Ptr(i *int64) *uint64 {
	if i == nil {
		return nil
	}
	v := uint64(*i)
	return &v
}
