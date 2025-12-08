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

	authV1 "kratos-admin/api/gen/go/authentication/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type UserCredentialRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewUserCredentialRepo(db *gorm.DB, logger log.Logger) *UserCredentialRepo {
	return &UserCredentialRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "user-credential/gormcli")),
	}
}

func (r *UserCredentialRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.SysUserCredential](r.db.WithContext(ctx)).Scopes(scopes...)
	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, authV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *UserCredentialRepo) List(ctx context.Context, req *pagination.PagingRequest) (*authV1.ListUserCredentialResponse, error) {
	if req == nil {
		return nil, authV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrderUserCredential(req.GetOrderBy()),
		scopeFieldMaskUserCredential(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.SysUserCredential](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
	total, err := r.Count(ctx, fScopes...)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, authV1.ErrorInternalServerError("query count failed")
	}

	entities, err := g.Find(ctx)
	if err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, authV1.ErrorInternalServerError("query list failed")
	}

	dtos := make([]*authV1.UserCredential, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toUserCredentialDTO(&entities[i]))
	}

	return &authV1.ListUserCredentialResponse{Total: uint32(total), Items: dtos}, nil
}

func (r *UserCredentialRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.SysUserCredential](r.db.WithContext(ctx)).
		Where("id = ?", id).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, authV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *UserCredentialRepo) Get(ctx context.Context, req *authV1.GetUserCredentialRequest) (*authV1.UserCredential, error) {
	if req == nil {
		return nil, authV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.SysUserCredential](r.db.WithContext(ctx)).
		Where("id = ?", req.GetId()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, authV1.ErrorNotFound("credential not found")
		}
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, authV1.ErrorInternalServerError("query data failed")
	}

	return toUserCredentialDTO(&entity), nil
}

func (r *UserCredentialRepo) GetCredentialByIdentifier(ctx context.Context, req *authV1.GetUserCredentialByIdentifierRequest) (*authV1.UserCredential, error) {
	if req == nil || req.GetIdentifier() == "" {
		return nil, authV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := gorm.G[model.SysUserCredential](r.db.WithContext(ctx)).
		Where("identity_type = ? AND identifier = ?", req.GetIdentityType().String(), req.GetIdentifier()).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, authV1.ErrorNotFound("credential not found")
		}
		r.log.Errorf("query credential failed: %s", err.Error())
		return nil, authV1.ErrorInternalServerError("query data failed")
	}

	return toUserCredentialDTO(&entity), nil
}

func (r *UserCredentialRepo) GetByIdentifier(ctx context.Context, req *authV1.GetUserCredentialByIdentifierRequest) (*authV1.UserCredential, error) {
	return r.GetCredentialByIdentifier(ctx, req)
}

