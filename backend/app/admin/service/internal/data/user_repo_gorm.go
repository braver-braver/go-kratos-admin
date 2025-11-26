package data

import (
	"context"

	userV1 "kratos-admin/api/gen/go/user/service/v1"
	"kratos-admin/app/admin/service/internal/data/gorm/repositories"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gorm"
)

// UserRepoGorm GORM 版本的 User Repository
type UserRepoGorm struct {
	repo *repositories.UserRepo
	log  *log.Helper
}

func NewUserRepoGorm(logger log.Logger, db interface{}) *UserRepoGorm {
	// 这里需要从 Data 结构体中获取 GORM 客户端
	// 暂时使用类型断言，实际使用时需要调整
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("invalid database client type")
	}

	repo := repositories.NewUserRepo(gormDB, logger)

	return &UserRepoGorm{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "user/repo-gorm/admin-service")),
	}
}

func (r *UserRepoGorm) Count(ctx context.Context, whereCond []interface{}) (int, error) {
	// 将 Ent 的 whereCond 转换为 GORM 的 conditions
	conditions := make(map[string]interface{})
	// 这里需要实现条件转换逻辑
	// 简化实现
	count, err := r.repo.Count(ctx, conditions)
	return int(count), err
}

func (r *UserRepoGorm) List(ctx context.Context, req *pagination.PagingRequest) (*userV1.ListUserResponse, error) {
	return r.repo.List(ctx, req)
}

func (r *UserRepoGorm) IsExist(ctx context.Context, id uint32) (bool, error) {
	return r.repo.IsExist(ctx, id)
}

func (r *UserRepoGorm) Get(ctx context.Context, userId uint32) (*userV1.User, error) {
	return r.repo.Get(ctx, userId)
}

func (r *UserRepoGorm) Create(ctx context.Context, req *userV1.CreateUserRequest) (*userV1.User, error) {
	return r.repo.Create(ctx, req)
}

func (r *UserRepoGorm) Update(ctx context.Context, req *userV1.UpdateUserRequest) error {
	return r.repo.Update(ctx, req)
}

func (r *UserRepoGorm) Delete(ctx context.Context, userId uint32) error {
	return r.repo.Delete(ctx, userId)
}

func (r *UserRepoGorm) GetUserByUserName(ctx context.Context, userName string) (*userV1.User, error) {
	return r.repo.GetUserByUserName(ctx, userName)
}

func (r *UserRepoGorm) UserExists(ctx context.Context, req *userV1.UserExistsRequest) (*userV1.UserExistsResponse, error) {
	return r.repo.UserExists(ctx, req)
}

func (r *UserRepoGorm) GetUsersByIds(ctx context.Context, ids []uint32) ([]*userV1.User, error) {
	return r.repo.GetUsersByIds(ctx, ids)
}
