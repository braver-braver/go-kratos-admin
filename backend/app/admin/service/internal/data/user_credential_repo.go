package data

import (
	"context"
	"errors"
	"strings"
	"time"

	authenticationV1 "kratos-admin/api/gen/go/authentication/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/model"
	"kratos-admin/app/admin/service/internal/data/gorm/query"
	"kratos-admin/pkg/datautil"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/password"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
)

type UserCredentialRepo struct {
	log    *log.Helper
	q      *query.Query
	crypto password.Crypto
}

func NewUserCredentialRepo(logger log.Logger, data *Data, passwordCrypto password.Crypto) *UserCredentialRepo {
	return &UserCredentialRepo{
		log:    log.NewHelper(log.With(logger, "module", "user-credential/repo/admin-service")),
		q:      query.Use(data.db),
		crypto: passwordCrypto,
	}
}

func (r *UserCredentialRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	count, err := r.q.UserCredential.WithContext(ctx).
		Where(r.q.UserCredential.ID.Eq(int32(id))).
		Count()
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, authenticationV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

func (r *UserCredentialRepo) Count(ctx context.Context, _ []func(any)) (int, error) {
	count, err := r.q.UserCredential.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, authenticationV1.ErrorInternalServerError("query count failed")
	}
	return int(count), nil
}

func (r *UserCredentialRepo) List(ctx context.Context, req *pagination.PagingRequest) (*authenticationV1.ListUserCredentialResponse, error) {
	if req == nil {
		return nil, authenticationV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.q.UserCredential.WithContext(ctx)
	builder = builder.Order(r.q.UserCredential.CreatedAt.Desc())

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
		return nil, authenticationV1.ErrorInternalServerError("query list failed")
	}

	total, err := r.q.UserCredential.WithContext(ctx).Count()
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, authenticationV1.ErrorInternalServerError("query count failed")
	}

	items := make([]*authenticationV1.UserCredential, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.toDTO(e))
	}

	return &authenticationV1.ListUserCredentialResponse{
		Total: uint32(total),
		Items: items,
	}, nil
}