func (r *UserCredentialRepo) Create(ctx context.Context, req *authV1.CreateUserCredentialRequest) error {
	if req == nil || req.Data == nil {
		return authV1.ErrorBadRequest("invalid parameter")
	}

	entity := toUserCredentialModel(req.Data)
	now := time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	if entity.UpdatedAt.IsZero() {
		entity.UpdatedAt = now
	}

	if err := gorm.G[model.SysUserCredential](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return authV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *UserCredentialRepo) Update(ctx context.Context, req *authV1.UpdateUserCredentialRequest) error {
	if req == nil || req.Data == nil {
		return authV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &authV1.CreateUserCredentialRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	updates := map[string]any{}
	for _, p := range req.GetUpdateMask().GetPaths() {
		switch p {
		case "tenant_id":
			updates["tenant_id"] = req.Data.GetTenantId()
		case "user_id":
			updates["user_id"] = req.Data.GetUserId()
		case "identity_type":
			updates["identity_type"] = req.Data.GetIdentityType().String()
		case "identifier":
			updates["identifier"] = req.Data.GetIdentifier()
		case "credential_type":
			updates["credential_type"] = req.Data.GetCredentialType().String()
		case "credential":
			updates["credential"] = req.Data.GetCredential()
		case "is_primary":
			updates["is_primary"] = req.Data.GetIsPrimary()
		case "status":
			updates["status"] = req.Data.GetStatus().String()
		case "extra_info":
			updates["extra_info"] = req.Data.GetExtraInfo()
		case "provider":
			updates["provider"] = req.Data.GetProvider()
		case "provider_account_id":
			updates["provider_account_id"] = req.Data.GetProviderAccountId()
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

	if err := r.db.WithContext(ctx).Model(&model.SysUserCredential{}).
		Where("id = ?", req.Data.GetId()).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update one data failed: %s", err.Error())
		return authV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *UserCredentialRepo) VerifyCredential(ctx context.Context, req *authV1.VerifyCredentialRequest) (*authV1.VerifyCredentialResponse, error) {
	if req == nil || req.GetIdentifier() == "" {
		return nil, authV1.ErrorBadRequest("invalid parameter")
	}
	cred, err := r.GetCredentialByIdentifier(ctx, &authV1.GetUserCredentialByIdentifierRequest{
		IdentityType: req.GetIdentityType(),
		Identifier:   req.GetIdentifier(),
	})
	if err != nil {
		return nil, err
	}
	return &authV1.VerifyCredentialResponse{
		Success: cred.GetCredential() == req.GetCredential(),
	}, nil
}

func (r *UserCredentialRepo) ChangeCredential(ctx context.Context, req *authV1.ChangeCredentialRequest) error {
	if req == nil || req.GetIdentifier() == "" {
		return authV1.ErrorBadRequest("invalid parameter")
	}
	cred, err := r.GetCredentialByIdentifier(ctx, &authV1.GetUserCredentialByIdentifierRequest{
		IdentityType: req.GetIdentityType(),
		Identifier:   req.GetIdentifier(),
	})
	if err != nil {
		return err
	}
	if cred.GetCredential() != req.GetOldCredential() {
		return authV1.ErrorBadRequest("credential mismatch")
	}
	if err := r.db.WithContext(ctx).
		Model(&model.SysUserCredential{}).
		Where("identity_type = ? AND identifier = ?", req.GetIdentityType().String(), req.GetIdentifier()).
		Updates(map[string]any{
			"credential": req.GetNewCredential(),
			"updated_at": time.Now(),
		}).Error; err != nil {
		r.log.Errorf("change credential failed: %s", err.Error())
		return authV1.ErrorInternalServerError("change credential failed")
	}
	return nil
}

func (r *UserCredentialRepo) ResetCredential(ctx context.Context, req *authV1.ResetCredentialRequest) error {
	if req == nil || req.GetIdentifier() == "" {
		return authV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.db.WithContext(ctx).
		Model(&model.SysUserCredential{}).
		Where("identity_type = ? AND identifier = ?", req.GetIdentityType().String(), req.GetIdentifier()).
		Updates(map[string]any{
			"credential": req.GetNewCredential(),
			"updated_at": time.Now(),
		}).Error; err != nil {
		r.log.Errorf("reset credential failed: %s", err.Error())
		return authV1.ErrorInternalServerError("reset credential failed")
	}
	return nil
}

func (r *UserCredentialRepo) Delete(ctx context.Context, id uint32) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SysUserCredential{}).Error; err != nil {
		r.log.Errorf("delete credential failed: %s", err.Error())
		return authV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

var userCredentialColumns = map[string]string{
	"id":                  "id",
	"tenant_id":           "tenant_id",
	"user_id":             "user_id",
	"identity_type":       "identity_type",
	"identifier":          "identifier",
	"credential_type":     "credential_type",
	"credential":          "credential",
	"is_primary":          "is_primary",
	"status":              "status",
	"extra_info":          "extra_info",
	"provider":            "provider",
	"provider_account_id": "provider_account_id",
	"created_at":          "created_at",
	"updated_at":          "updated_at",
	"deleted_at":          "deleted_at",
}

func scopeOrderUserCredential(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var cols []clause.OrderByColumn
		for _, ob := range orderBy {
			desc := strings.HasPrefix(ob, "-")
			key := strings.TrimPrefix(ob, "-")
			if col, ok := userCredentialColumns[key]; ok {
				cols = append(cols, clause.OrderByColumn{Column: clause.Column{Name: col}, Desc: desc})
			}
		}
		if len(cols) > 0 {
			db.Order(clause.OrderBy{Columns: cols})
		}
	}
}

func scopeFieldMaskUserCredential(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		var cols []string
		for _, p := range paths {
			if col, ok := userCredentialColumns[strings.ToLower(p)]; ok {
				cols = append(cols, col)
			}
		}
		if len(cols) > 0 {
			db.Select(cols)
		}
	}
}

func toUserCredentialDTO(entity *model.SysUserCredential) *authV1.UserCredential {
	dto := &authV1.UserCredential{}
	if entity == nil {
		return dto
	}
	dto.Id = uint32(entity.ID)
	dto.TenantId = toUint32Ptr(entity.TenantID)
	dto.UserId = toUint32Ptr(entity.UserID)
	dto.IdentityType = parseIdentityType(entity.IdentityType)
	dto.Identifier = &entity.Identifier
	dto.CredentialType = parseCredentialType(entity.CredentialType)
	dto.Credential = &entity.Credential
	dto.IsPrimary = &entity.IsPrimary
	dto.Status = parseCredentialStatus(entity.Status)
	dto.ExtraInfo = &entity.ExtraInfo
	dto.Provider = &entity.Provider
	dto.ProviderAccountId = &entity.ProviderAccountID
	if !entity.CreatedAt.IsZero() {
		dto.CreatedAt = timestamppb.New(entity.CreatedAt)
	}
	if !entity.UpdatedAt.IsZero() {
		dto.UpdatedAt = timestamppb.New(entity.UpdatedAt)
	}
	if entity.DeletedAt.Valid {
		dto.DeletedAt = timestamppb.New(entity.DeletedAt.Time)
	}
	return dto
}

func toUserCredentialModel(dto *authV1.UserCredential) *model.SysUserCredential {
	if dto == nil {
		return nil
	}
	entity := &model.SysUserCredential{
		ID:                int64(dto.GetId()),
		TenantID:          int64(dto.GetTenantId()),
		UserID:            int64(dto.GetUserId()),
		Identifier:        dto.GetIdentifier(),
		Credential:        dto.GetCredential(),
		IsPrimary:         dto.GetIsPrimary(),
		ExtraInfo:         dto.GetExtraInfo(),
		Provider:          dto.GetProvider(),
		ProviderAccountID: dto.GetProviderAccountId(),
	}
	if dto.IdentityType != nil {
		entity.IdentityType = dto.GetIdentityType().String()
	}
	if dto.CredentialType != nil {
		entity.CredentialType = dto.GetCredentialType().String()
	}
	if dto.Status != nil {
		entity.Status = dto.GetStatus().String()
	}
	if dto.CreatedAt != nil {
		entity.CreatedAt = dto.CreatedAt.AsTime()
	}
	if dto.UpdatedAt != nil {
		entity.UpdatedAt = dto.UpdatedAt.AsTime()
	}
	if dto.DeletedAt != nil {
		entity.DeletedAt = gorm.DeletedAt{Time: dto.DeletedAt.AsTime(), Valid: true}
	}
	return entity
}

func parseIdentityType(v string) *authV1.UserCredential_IdentityType {
	if v == "" {
		return nil
	}
	if val, ok := authV1.UserCredential_IdentityType_value[v]; ok {
		t := authV1.UserCredential_IdentityType(val)
		return &t
	}
	return nil
}

func parseCredentialType(v string) *authV1.UserCredential_CredentialType {
	if v == "" {
		return nil
	}
	if val, ok := authV1.UserCredential_CredentialType_value[v]; ok {
		t := authV1.UserCredential_CredentialType(val)
		return &t
	}
	return nil
}

func parseCredentialStatus(v string) *authV1.UserCredential_Status {
	if v == "" {
		return nil
	}
	if val, ok := authV1.UserCredential_Status_value[v]; ok {
		t := authV1.UserCredential_Status(val)
		return &t
	}
	return nil
}
