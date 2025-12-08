package gormcli

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	fileV1 "kratos-admin/api/gen/go/file/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type FileRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewFileRepo(db *gorm.DB, logger log.Logger) *FileRepo {
	return &FileRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "file/gormcli")),
	}
}

func (r *FileRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.File](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, fileV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *FileRepo) List(ctx context.Context, req *pagination.PagingRequest) (*fileV1.ListFileResponse, error) {
	if req == nil {
		return nil, fileV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderFile(req.GetOrderBy()),
		scopeFieldMaskFile(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.File](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
	total, err := r.Count(ctx, fScopes...)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, fileV1.ErrorInternalServerError("query count failed")
	}

	entities, err := g.Find(ctx)
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, fileV1.ErrorInternalServerError("query list failed")
	}

	dtos := make([]*fileV1.File, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toFileDTO(&entities[i]))
	}

	return &fileV1.ListFileResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *FileRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.File](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, fileV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *FileRepo) Get(ctx context.Context, req *fileV1.GetFileRequest) (*fileV1.File, error) {
	if req == nil {
		return nil, fileV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.File](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskFile(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fileV1.ErrorFileNotFound("file not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, fileV1.ErrorInternalServerError("query data failed")
	}

	return toFileDTO(&entity), nil
}

func (r *FileRepo) Create(ctx context.Context, req *fileV1.CreateFileRequest) error {
	if req == nil || req.Data == nil {
		return fileV1.ErrorBadRequest("invalid parameter")
	}

	entity := toFileModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.File](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return fileV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *FileRepo) Update(ctx context.Context, req *fileV1.UpdateFileRequest) error {
	if req == nil || req.Data == nil {
		return fileV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &fileV1.CreateFileRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch strings.ToLower(p) {
		case "provider":
			updates["provider"] = req.Data.GetProvider().String()
		case "bucket_name":
			updates["bucket_name"] = req.Data.GetBucketName()
		case "file_directory":
			updates["file_directory"] = req.Data.GetFileDirectory()
		case "file_guid":
			updates["file_guid"] = req.Data.GetFileGuid()
		case "save_file_name":
			updates["save_file_name"] = req.Data.GetSaveFileName()
		case "file_name":
			updates["file_name"] = req.Data.GetFileName()
		case "extension":
			updates["extension"] = req.Data.GetExtension()
		case "size":
			updates["size"] = req.Data.GetSize()
		case "size_format":
			updates["size_format"] = req.Data.GetSizeFormat()
		case "link_url":
			updates["link_url"] = req.Data.GetLinkUrl()
		case "md5":
			updates["md5"] = req.Data.GetMd5()
		case "updated_by":
			updates["updated_by"] = req.Data.GetUpdatedBy()
		case "updated_at":
			if req.Data.GetUpdatedAt() != nil {
				updates["updated_at"] = req.Data.GetUpdatedAt().AsTime()
			}
		}
	}

	if len(updates) == 0 {
		return nil
	}
	if _, ok := updates["updated_at"]; !ok {
		updates["updated_at"] = time.Now()
	}

	if err := r.db.WithContext(ctx).
		Model(&model.File{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return fileV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *FileRepo) Delete(ctx context.Context, req *fileV1.DeleteFileRequest) error {
	if req == nil {
		return fileV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).
		Where("id = ?", req.GetId()).
		Delete(&model.File{}).Error; err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
		return fileV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func toFileDTO(entity *model.File) *fileV1.File {
	dto := &fileV1.File{}
	if entity == nil {
		return dto
	}
	if entity.ID != 0 {
		dto.Id = toUint32Ptr(entity.ID)
	}
	if entity.Provider != "" {
		if v, ok := fileV1.OSSProvider_value[entity.Provider]; ok {
			p := fileV1.OSSProvider(v)
			dto.Provider = &p
		}
	}
	if entity.BucketName != "" {
		dto.BucketName = &entity.BucketName
	}
	if entity.FileDirectory != "" {
		dto.FileDirectory = &entity.FileDirectory
	}
	if entity.FileGUID != "" {
		dto.FileGuid = &entity.FileGUID
	}
	if entity.SaveFileName != "" {
		dto.SaveFileName = &entity.SaveFileName
	}
	if entity.FileName != "" {
		dto.FileName = &entity.FileName
	}
	if entity.Extension != "" {
		dto.Extension = &entity.Extension
	}
	if entity.Size != 0 {
		size := uint64(entity.Size)
		dto.Size = &size
	}
	if entity.SizeFormat != "" {
		dto.SizeFormat = &entity.SizeFormat
	}
	if entity.LinkURL != "" {
		dto.LinkUrl = &entity.LinkURL
	}
	if entity.Md5 != "" {
		dto.Md5 = &entity.Md5
	}
	dto.CreatedBy = toUint32Ptr(entity.CreatedBy)
	dto.UpdatedBy = toUint32Ptr(entity.UpdatedBy)
	dto.DeletedBy = toUint32Ptr(entity.DeletedBy)
	if !entity.CreatedAt.IsZero() {
		dto.CreatedAt = timestamppb.New(entity.CreatedAt)
	}
	if !entity.UpdatedAt.IsZero() {
		dto.UpdatedAt = timestamppb.New(entity.UpdatedAt)
	}
	if entity.DeletedAt.Valid && entity.DeletedAt.Time.Unix() > 0 {
		dto.DeletedAt = timestamppb.New(entity.DeletedAt.Time)
	}
	return dto
}

func toFileModel(dto *fileV1.File) *model.File {
	entity := &model.File{}
	if dto == nil {
		return entity
	}
	if dto.Id != nil {
		entity.ID = int64(dto.GetId())
	}
	if dto.Provider != nil {
		entity.Provider = dto.GetProvider().String()
	}
	if dto.BucketName != nil {
		entity.BucketName = dto.GetBucketName()
	}
	if dto.FileDirectory != nil {
		entity.FileDirectory = dto.GetFileDirectory()
	}
	if dto.FileGuid != nil {
		entity.FileGUID = dto.GetFileGuid()
	}
	if dto.SaveFileName != nil {
		entity.SaveFileName = dto.GetSaveFileName()
	}
	if dto.FileName != nil {
		entity.FileName = dto.GetFileName()
	}
	if dto.Extension != nil {
		entity.Extension = dto.GetExtension()
	}
	if dto.Size != nil {
		entity.Size = int64(dto.GetSize())
	}
	if dto.SizeFormat != nil {
		entity.SizeFormat = dto.GetSizeFormat()
	}
	if dto.LinkUrl != nil {
		entity.LinkURL = dto.GetLinkUrl()
	}
	if dto.Md5 != nil {
		entity.Md5 = dto.GetMd5()
	}
	entity.CreatedBy = int64(dto.GetCreatedBy())
	entity.UpdatedBy = int64(dto.GetUpdatedBy())
	entity.DeletedBy = int64(dto.GetDeletedBy())
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.GetCreatedAt().AsTime()
	}
	if dto.UpdatedAt != nil {
		entity.UpdatedAt = dto.GetUpdatedAt().AsTime()
	}
	if dto.DeletedAt != nil {
		entity.DeletedAt.Time = dto.GetDeletedAt().AsTime()
		entity.DeletedAt.Valid = true
	}
	return entity
}

var fileColumns = map[string]string{
	"id":             "id",
	"provider":       "provider",
	"bucket_name":    "bucket_name",
	"file_directory": "file_directory",
	"file_guid":      "file_guid",
	"save_file_name": "save_file_name",
	"file_name":      "file_name",
	"extension":      "extension",
	"size":           "size",
	"size_format":    "size_format",
	"link_url":       "link_url",
	"md5":            "md5",
	"created_by":     "created_by",
	"updated_by":     "updated_by",
	"created_at":     "created_at",
	"updated_at":     "updated_at",
}

func scopeOrderFile(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			col, ok := fileColumns[strings.ToLower(key)]
			if !ok {
				continue
			}
			cols = append(cols, clause.OrderByColumn{
				Column: clause.Column{Name: col},
				Desc:   desc,
			})
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskFile(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := fileColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}