func (r *UserCredentialRepo) Create(ctx context.Context, req *authenticationV1.CreateUserCredentialRequest) error {
	if req == nil || req.Data == nil {
		return authenticationV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.fromCreate(req.Data)
	if err != nil {
		return err
	}

	return r.q.UserCredential.WithContext(ctx).Create(entity)
}

func (r *UserCredentialRepo) Update(ctx context.Context, req *authenticationV1.UpdateUserCredentialRequest) error {
	if req == nil || req.Data == nil {
		return authenticationV1.ErrorBadRequest("invalid parameter")
	}

	update := r.buildUpdateMap(req.Data)
	_, err := r.q.UserCredential.WithContext(ctx).
		Where(r.q.UserCredential.ID.Eq(int32(req.Data.GetId()))).
		Updates(update)
	if err != nil {
		r.log.Errorf("update data failed: %s", err.Error())
		return authenticationV1.ErrorInternalServerError("update data failed")
	}
	return nil
}

func (r *UserCredentialRepo) Delete(ctx context.Context, id uint32) error {
	_, err := r.q.UserCredential.WithContext(ctx).
		Where(r.q.UserCredential.ID.Eq(int32(id))).
		Delete()
	if err != nil {
		r.log.Errorf("delete data failed: %s", err.Error())
		return authenticationV1.ErrorInternalServerError("delete data failed")
	}
	return nil
}

func (r *UserCredentialRepo) DeleteByUserId(ctx context.Context, userId uint32) error {
	_, err := r.q.UserCredential.WithContext(ctx).
		Where(r.q.UserCredential.UserID.Eq(int64(userId))).
		Delete()
	if err != nil {
		r.log.Errorf("delete by user id failed: %s", err.Error())
		return authenticationV1.ErrorInternalServerError("delete data failed")
	}
	return nil
}

func (r *UserCredentialRepo) DeleteByIdentifier(ctx context.Context, identityType authenticationV1.IdentityType, identifier string) error {
	_, err := r.q.UserCredential.WithContext(ctx).
		Where(
			r.q.UserCredential.IdentityType.Eq(identityType.String()),
			r.q.UserCredential.Identifier.Eq(identifier),
		).
		Delete()
	if err != nil {
		r.log.Errorf("delete by identifier failed: %s", err.Error())
		return authenticationV1.ErrorInternalServerError("delete data failed")
	}
	return nil
}

func (r *UserCredentialRepo) Get(ctx context.Context, req *authenticationV1.GetUserCredentialRequest) (*authenticationV1.UserCredential, error) {
	if req == nil {
		return nil, authenticationV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := r.q.UserCredential.WithContext(ctx).
		Where(r.q.UserCredential.ID.Eq(int32(req.GetId()))).
		First()
	if err != nil {
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, authenticationV1.ErrorNotFound("user credential not found")
	}
	return r.toDTO(entity), nil
}

func (r *UserCredentialRepo) GetByIdentifier(ctx context.Context, req *authenticationV1.GetUserCredentialByIdentifierRequest) (*authenticationV1.UserCredential, error) {
	if req == nil {
		return nil, authenticationV1.ErrorBadRequest("invalid parameter")
	}
	entity, err := r.q.UserCredential.WithContext(ctx).
		Where(
			r.q.UserCredential.IdentityType.Eq(req.GetIdentityType().String()),
			r.q.UserCredential.Identifier.Eq(req.GetIdentifier()),
		).
		First()
	if err != nil {
		return nil, authenticationV1.ErrorNotFound("user credential not found")
	}
	return r.toDTO(entity), nil
}

func (r *UserCredentialRepo) VerifyCredential(ctx context.Context, req *authenticationV1.VerifyCredentialRequest) (*authenticationV1.VerifyCredentialResponse, error) {
	if req == nil {
		return nil, authenticationV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.q.UserCredential.WithContext(ctx).
		Where(
			r.q.UserCredential.IdentityType.Eq(req.GetIdentityType().String()),
			r.q.UserCredential.Identifier.Eq(req.GetIdentifier()),
		).
		First()
	if err != nil {
		return nil, authenticationV1.ErrorUnauthorized("invalid credential")
	}

	credVal := ""
	if entity.Credential != nil {
		credVal = *entity.Credential
	}
	if !r.verifyCredential(entity.CredentialType, req.GetCredential(), credVal) {
		return nil, authenticationV1.ErrorUnauthorized("invalid credential")
	}

	return &authenticationV1.VerifyCredentialResponse{
		Success: true,
	}, nil
}

func (r *UserCredentialRepo) ChangeCredential(ctx context.Context, req *authenticationV1.ChangeCredentialRequest) error {
	if req == nil {
		return authenticationV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := r.q.UserCredential.WithContext(ctx).
		Where(
			r.q.UserCredential.IdentityType.Eq(req.GetIdentityType().String()),
			r.q.UserCredential.Identifier.Eq(req.GetIdentifier()),
		).
		First()
	if err != nil {
		return authenticationV1.ErrorNotFound("user credential not found")
	}

	old := ""
	if entity.Credential != nil {
		old = *entity.Credential
	}
	if !r.verifyCredential(entity.CredentialType, req.GetOldCredential(), old) {
		return authenticationV1.ErrorUnauthorized("invalid credential")
	}

	cred, err := r.prepareCredential(entity.CredentialType, req.GetNewCredential())
	if err != nil {
		return authenticationV1.ErrorInternalServerError("prepare credential failed")
	}

	_, err = r.q.UserCredential.WithContext(ctx).
		Where(r.q.UserCredential.ID.Eq(entity.ID)).
		Update(r.q.UserCredential.Credential, cred)
	return err
}

func (r *UserCredentialRepo) ResetCredential(ctx context.Context, req *authenticationV1.ResetCredentialRequest) error {
	if req == nil {
		return authenticationV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.q.UserCredential.WithContext(ctx).
		Where(
			r.q.UserCredential.IdentityType.Eq(req.GetIdentityType().String()),
			r.q.UserCredential.Identifier.Eq(req.GetIdentifier()),
		)

	entity, err := builder.First()
	if err != nil {
		return authenticationV1.ErrorNotFound("user credential not found")
	}

	cred, err := r.prepareCredential(entity.CredentialType, req.GetNewCredential())
	if err != nil {
		return authenticationV1.ErrorInternalServerError("prepare credential failed")
	}

	_, err = builder.Update(r.q.UserCredential.Credential, cred)
	return err
}

func (r *UserCredentialRepo) fromCreate(data *authenticationV1.UserCredential) (*model.UserCredential, error) {
	credType := credentialTypeToString(data.CredentialType)
	cred, err := r.prepareCredential(credType, data.GetCredential())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &model.UserCredential{
		CreatedAt:      &now,
		UpdatedAt:      &now,
		TenantID:       cloneInt64FromUint32(data.TenantId),
		UserID:         cloneInt64FromUint32(data.UserId),
		IdentityType:   identityTypeToString(data.IdentityType),
		Identifier:     cloneStringPtr(data.Identifier),
		CredentialType: credType,
		Credential:     cloneStringPtr(&cred),
		IsPrimary:      cloneBoolPtr(data.IsPrimary),
		Status:         credentialStatusToString(data.Status),
		ExtraInfo:      cloneStringPtr(data.ExtraInfo),
	}, nil
}

func (r *UserCredentialRepo) buildUpdateMap(data *authenticationV1.UserCredential) map[string]any {
	update := map[string]any{}
	if data.TenantId != nil {
		update["tenant_id"] = data.GetTenantId()
	}
	if data.UserId != nil {
		update["user_id"] = data.GetUserId()
	}
	if data.IdentityType != nil {
		update["identity_type"] = data.GetIdentityType().String()
	}
	if data.Identifier != nil {
		update["identifier"] = data.GetIdentifier()
	}
	if data.CredentialType != nil {
		update["credential_type"] = data.GetCredentialType().String()
	}
	if data.Credential != nil {
		credType := credentialTypeToString(data.CredentialType)
		if cred, err := r.prepareCredential(credType, data.GetCredential()); err == nil {
			update["credential"] = cred
		}
	}
	if data.IsPrimary != nil {
		update["is_primary"] = data.GetIsPrimary()
	}
	if data.Status != nil {
		update["status"] = data.GetStatus().String()
	}
	if data.ExtraInfo != nil {
		update["extra_info"] = data.GetExtraInfo()
	}
	update["updated_at"] = time.Now()
	return update
}

func (r *UserCredentialRepo) toDTO(entity *model.UserCredential) *authenticationV1.UserCredential {
	if entity == nil {
		return nil
	}

	return &authenticationV1.UserCredential{
		Id:             uint32(entity.ID),
		TenantId:       datautil.CloneUint32(toUint32(entity.TenantID)),
		UserId:         datautil.CloneUint32(toUint32(entity.UserID)),
		IdentityType:   stringToIdentityType(entity.IdentityType),
		Identifier:     cloneStringPtr(entity.Identifier),
		CredentialType: stringToCredentialType(entity.CredentialType),
		Credential:     entity.Credential,
		IsPrimary:      cloneBoolPtr(entity.IsPrimary),
		Status:         stringToCredentialStatus(entity.Status),
		ExtraInfo:      cloneStringPtr(entity.ExtraInfo),
		CreatedAt:      timeutil.TimeToTimestamppb(entity.CreatedAt),
		UpdatedAt:      timeutil.TimeToTimestamppb(entity.UpdatedAt),
		DeletedAt:      timeutil.TimeToTimestamppb(entity.DeletedAt),
	}
}

func (r *UserCredentialRepo) verifyCredential(credentialType *string, plainCredential, targetCredential string) bool {
	if credentialType == nil {
		return false
	}
	switch strings.ToUpper(*credentialType) {
	case authenticationV1.UserCredential_PASSWORD_HASH.String():
		ok, _ := r.crypto.Verify(targetCredential, plainCredential)
		return ok
	default:
		return false
	}
}

func (r *UserCredentialRepo) prepareCredential(credentialType *string, plainCredential string) (string, error) {
	if credentialType == nil {
		return "", errors.New("missing credential type")
	}
	switch strings.ToUpper(*credentialType) {
	case authenticationV1.UserCredential_PASSWORD_HASH.String():
		return r.crypto.Encrypt(plainCredential)
	default:
		return "", errors.New("unsupported credential type")
	}
}

func stringToIdentityType(s *string) *authenticationV1.IdentityType {
	if s == nil {
		return nil
	}
	if v, ok := authenticationV1.IdentityType_value[*s]; ok {
		val := authenticationV1.IdentityType(v)
		return &val
	}
	return nil
}

func stringToCredentialType(s *string) *authenticationV1.UserCredential_Type {
	if s == nil {
		return nil
	}
	if v, ok := authenticationV1.UserCredential_Type_value[*s]; ok {
		val := authenticationV1.UserCredential_Type(v)
		return &val
	}
	return nil
}

func stringToCredentialStatus(s *string) *authenticationV1.UserCredential_Status {
	if s == nil {
		return nil
	}
	if v, ok := authenticationV1.UserCredential_Status_value[*s]; ok {
		val := authenticationV1.UserCredential_Status(v)
		return &val
	}
	return nil
}

func identityTypeToString(t *authenticationV1.IdentityType) *string {
	if t == nil {
		return nil
	}
	s := t.String()
	return &s
}

func credentialTypeToString(t *authenticationV1.UserCredential_Type) *string {
	if t == nil {
		return nil
	}
	s := t.String()
	return &s
}

func credentialStatusToString(s *authenticationV1.UserCredential_Status) *string {
	if s == nil {
		return nil
	}
	str := s.String()
	return &str
}
