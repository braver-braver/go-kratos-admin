package data

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/model"
	"kratos-admin/app/admin/service/internal/data/gorm/query"
)

func newTestUserRepo(t *testing.T) (*UserRepo, *gorm.DB, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.User{})
	require.NoError(t, err)

	query.SetDefault(db)

	data := &Data{
		db:  db,
		log: log.NewHelper(log.With(log.DefaultLogger, "module", "user-repo-test")),
	}

	repo := NewUserRepo(log.DefaultLogger, data)

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}

	return repo, db, cleanup
}

func createUserForTest(t *testing.T, repo *UserRepo, ctx context.Context, username string, tenant uint32) *userV1.User {
	t.Helper()

	nickname := fmt.Sprintf("nick-%s", username)
	status := userV1.User_ON
	authority := userV1.User_CUSTOMER_USER

	req := &userV1.CreateUserRequest{
		Data: &userV1.User{
			Username:  &username,
			Nickname:  &nickname,
			Status:    &status,
			Authority: &authority,
			TenantId:  proto.Uint32(tenant),
		},
	}

	created, err := repo.Create(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, created)

	return created
}

func TestUserRepo_CreateAndGet(t *testing.T) {
	repo, _, cleanup := newTestUserRepo(t)
	defer cleanup()

	ctx := context.Background()

	user := createUserForTest(t, repo, ctx, "alice", 1)

	require.NotZero(t, user.GetId())
	require.Equal(t, "alice", user.GetUsername())
	require.Equal(t, userV1.User_ON, user.GetStatus())

	count, err := repo.Count(ctx, query.User.Username.Eq("alice"))
	require.NoError(t, err)
	require.Equal(t, 1, count)

	fetched, err := repo.Get(ctx, user.GetId())
	require.NoError(t, err)
	require.Equal(t, user.GetId(), fetched.GetId())
	require.Equal(t, user.GetNickname(), fetched.GetNickname())

	byName, err := repo.GetUserByUserName(ctx, "alice")
	require.NoError(t, err)
	require.Equal(t, user.GetId(), byName.GetId())
}

func TestUserRepo_UpdateWithFieldMask(t *testing.T) {
	repo, _, cleanup := newTestUserRepo(t)
	defer cleanup()

	ctx := context.Background()

	user := createUserForTest(t, repo, ctx, "bob", 2)

	newNickname := "updated-nick"
	newEmail := "bob@example.com"

	updateReq := &userV1.UpdateUserRequest{
		Data: &userV1.User{
			Id:        proto.Uint32(user.GetId()),
			Nickname:  &newNickname,
			Email:     &newEmail,
			UpdatedBy: proto.Uint32(77),
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"nickname", "email", "updated_by"}},
	}

	err := repo.Update(ctx, updateReq)
	require.NoError(t, err)

	updated, err := repo.Get(ctx, user.GetId())
	require.NoError(t, err)
	require.Equal(t, newNickname, updated.GetNickname())
	require.Equal(t, newEmail, updated.GetEmail())
	require.Equal(t, uint32(77), updated.GetUpdatedBy())
}

func TestUserRepo_ListWithQuery(t *testing.T) {
	repo, _, cleanup := newTestUserRepo(t)
	defer cleanup()

	ctx := context.Background()

	createUserForTest(t, repo, ctx, "carol", 3)
	createUserForTest(t, repo, ctx, "carter", 3)
	createUserForTest(t, repo, ctx, "dave", 3)

	page := int32(1)
	pageSize := int32(20)
	filter := "{\"username__contains\":\"car\"}"

	req := &pagination.PagingRequest{
		Page:     &page,
		PageSize: &pageSize,
		Query:    &filter,
		OrderBy:  []string{"username"},
	}

	resp, err := repo.List(ctx, req)
	require.NoError(t, err)
	require.Equal(t, uint32(2), resp.Total)
	require.Len(t, resp.Items, 2)
	for _, item := range resp.Items {
		require.Contains(t, item.GetUsername(), "car")
	}
}

func TestUserRepo_Delete(t *testing.T) {
	repo, db, cleanup := newTestUserRepo(t)
	defer cleanup()

	ctx := context.Background()

	user := createUserForTest(t, repo, ctx, "eric", 4)

	err := repo.Delete(ctx, user.GetId())
	require.NoError(t, err)

	exist, err := repo.IsExist(ctx, user.GetId())
	require.NoError(t, err)
	require.False(t, exist)

	var count int64
	err = db.WithContext(ctx).Table(model.TableNameUser).Count(&count).Error
	require.NoError(t, err)
	require.Zero(t, count)
}
