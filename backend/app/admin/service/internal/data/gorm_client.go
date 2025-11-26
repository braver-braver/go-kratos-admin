package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"

	gormcli "kratos-admin/app/admin/service/internal/data/gorm"
)

// NewGormClient wires the generated gorm client into the data provider set.
func NewGormClient(cfg *conf.Bootstrap, logger log.Logger) *gorm.DB {
	return gormcli.NewGormClient(cfg, logger)
}
