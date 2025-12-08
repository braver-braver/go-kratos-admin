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

type AdminLoginLogRepo struct {
	log *log.Helper
	db  *gorm.DB
}

func NewAdminLoginLogRepo(db *gorm.DB, logger log.Logger) *AdminLoginLogRepo {
	return &AdminLoginLogRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "admin-login-log/gormcli")),
	}
}

func (r *AdminLoginLogRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysAdminLoginLog](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *AdminLoginLogRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListAdminLoginLogResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderLogin(req.GetOrderBy()),
		scopeFieldMaskLogin(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysAdminLoginLog](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
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

	dtos := make([]*adminV1.AdminLoginLog, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toAdminLoginLogDTO(&entities[i]))
	}

	return &adminV1.ListAdminLoginLogResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func scopeOrderLogin(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			col, ok := adminLoginColumns[key]
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

func scopeFieldMaskLogin(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := adminLoginColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func (r *AdminLoginLogRepo) Get(ctx context.Context, req *adminV1.GetAdminLoginLogRequest) (*adminV1.AdminLoginLog, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := gorm.G[model.SysAdminLoginLog](r.db.WithContext(ctx)).
		Scopes(scopeFieldMaskLogin(req.GetViewMask().GetPaths())).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, adminV1.ErrorNotFound("admin login log not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}
	return toAdminLoginLogDTO(&entity), nil
}

func (r *AdminLoginLogRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysAdminLoginLog](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *AdminLoginLogRepo) Create(ctx context.Context, req *adminV1.CreateAdminLoginLogRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	entity := toAdminLoginLogModel(req.Data)
	if entity.LoginTime.IsZero() {
		entity.LoginTime = time.Now()
	}
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = time.Now()
	}

	if err := gorm.G[model.SysAdminLoginLog](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func toAdminLoginLogDTO(entity *model.SysAdminLoginLog) *adminV1.AdminLoginLog {
	dto := &adminV1.AdminLoginLog{}
	if entity == nil {
		return dto
	}
	if entity.ID != 0 {
		id := uint32(entity.ID)
		dto.Id = &id
	}
	if entity.LoginIP != "" {
		dto.LoginIp = &entity.LoginIP
	}
	if entity.LoginMac != "" {
		dto.LoginMac = &entity.LoginMac
	}
	if !entity.LoginTime.IsZero() {
		dto.LoginTime = timestamppb.New(entity.LoginTime)
	}
	if entity.UserAgent != "" {
		dto.UserAgent = &entity.UserAgent
	}
	if entity.BrowserName != "" {
		dto.BrowserName = &entity.BrowserName
	}
	if entity.BrowserVersion != "" {
		dto.BrowserVersion = &entity.BrowserVersion
	}
	if entity.ClientID != "" {
		dto.ClientId = &entity.ClientID
	}
	if entity.ClientName != "" {
		dto.ClientName = &entity.ClientName
	}
	if entity.OsName != "" {
		dto.OsName = &entity.OsName
	}
	if entity.OsVersion != "" {
		dto.OsVersion = &entity.OsVersion
	}
	if entity.UserID != 0 {
		uid := uint32(entity.UserID)
		dto.UserId = &uid
	}
	if entity.Username != "" {
		dto.Username = &entity.Username
	}
	if entity.StatusCode != 0 {
		dto.StatusCode = &entity.StatusCode
	}
	dto.Success = &entity.Success
	if entity.Reason != "" {
		dto.Reason = &entity.Reason
	}
	if entity.Location != "" {
		dto.Location = &entity.Location
	}
	if !entity.CreatedAt.IsZero() {
		dto.CreatedAt = timestamppb.New(entity.CreatedAt)
	}
	return dto
}

func toAdminLoginLogModel(dto *adminV1.AdminLoginLog) *model.SysAdminLoginLog {
	if dto == nil {
		return nil
	}
	entity := &model.SysAdminLoginLog{
		ID:             int64(dto.GetId()),
		LoginIP:        strings.TrimSpace(dto.GetLoginIp()),
		LoginMac:       dto.GetLoginMac(),
		UserAgent:      dto.GetUserAgent(),
		BrowserName:    dto.GetBrowserName(),
		BrowserVersion: dto.GetBrowserVersion(),
		ClientID:       dto.GetClientId(),
		ClientName:     dto.GetClientName(),
		OsName:         dto.GetOsName(),
		OsVersion:      dto.GetOsVersion(),
		UserID:         int64(dto.GetUserId()),
		Username:       dto.GetUsername(),
		StatusCode:     dto.GetStatusCode(),
		Success:        dto.GetSuccess(),
		Reason:         dto.GetReason(),
		Location:       dto.GetLocation(),
	}
	if dto.LoginTime != nil {
		entity.LoginTime = dto.LoginTime.AsTime()
	}
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.CreatedAt.AsTime()
	}
	return entity
}

var adminLoginColumns = map[string]string{
	"id":              "id",
	"created_at":      "created_at",
	"login_ip":        "login_ip",
	"login_mac":       "login_mac",
	"login_time":      "login_time",
	"user_agent":      "user_agent",
	"browser_name":    "browser_name",
	"browser_version": "browser_version",
	"client_id":       "client_id",
	"client_name":     "client_name",
	"os_name":         "os_name",
	"os_version":      "os_version",
	"user_id":         "user_id",
	"username":        "username",
	"status_code":     "status_code",
	"success":         "success",
	"reason":          "reason",
	"location":        "location",
}
