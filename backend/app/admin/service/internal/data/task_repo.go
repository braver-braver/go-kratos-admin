package data

import (
	"context"
	"encoding/json"
	"time"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/model"
	"kratos-admin/app/admin/service/internal/data/gorm/query"
	"kratos-admin/pkg/datautil"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

type TaskRepo struct {
	log *log.Helper
	q   *query.Query
}

func NewTaskRepo(data *Data, logger log.Logger) *TaskRepo {
	return &TaskRepo{
		log: log.NewHelper(log.With(logger, "module", "task/repo/admin-service")),
		q:   query.Use(data.db),
	}
}

func (r *TaskRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.q.SysTask.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}
	return int(count), nil
}

func (r *TaskRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListTaskResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.q.SysTask.WithContext(ctx).Order(r.q.SysTask.CreatedAt.Desc())

	if !req.GetNoPaging() {
		ps := int(req.GetPageSize())
		if ps <= 0 {
			ps = 10
		}
		offset := int(req.GetPage()-1) * ps
		if offset < 0 {
			offset = 0
		}
		builder = builder.Offset(offset).Limit(ps)
	}

	entities, err := builder.Find()
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query list failed")
	}

	total, err := r.q.SysTask.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query count failed")
	}

	items := make([]*adminV1.Task, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}

	return &adminV1.ListTaskResponse{
		Total: uint32(total),
		Items: items,
	}, nil
}

func (r *TaskRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	count, err := r.q.SysTask.WithContext(ctx).
		Where(r.q.SysTask.ID.Eq(int32(id))).
		Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *TaskRepo) Get(ctx context.Context, id uint32) (*adminV1.Task, error) {
	entity, err := r.q.SysTask.WithContext(ctx).
		Where(r.q.SysTask.ID.Eq(int32(id))).
		First()
	if err != nil {
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorNotFound("task not found")
	}
	return r.toDTO(entity), nil
}

func (r *TaskRepo) GetByTypeName(ctx context.Context, typeName string) (*adminV1.Task, error) {
	entity, err := r.q.SysTask.WithContext(ctx).
		Where(r.q.SysTask.TypeName.Eq(typeName)).
		First()
	if err != nil {
		return nil, adminV1.ErrorNotFound("task not found")
	}
	return r.toDTO(entity), nil
}

func (r *TaskRepo) Create(ctx context.Context, req *adminV1.CreateTaskRequest) (*adminV1.Task, error) {
	if req == nil || req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	now := time.Now()
	entity := &model.SysTask{
		CreatedAt:   &now,
		UpdatedAt:   &now,
		CreateBy:    cloneInt64FromUint32(req.Data.CreatedBy),
		UpdateBy:    cloneInt64FromUint32(req.Data.UpdatedBy),
		Remark:      cloneStringPtr(req.Data.Remark),
		Type:        taskTypeToString(req.Data.Type),
		TypeName:    cloneStringPtr(req.Data.TypeName),
		TaskPayload: cloneStringPtr(req.Data.TaskPayload),
		CronSpec:    cloneStringPtr(req.Data.CronSpec),
		TaskOption:  encodeTaskOption(req.Data.TaskOptions),
		Enable:      cloneBoolPtr(req.Data.Enable),
	}

	if err := r.q.SysTask.WithContext(ctx).Create(entity); err != nil {
		r.log.Errorf("insert data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("insert data failed")
	}

	return r.toDTO(entity), nil
}

func (r *TaskRepo) Update(ctx context.Context, req *adminV1.UpdateTaskRequest) (*adminV1.Task, error) {
	if req == nil || req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	update := map[string]any{
		"updated_at": time.Now(),
	}
	if req.Data.Type != nil {
		update["type"] = req.Data.GetType().String()
	}
	if req.Data.TypeName != nil {
		update["type_name"] = req.Data.GetTypeName()
	}
	if req.Data.TaskPayload != nil {
		update["task_payload"] = req.Data.GetTaskPayload()
	}
	if req.Data.CronSpec != nil {
		update["cron_spec"] = req.Data.GetCronSpec()
	}
	if req.Data.TaskOptions != nil {
		update["task_options"] = encodeTaskOption(req.Data.TaskOptions)
	}
	if req.Data.Enable != nil {
		update["enable"] = req.Data.GetEnable()
	}
	if req.Data.UpdatedBy != nil {
		update["update_by"] = req.Data.GetUpdatedBy()
	}
	if req.Data.Remark != nil {
		update["remark"] = req.Data.GetRemark()
	}

	_, err := r.q.SysTask.WithContext(ctx).
		Where(r.q.SysTask.ID.Eq(int32(req.Data.GetId()))).
		Updates(update)
	if err != nil {
		r.log.Errorf("update data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("update data failed")
	}

	return r.Get(ctx, req.Data.GetId())
}

func (r *TaskRepo) Delete(ctx context.Context, req *adminV1.DeleteTaskRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	_, err := r.q.SysTask.WithContext(ctx).
		Where(r.q.SysTask.ID.Eq(int32(req.GetId()))).
		Delete()
	if err != nil {
		r.log.Errorf("delete data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete data failed")
	}
	return nil
}

func (r *TaskRepo) toDTO(entity *model.SysTask) *adminV1.Task {
	if entity == nil {
		return nil
	}
	return &adminV1.Task{
		Id:          datautil.CloneUint32(uint32(entity.ID)),
		Type:        stringToTaskType(entity.Type),
		TypeName:    cloneStringPtr(entity.TypeName),
		TaskPayload: cloneStringPtr(entity.TaskPayload),
		TaskOptions: decodeTaskOption(entity.TaskOption),
		CronSpec:    cloneStringPtr(entity.CronSpec),
		Enable:      cloneBoolPtr(entity.Enable),
		Remark:      cloneStringPtr(entity.Remark),
		CreatedBy:   datautil.CloneUint32(toUint32(entity.CreateBy)),
		UpdatedBy:   datautil.CloneUint32(toUint32(entity.UpdateBy)),
		CreatedAt:   timeutil.TimeToTimestamppb(entity.CreatedAt),
		UpdatedAt:   timeutil.TimeToTimestamppb(entity.UpdatedAt),
		DeletedAt:   timeutil.TimeToTimestamppb(entity.DeletedAt),
	}
}

func stringToTaskType(s *string) *adminV1.Task_Type {
	if s == nil {
		return nil
	}
	if v, ok := adminV1.Task_Type_value[*s]; ok {
		val := adminV1.Task_Type(v)
		return &val
	}
	return nil
}

func taskTypeToString(tp *adminV1.Task_Type) *string {
	if tp == nil {
		return nil
	}
	s := tp.String()
	return &s
}

func encodeTaskOption(option *adminV1.TaskOption) *string {
	if option == nil {
		return nil
	}
	bytes, err := json.Marshal(option)
	if err != nil {
		return nil
	}
	str := string(bytes)
	return &str
}

func decodeTaskOption(raw *string) *adminV1.TaskOption {
	if raw == nil {
		return nil
	}
	var opt adminV1.TaskOption
	if err := json.Unmarshal([]byte(*raw), &opt); err != nil {
		return nil
	}
	return &opt
}
