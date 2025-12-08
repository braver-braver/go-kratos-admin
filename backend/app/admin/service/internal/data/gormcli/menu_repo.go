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

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type MenuRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewMenuRepo(db *gorm.DB, logger log.Logger) *MenuRepo {
	return &MenuRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "menu/gormcli")),
	}
}

func (r *MenuRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysMenu](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *MenuRepo) List(ctx context.Context, req *pagination.PagingRequest, treeTravel bool) (*adminV1.ListMenuResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderMenu(req.GetOrderBy()),
		scopeFieldMaskMenu(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysMenu](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
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

	dtos := make([]*adminV1.Menu, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toMenuDTO(&entities[i]))
	}

	if treeTravel {
		roots := make([]*adminV1.Menu, 0)
		for _, node := range dtos {
			if node.GetParentId() == 0 {
				roots = append(roots, node)
				continue
			}
			travelChild(roots, node)
		}
		dtos = roots
	}

	return &adminV1.ListMenuResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func (r *MenuRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysMenu](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *MenuRepo) Get(ctx context.Context, req *adminV1.GetMenuRequest) (*adminV1.Menu, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysMenu](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskMenu(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, adminV1.ErrorNotFound("menu not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}
	return toMenuDTO(&entity), nil
}

func (r *MenuRepo) Create(ctx context.Context, req *adminV1.CreateMenuRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	entity := toMenuModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if err := gorm.G[model.SysMenu](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *MenuRepo) Update(ctx context.Context, req *adminV1.UpdateMenuRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &adminV1.CreateMenuRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch strings.ToLower(p) {
		case "parent_id":
			updates["parent_id"] = req.Data.GetParentId()
		case "type":
			updates["type"] = req.Data.GetType().String()
		case "path":
			updates["path"] = req.Data.GetPath()
		case "redirect":
			updates["redirect"] = req.Data.GetRedirect()
		case "alias":
			updates["alias"] = req.Data.GetAlias()
		case "name":
			updates["name"] = req.Data.GetName()
		case "component":
			updates["component"] = req.Data.GetComponent()
		case "status":
			updates["status"] = req.Data.GetStatus().String()
		case "meta":
			updates["meta"] = req.Data.Meta
		case "updated_by":
			updates["updated_by"] = req.Data.GetUpdatedBy()
		case "updated_at":
			updates["updated_at"] = toTime(req.Data.GetUpdatedAt())
		}
	}
	if len(updates) == 0 {
		return nil
	}
	if _, ok := updates["updated_at"]; !ok {
		updates["updated_at"] = time.Now()
	}

	db := r.db.WithContext(ctx).Model(&model.SysMenu{}).Where("id = ?", req.Data.GetId())
	if err := db.Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *MenuRepo) Delete(ctx context.Context, req *adminV1.DeleteMenuRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}
	ids := []int64{int64(req.GetId())}
	var children []int64
	if err := r.db.WithContext(ctx).Raw("WITH RECURSIVE cte AS (SELECT id, parent_id FROM sys_menus WHERE id = ? UNION ALL SELECT s.id, s.parent_id FROM sys_menus s INNER JOIN cte ON s.parent_id = cte.id) SELECT id FROM cte", req.GetId()).Scan(&children).Error; err == nil {
		ids = append(ids, children...)
	}
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.SysMenu{}).Error; err != nil {
		r.log.Errorf("delete menus failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete menus failed")
	}
	return nil
}

func toMenuDTO(entity *model.SysMenu) *adminV1.Menu {
	dto := &adminV1.Menu{}
	if entity == nil {
		return dto
	}
	id := uint32(entity.ID)
	dto.Id = &id
	dto.ParentId = toUint32Ptr(entity.ParentID)
	dto.Type = parseMenuType(entity.Type)
	dto.Path = &entity.Path
	dto.Redirect = &entity.Redirect
	dto.Alias = &entity.Alias
	dto.Name = &entity.Name
	dto.Component = &entity.Component
	dto.Status = parseMenuStatus(entity.Status)
	// Meta stored as JSON string in table; keep nil for parity with ent unless custom parsing is added.
	dto.CreatedBy = toUint32Ptr(entity.CreatedBy)
	dto.UpdatedBy = toUint32Ptr(entity.UpdatedBy)
	if !entity.CreatedAt.IsZero() {
		dto.CreatedAt = timestamppb.New(entity.CreatedAt)
	}
	if !entity.UpdatedAt.IsZero() {
		dto.UpdatedAt = timestamppb.New(entity.UpdatedAt)
	}
	return dto
}

func toMenuModel(dto *adminV1.Menu) *model.SysMenu {
	if dto == nil {
		return nil
	}
	entity := &model.SysMenu{
		ID:        int64(dto.GetId()),
		ParentID:  int64(dto.GetParentId()),
		Type:      dto.GetType().String(),
		Path:      dto.GetPath(),
		Redirect:  dto.GetRedirect(),
		Alias:     dto.GetAlias(),
		Name:      dto.GetName(),
		Component: dto.GetComponent(),
		Status:    dto.GetStatus().String(),
		CreatedBy: int64(dto.GetCreatedBy()),
		UpdatedBy: int64(dto.GetUpdatedBy()),
	}
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.CreatedAt.AsTime()
	}
	if dto.UpdatedAt != nil {
		entity.UpdatedAt = dto.UpdatedAt.AsTime()
	}
	return entity
}

var menuColumns = map[string]string{
	"id":         "id",
	"parent_id":  "parent_id",
	"type":       "type",
	"path":       "path",
	"redirect":   "redirect",
	"alias":      "alias",
	"name":       "name",
	"component":  "component",
	"status":     "status",
	"meta":       "meta",
	"created_by": "created_by",
	"updated_by": "updated_by",
	"created_at": "created_at",
	"updated_at": "updated_at",
	"deleted_at": "deleted_at",
	"deleted_by": "deleted_by",
}

func scopeOrderMenu(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := menuColumns[key]; ok {
				cols = append(cols, clause.OrderByColumn{
					Column: clause.Column{Name: col},
					Desc:   desc,
				})
			}
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskMenu(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := menuColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func parseMenuType(s string) *adminV1.Menu_Type {
	if v, ok := adminV1.Menu_Type_value[s]; ok {
		t := adminV1.Menu_Type(v)
		return &t
	}
	return nil
}

func parseMenuStatus(s string) *adminV1.Menu_Status {
	if v, ok := adminV1.Menu_Status_value[s]; ok {
		st := adminV1.Menu_Status(v)
		return &st
	}
	return nil
}

func travelChild(nodes []*adminV1.Menu, node *adminV1.Menu) bool {
	if nodes == nil {
		return false
	}
	if node.ParentId == nil || node.GetParentId() == 0 {
		return false
	}
	for _, n := range nodes {
		if n == nil || n.Id == nil {
			continue
		}
		if n.GetId() == node.GetParentId() {
			n.Children = append(n.Children, node)
			return true
		}
		if travelChild(n.Children, node) {
			return true
		}
	}
	return false
}
