package data

import (
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"

	authnEngine "github.com/tx7do/kratos-authn/engine"
	"github.com/tx7do/kratos-authn/engine/jwt"

	"github.com/tx7do/go-utils/entgo"
	"github.com/tx7do/go-utils/password"
	"gorm.io/gorm"

	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"
	redisClient "github.com/tx7do/kratos-bootstrap/cache/redis"

	"kratos-admin/app/admin/service/internal/data/ent"

	"kratos-admin/pkg/oss"
)

// Data .
type Data struct {
	log *log.Helper

	rdb *redis.Client

	// ent (baseline) client
	db *entgo.EntClient[*ent.Client]
	// gorm client when enabled
	gormDB *gorm.DB

	authenticator authnEngine.Authenticator
	authorizer    *Authorizer

	useGorm bool
}

// NewData .
func NewData(
	cfg *conf.Bootstrap,
	logger log.Logger,
	db *entgo.EntClient[*ent.Client],
	gormDB *gorm.DB,
	rdb *redis.Client,
) (*Data, func(), error) {
	l := log.NewHelper(log.With(logger, "module", "data/admin-service"))

	d := &Data{
		log: l,

		db:     db,
		gormDB: gormDB,
		rdb:    rdb,

		useGorm: shouldUseGorm(cfg),
	}

	return d, func() {
		l.Info("closing the data resources")

		if d.db != nil {
			if err := d.db.Close(); err != nil {
				l.Error(err)
			}
		}

		if d.gormDB != nil {
			if sqlDB, err := d.gormDB.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}

		if d.rdb != nil {
			if err := d.rdb.Close(); err != nil {
				l.Error(err)
			}
		}
	}, nil
}

func (d *Data) UseGorm() bool {
	return d.useGorm && d.gormDB != nil
}

func (d *Data) EntClient() *entgo.EntClient[*ent.Client] {
	return d.db
}

func (d *Data) GormDB() *gorm.DB {
	return d.gormDB
}

func shouldUseGorm(cfg *conf.Bootstrap) bool {
	if cfg == nil {
		return false
	}
	if v := strings.ToLower(os.Getenv("USE_GORM")); v == "true" || v == "1" {
		return true
	}
	if cfg.GetData() != nil && cfg.GetData().GetDatabase() != nil {
		drv := strings.ToLower(cfg.GetData().GetDatabase().GetDriver())
		if strings.Contains(drv, "gorm") {
			return true
		}
	}
	return false
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(cfg *conf.Bootstrap, logger log.Logger) *redis.Client {
	l := log.NewHelper(log.With(logger, "module", "redis/data/admin-service"))
	return redisClient.NewClient(cfg.Data, l)
}

// NewAuthenticator 创建认证器
func NewAuthenticator(cfg *conf.Bootstrap) authnEngine.Authenticator {
	if cfg.Authn == nil {
		return nil
	}

	switch cfg.GetAuthn().GetType() {
	default:
		return nil

	case "jwt":
		authenticator, err := jwt.NewAuthenticator(
			jwt.WithKey([]byte(cfg.Authn.GetJwt().GetKey())),
			jwt.WithSigningMethod(cfg.Authn.GetJwt().GetMethod()),
		)
		if err != nil {
			return nil
		}
		return authenticator

	case "oidc":
		return nil

	case "preshared_key":
		return nil
	}
}

func NewUserTokenRepo(logger log.Logger, rdb *redis.Client, authenticator authnEngine.Authenticator, cfg *conf.Bootstrap) *UserTokenCacheRepo {
	return NewUserTokenCacheRepo(
		logger,
		rdb, authenticator,
		cfg.GetServer().GetRest().GetMiddleware().GetAuth().GetAccessTokenKeyPrefix(),
		cfg.GetServer().GetRest().GetMiddleware().GetAuth().GetRefreshTokenKeyPrefix(),
		cfg.GetServer().GetRest().GetMiddleware().GetAuth().GetAccessTokenExpires().AsDuration(),
		cfg.GetServer().GetRest().GetMiddleware().GetAuth().GetRefreshTokenExpires().AsDuration(),
	)
}

func NewMinIoClient(cfg *conf.Bootstrap, logger log.Logger) *oss.MinIOClient {
	return oss.NewMinIoClient(cfg, logger)
}

func NewPasswordCrypto() password.Crypto {
	crypto, err := password.CreateCrypto("bcrypt")
	if err != nil {
		panic(err)
	}
	return crypto
}
