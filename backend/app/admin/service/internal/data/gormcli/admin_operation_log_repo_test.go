package gormcli

import (
	"kratos-admin/app/admin/service/internal/data/gormcli/generated"
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/encoding"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gogf/gf/v2/util/gutil"
)

var (
	db        *gorm.DB
	dsn       = "host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=Abcd1234 dbname=postgres sslmode=disable TimeZone=Asia/Shanghai"
	newLogger = logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  true,          // Disable color
		},
	)
)

func openDB(t *testing.T) {
	t.Helper()
	var err error
	db, err = gorm.Open(
		postgres.Open(dsn), &gorm.Config{
			Logger: newLogger,
		},
	)
	if err != nil {
		t.Error(err)
	}
}

func TestAdminOperationLogRepo_Count(t *testing.T) {
	openDB(t)
	repo := NewAdminOperationLogRepo(db, nil)
	scope := func(db *gorm.Statement) {
		db.Where(generated.AdminOperationLog.Method.Eq("POST"))
	}
	scope2 := func(db *gorm.Statement) {
		db.Where(generated.AdminOperationLog.Reason.IsNotNull())
	}
	cnt, err := repo.Count(t.Context(), scope, scope2)
	if err != nil {
		t.Error(err)
	}
	t.Logf("cnt: %d", cnt)
}

func TestAdminOperationLogRepo_List(t *testing.T) {
	openDB(t)
	repo := NewAdminOperationLogRepo(db, nil)
	req := &pagination.PagingRequest{}
	var (
		page      int32 = 1
		pageSize  int32 = 20
		query     string
		orQuery   string
		orderBy   []string
		fieldMask fieldmaskpb.FieldMask
		noPaging  bool
		tenantId  uint32
	)
	queryJsonStr := `{"username":"admin","method":"POST","create_time__gte":"2025-11-04 15:42:54","create_time__lte":"2025-11-28 10:37:57"}`
	req.Page = &page
	req.PageSize = &pageSize
	codec := encoding.GetCodec("json")
	res, err := codec.Marshal(queryJsonStr)
	if err != nil {
		t.Error(err)
	}
	query = string(res)
	req.Query = &query
	req.OrQuery = &orQuery
	req.OrderBy = orderBy
	req.FieldMask = &fieldMask
	req.NoPaging = &noPaging
	req.TenantId = &tenantId

	list, err := repo.List(t.Context(), req)
	if err != nil {
		t.Error(list)
	}

	gutil.Dump(list)
}
