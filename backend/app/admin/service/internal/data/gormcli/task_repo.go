package gormcli

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type TaskRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewTaskRepo(db *gorm.DB, logger log.Logger) *TaskRepo {
	return &TaskRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "task/gormcli")),
	}
}

func (r *TaskRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysTask](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *TaskRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListTaskResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderTask(req.GetOrderBy()),
		scopeFieldMaskTask(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysTask](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
	total, err := r.Count(ctx, fScopes...)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query count failed")
	}

	entities, err := g.Find(ctx)
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query list failed")
	}

	dtos := make([]*adminV1.Task, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toTaskDTO(&entities[i]))
	}

	return &adminV1.ListTaskResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *TaskRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysTask](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *TaskRepo) Get(ctx context.Context, req *adminV1.GetTaskRequest) (*adminV1.Task, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.SysTask](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskTask(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, adminV1.ErrorNotFound("task not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}
	return toTaskDTO(&entity), nil
}

func (r *TaskRepo) GetByTypeName(ctx context.Context, typeName string) (*adminV1.Task, error) {
	if strings.TrimSpace(typeName) == "" {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysTask](r.db.WithContext(ctx)).
		Where("type_name = ?", typeName).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, adminV1.ErrorNotFound("task not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}
	return toTaskDTO(&entity), nil
}

func (r *TaskRepo) Create(ctx context.Context, req *adminV1.CreateTaskRequest) (*adminV1.Task, error) {
	if req == nil || req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	entity := toTaskModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.SysTask](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("insert data failed")
	}
	return toTaskDTO(entity), nil
}

func (r *TaskRepo) Update(ctx context.Context, req *adminV1.UpdateTaskRequest) (*adminV1.Task, error) {
	if req == nil || req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return nil, err
		}
		if !exist {
			createReq := &adminV1.CreateTaskRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch strings.ToLower(p) {
		case "type":
			updates["type"] = req.Data.GetType().String()
		case "type_name":
			updates["type_name"] = req.Data.GetTypeName()
		case "task_payload":
			updates["task_payload"] = req.Data.GetTaskPayload()
		case "cron_spec":
			updates["cron_spec"] = req.Data.GetCronSpec()
		case "task_options":
			if req.Data.TaskOptions != nil {
				if buf, err := protojson.Marshal(req.Data.TaskOptions); err == nil {
					updates["task_options"] = string(buf)
				}
			} else {
				updates["task_options"] = ""
			}
		case "enable":
			updates["enable"] = req.Data.GetEnable()
		case "remark":
			updates["remark"] = req.Data.GetRemark()
		case "updated_by":
			updates["updated_by"] = req.Data.GetUpdatedBy()
		case "updated_at":
			if req.Data.GetUpdatedAt() != nil {
				updates["updated_at"] = req.Data.GetUpdatedAt().AsTime()
			}
		}
	}
	if len(updates) == 0 {
		return r.Get(ctx, &adminV1.GetTaskRequest{Id: req.GetData().GetId()})
	}
	if _, ok := updates["updated_at"]; !ok {
		updates["updated_at"] = time.Now()
	}

	if err := r.db.WithContext(ctx).
		Model(&model.SysTask{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("update data failed")
	}

	return r.Get(ctx, &adminV1.GetTaskRequest{Id: req.GetData().GetId()})
}

func (r *TaskRepo) Delete(ctx context.Context, req *adminV1.DeleteTaskRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).
		Where("id = ?", req.GetId()).
		Delete(&model.SysTask{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return adminV1.ErrorNotFound("task not found")
		}
		r.log.Errorf("delete one data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func toTaskDTO(entity *model.SysTask) *adminV1.Task {
	dto := &adminV1.Task{}
	if entity == nil {
		return dto
	}
	dto.Id = toUint32Ptr(entity.ID)
	if entity.Type != "" {
		if v, ok := adminV1.Task_Type_value[entity.Type]; ok {
			t := adminV1.Task_Type(v)
			dto.Type = &t
		}
	}
	if entity.TypeName != "" {
		dto.TypeName = &entity.TypeName
	}
	if entity.TaskPayload != "" {
		dto.TaskPayload = &entity.TaskPayload
	}
	if entity.CronSpec != "" {
		dto.CronSpec = &entity.CronSpec
	}
	if entity.TaskOptions != "" {
		opt := &adminV1.TaskOption{}
		if err := protojson.Unmarshal([]byte(entity.TaskOptions), opt); err == nil {
			dto.TaskOptions = opt
		}
	}
	dto.Enable = &entity.Enable
	if entity.Remark != "" {
		dto.Remark = &entity.Remark
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

func toTaskModel(dto *adminV1.Task) *model.SysTask {
	entity := &model.SysTask{}
	if dto == nil {
		return entity
	}
	if dto.Id != nil {
		entity.ID = int64(dto.GetId())
	}
	entity.Type = dto.GetType().String()
	if dto.TypeName != nil {
		entity.TypeName = dto.GetTypeName()
	}
	if dto.TaskPayload != nil {
		entity.TaskPayload = dto.GetTaskPayload()
	}
	if dto.CronSpec != nil {
		entity.CronSpec = dto.GetCronSpec()
	}
	if dto.TaskOptions != nil {
		if buf, err := protojson.Marshal(dto.TaskOptions); err == nil {
			entity.TaskOptions = string(buf)
		}
	}
	if dto.Enable != nil {
		entity.Enable = dto.GetEnable()
	}
	if dto.Remark != nil {
		entity.Remark = dto.GetRemark()
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

var taskColumns = map[string]string{
	"id":           "id",
	"type":         "type",
	"type_name":    "type_name",
	"task_payload": "task_payload",
	"cron_spec":    "cron_spec",
	"task_options": "task_options",
	"enable":       "enable",
	"remark":       "remark",
	"created_by":   "created_by",
	"updated_by":   "updated_by",
	"created_at":   "created_at",
	"updated_at":   "updated_at",
}

func scopeOrderTask(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			col, ok := taskColumns[strings.ToLower(key)]
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

func scopeFieldMaskTask(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := taskColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}
