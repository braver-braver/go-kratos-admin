package gormcli

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/stringcase"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/cli/gorm/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
	"kratos-admin/app/admin/service/internal/data/gormcli/generated"
	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

type AdminOperationLogRepo struct {
	log *log.Helper
	db  *gorm.DB
}

func NewAdminOperationLogRepo(db *gorm.DB, logger log.Logger) *AdminOperationLogRepo {
	return &AdminOperationLogRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "admin-operation-log/gormcli")),
	}
}

func (r *AdminOperationLogRepo) Get(ctx context.Context, req *adminV1.GetAdminOperationLogRequest) (*adminV1.AdminOperationLog, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	g := gorm.G[model.AdminOperationLog](r.db.WithContext(ctx)).
		Scopes(scopeFieldMask(req.GetViewMask().GetPaths()))

	entity, err := g.Where(generated.AdminOperationLog.ID.Eq(req.GetId())).Take(ctx)
	switch {
	case err == nil:
		return toAdminOperationLogDTO(&entity), nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, adminV1.ErrorNotFound("admin operation log not found")
	default:
		r.log.Errorf("query one data failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}
}

func (r *AdminOperationLogRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	total, err := gorm.G[model.AdminOperationLog](r.db.WithContext(ctx)).
		Where(generated.AdminOperationLog.ID.Eq(id)).
		Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return total > 0, nil
}

func (r *AdminOperationLogRepo) Create(ctx context.Context, req *adminV1.CreateAdminOperationLogRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	entity := toAdminOperationLogModel(req.Data)
	if entity.CreatedAt == nil {
		now := time.Now()
		entity.CreatedAt = &now
	}

	if err := gorm.G[model.AdminOperationLog](r.db.WithContext(ctx)).Create(ctx, entity); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert data failed")
	}
	return nil
}

func (r *AdminOperationLogRepo) Count(ctx context.Context, scopes ...func(db *gorm.Statement)) (int, error) {
	g := gorm.G[model.AdminOperationLog](r.db.WithContext(ctx)).Scopes(scopes...)

	total, err := g.Count(ctx, "*")
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}
	return int(total), nil
}

func (r *AdminOperationLogRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListAdminOperationLogResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	scopes := []func(db *gorm.Statement){
		scopePaging(req.GetNoPaging(), req.GetPage(), req.GetPageSize()),
		scopeOrder(req.GetOrderBy()),
		scopeFieldMask(req.GetFieldMask().GetPaths()),
		defaultLimitGuard(),
	}
	// 过滤掉 nil 的 scope
	scopes = removeNilScopes(scopes)

	fScopes := []func(db *gorm.Statement){
		scopeFilters(req.GetQuery(), false),
		scopeFilters(req.GetOrQuery(), true),
	}
	// 过滤掉 nil 的 scope
	fScopes = removeNilScopes(fScopes)

	g := gorm.G[model.AdminOperationLog](r.db.WithContext(ctx)).Scopes(scopes...).Scopes(fScopes...)
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

	dtos := make([]*adminV1.AdminOperationLog, 0, len(entities))
	for i := range entities {
		dtos = append(dtos, toAdminOperationLogDTO(&entities[i]))
	}

	return &adminV1.ListAdminOperationLogResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

func toAdminOperationLogDTO(entity *model.AdminOperationLog) *adminV1.AdminOperationLog {
	dto := &adminV1.AdminOperationLog{}

	if entity.ID != 0 {
		id := entity.ID
		dto.Id = &id
	}
	if entity.CostTime != nil {
		duration := time.Duration(float64(time.Second) * *entity.CostTime)
		dto.CostTime = durationpb.New(duration)
	}
	if entity.Success != nil {
		dto.Success = entity.Success
	}
	if entity.RequestID != nil {
		dto.RequestId = entity.RequestID
	}
	if entity.StatusCode != nil {
		dto.StatusCode = entity.StatusCode
	}
	if entity.Reason != nil {
		dto.Reason = entity.Reason
	}
	if entity.Location != nil {
		dto.Location = entity.Location
	}
	if entity.Operation != nil {
		dto.Operation = entity.Operation
	}
	if entity.Method != nil {
		dto.Method = entity.Method
	}
	if entity.Path != nil {
		dto.Path = entity.Path
	}
	if entity.Referer != nil {
		dto.Referer = entity.Referer
	}
	if entity.RequestURI != nil {
		dto.RequestUri = entity.RequestURI
	}
	if entity.RequestHeader != nil {
		dto.RequestHeader = entity.RequestHeader
	}
	if entity.RequestBody != nil {
		dto.RequestBody = entity.RequestBody
	}
	if entity.Response != nil {
		dto.Response = entity.Response
	}
	if entity.UserID != nil {
		dto.UserId = entity.UserID
	}
	if entity.Username != nil {
		dto.Username = entity.Username
	}
	if entity.ClientIP != nil {
		dto.ClientIp = entity.ClientIP
	}
	if entity.UserAgent != nil {
		dto.UserAgent = entity.UserAgent
	}
	if entity.BrowserName != nil {
		dto.BrowserName = entity.BrowserName
	}
	if entity.BrowserVersion != nil {
		dto.BrowserVersion = entity.BrowserVersion
	}
	if entity.ClientID != nil {
		dto.ClientId = entity.ClientID
	}
	if entity.ClientName != nil {
		dto.ClientName = entity.ClientName
	}
	if entity.OsName != nil {
		dto.OsName = entity.OsName
	}
	if entity.OsVersion != nil {
		dto.OsVersion = entity.OsVersion
	}
	if entity.CreatedAt != nil {
		dto.CreatedAt = timestamppb.New(*entity.CreatedAt)
	}

	return dto
}

func toAdminOperationLogModel(dto *adminV1.AdminOperationLog) *model.AdminOperationLog {
	if dto == nil {
		return nil
	}
	return &model.AdminOperationLog{
		ID:             dto.GetId(),
		CostTime:       timeutil.DurationpbToSecond(dto.CostTime),
		Success:        dto.Success,
		RequestID:      dto.RequestId,
		StatusCode:     dto.StatusCode,
		Reason:         dto.Reason,
		Location:       dto.Location,
		Operation:      dto.Operation,
		Method:         dto.Method,
		Path:           dto.Path,
		Referer:        dto.Referer,
		RequestURI:     dto.RequestUri,
		RequestHeader:  dto.RequestHeader,
		RequestBody:    dto.RequestBody,
		Response:       dto.Response,
		UserID:         dto.UserId,
		Username:       dto.Username,
		ClientIP:       dto.ClientIp,
		UserAgent:      dto.UserAgent,
		BrowserName:    dto.BrowserName,
		BrowserVersion: dto.BrowserVersion,
		ClientID:       dto.ClientId,
		ClientName:     dto.ClientName,
		OsName:         dto.OsName,
		OsVersion:      dto.OsVersion,
		CreatedAt:      timestamppbToTimePtr(dto.CreatedAt),
	}
}

func timestamppbToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// resolveColumn maps external field names to generated safe column names.
func resolveColumn(input string) (field.ColumnInterface, bool) {
	col, ok := columnWhitelist[strings.ToLower(strings.TrimSpace(input))]
	return col, ok
}

// scopeOrder 拼接排序字段
func scopeOrder(orderBy []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(orderBy) == 0 {
			return
		}
		var desc bool
		var orderByColumns = make([]clause.OrderByColumn, 0)
		for _, o := range orderBy {
			desc = false
			if strings.HasPrefix(o, "-") {
				o = o[1:]
				desc = true
			}
			field, ok := columnWhitelist[o]
			if !ok {
				continue
			}
			orderByColumns = append(
				orderByColumns, clause.OrderByColumn{
					Column: field.Column(),
					Desc:   desc,
				},
			)
		}
		db.Order(
			clause.OrderBy{
				Columns: orderByColumns,
			},
		)
	}
}

// scopePaging 拼接分页请求
func scopePaging(noPaging bool, page, pageSize int32) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if noPaging {
			return
		}

		if page <= 0 {
			page = 1
		}

		if pageSize <= 0 {
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		db.Offset(int(offset)).Limit(int(pageSize))
	}
}

// scopeFieldMask 选择查询的字段.
func scopeFieldMask(paths []string) func(db *gorm.Statement) {
	return func(db *gorm.Statement) {
		if len(paths) == 0 {
			return
		}
		cols := make([]string, 0, len(paths))
		for _, p := range paths {
			if col, ok := resolveColumn(p); ok {
				cols = append(cols, col.Column().Name)
			}
		}
		if len(cols) == 0 {
			return
		}
		db.Select(cols)
		return
	}
}

// scopeFilters 将 query 过滤条件转化为 scope代码快.
func scopeFilters(strJson string, isOr bool) func(db *gorm.Statement) {
	if strings.TrimSpace(strJson) == "" {
		return nil
	}

	queryMap := make(map[string]string)

	var queryMapArray []map[string]string
	if err1 := json.Unmarshal([]byte(strJson), &queryMap); err1 != nil {
		if err2 := json.Unmarshal([]byte(strJson), &queryMapArray); err2 != nil {
			// 尝试解析为单个 map 失败，尝试解析为数组也失败，直接返回
			return nil
		}
		// Process the array of query maps
		return func(db *gorm.Statement) {
			for _, qMap := range queryMapArray {
				processQueryMap(db, qMap, isOr)
			}
		}
	}

	return func(db *gorm.Statement) {
		processQueryMap(db, queryMap, isOr)
	}
}

const (
	QueryDelimiter     = "__" // 分隔符
	JsonFieldDelimiter = "."  // JSONB字段分隔符
)

func processQueryMap(db *gorm.Statement, queryMap map[string]string, isOr bool) {
	for k, v := range queryMap {
		keys := strings.Split(k, QueryDelimiter)
		_ = makeFieldFilter(db, isOr, keys, v) // ignore errors for individual filters
	}
}

// makeFieldFilter 构建一个字段过滤器
func makeFieldFilter(db *gorm.Statement, isOr bool, keys []string, value string) error {
	if len(keys) == 0 {
		return errors.New("keys 为空")
	}
	if len(value) == 0 {
		return errors.New("value 为空")
	}
	field := keys[0]
	if len(field) == 0 {
		return errors.New("非法过滤条件")
	}
	switch len(keys) {
	case 1:
		field = stringcase.ToSnakeCase(field)
		if isOr {
			db.Or(field+" = ?", value)
		} else {
			db.Where(field+" = ?", value)
		}

	case 2:
		op := keys[1]
		if len(op) == 0 {
			return errors.New("未找到有效的操作符")
		}
		field = stringcase.ToSnakeCase(field)
		processOp(db, op, field, value)
	default:
		return errors.New("暂未支持两个以上操作符")
	}
	return nil
}

var columnWhitelist = map[string]field.ColumnInterface{
	"id":              generated.AdminOperationLog.ID,
	"created_at":      generated.AdminOperationLog.CreatedAt,
	"request_id":      generated.AdminOperationLog.RequestID,
	"method":          generated.AdminOperationLog.Method,
	"operation":       generated.AdminOperationLog.Operation,
	"path":            generated.AdminOperationLog.Path,
	"referer":         generated.AdminOperationLog.Referer,
	"request_uri":     generated.AdminOperationLog.RequestURI,
	"request_body":    generated.AdminOperationLog.RequestBody,
	"request_header":  generated.AdminOperationLog.RequestHeader,
	"response":        generated.AdminOperationLog.Response,
	"cost_time":       generated.AdminOperationLog.CostTime,
	"user_id":         generated.AdminOperationLog.UserID,
	"username":        generated.AdminOperationLog.Username,
	"client_ip":       generated.AdminOperationLog.ClientIP,
	"status_code":     generated.AdminOperationLog.StatusCode,
	"reason":          generated.AdminOperationLog.Reason,
	"success":         generated.AdminOperationLog.Success,
	"location":        generated.AdminOperationLog.Location,
	"user_agent":      generated.AdminOperationLog.UserAgent,
	"browser_name":    generated.AdminOperationLog.BrowserName,
	"browser_version": generated.AdminOperationLog.BrowserVersion,
	"client_id":       generated.AdminOperationLog.ClientID,
	"client_name":     generated.AdminOperationLog.ClientName,
	"os_name":         generated.AdminOperationLog.OsName,
	"os_version":      generated.AdminOperationLog.OsVersion,
}

func processOp(db *gorm.Statement, op, field, value string) {
	switch op {
	case ops[FilterNot]:
		filterNot(db, field, value)
	case ops[FilterIn]:
		filterIn(db, field, value)
	case ops[FilterNotIn]:
		filterNotIn(db, field, value)
	case ops[FilterGTE]:
		filterGTE(db, field, value)
	case ops[FilterGT]:
		filterGT(db, field, value)
	case ops[FilterLTE]:
		filterLTE(db, field, value)
	case ops[FilterLT]:
		filterLT(db, field, value)
	case ops[FilterRange]:
		filterRange(db, field, value)
	case ops[FilterIsNull]:
		filterIsNull(db, field, value)
	case ops[FilterNotIsNull]:
		filterIsNotNull(db, field, value)
	case ops[FilterContains]:
		filterContains(db, field, value)
	case ops[FilterInsensitiveContains]:
		filterInsensitiveContains(db, field, value)
	case ops[FilterStartsWith]:
		filterStartsWith(db, field, value)
	case ops[FilterInsensitiveStartsWith]:
		filterInsensitiveStartsWith(db, field, value)
	case ops[FilterEndsWith]:
		filterEndsWith(db, field, value)
	case ops[FilterInsensitiveEndsWith]:
		filterInsensitiveEndsWith(db, field, value)
	case ops[FilterExact]:
		filterExact(db, field, value)
	case ops[FilterInsensitiveExact]:
		filterInsensitiveExact(db, field, value)
	case ops[FilterRegex]:
		filterRegex(db, field, value)
	case ops[FilterInsensitiveRegex]:
		filterInsensitiveRegex(db, field, value)
	case ops[FilterSearch]:
		filterSearch(db, field, value)
	default:
		// 检查是否为日期提取类型的操作符
		handleDatePartOp(db, op, field, value)
	}

}

// handleDatePartOp 处理日期提取类型的操作符
func handleDatePartOp(db *gorm.Statement, op, field, value string) {
	// 检查操作符是否匹配日期部分
	switch op {
	case "date":
		db.Where("DATE("+field+") = ?", value)
	case "year":
		db.Where("EXTRACT(YEAR FROM "+field+") = ?", value)
	case "iso_year":
		db.Where("EXTRACT(ISOYEAR FROM "+field+") = ?", value)
	case "quarter":
		db.Where("EXTRACT(QUARTER FROM "+field+") = ?", value)
	case "month":
		db.Where("EXTRACT(MONTH FROM "+field+") = ?", value)
	case "week":
		db.Where("EXTRACT(WEEK FROM "+field+") = ?", value)
	case "week_day":
		db.Where("EXTRACT(DOW FROM "+field+") = ?", value) // PostgreSQL uses DOW (0-6, Sunday=0)
	case "iso_week_day":
		db.Where("EXTRACT(ISODOW FROM "+field+") = ?", value) // PostgreSQL uses ISODOW (1-7, Monday=1)
	case "day":
		db.Where("EXTRACT(DAY FROM "+field+") = ?", value)
	case "time":
		db.Where("TIME("+field+") = ?", value)
	case "hour":
		db.Where("EXTRACT(HOUR FROM "+field+") = ?", value)
	case "minute":
		db.Where("EXTRACT(MINUTE FROM "+field+") = ?", value)
	case "second":
		db.Where("EXTRACT(SECOND FROM "+field+") = ?", value)
	case "microsecond":
		db.Where("EXTRACT(MICROSECOND FROM "+field+") = ?", value)
	}
}

// removeNilScopes 移除 nil 的 scope
func removeNilScopes(scopes []func(db *gorm.Statement)) []func(db *gorm.Statement) {
	result := make([]func(db *gorm.Statement), 0, len(scopes))
	for _, scope := range scopes {
		if scope != nil {
			result = append(result, scope)
		}
	}
	return result
}

// filterEqual = 相等操作
// SQL: WHERE "name" = "tom"
func filterEqual(db *gorm.Statement, field, value string) {
	db.Where(field+" = ? ", value)
}

// filterNot NOT 不相等操作
// SQL: WHERE NOT ("name" = "tom")
// 或者： WHERE "name" <> "tom"
// 用NOT可以过滤出NULL，而用<>、!=则不能。
func filterNot(db *gorm.Statement, field, value string) {
	db.Where(field+" <> ? ", value)
}

// filterIn IN操作
// SQL: WHERE name IN ("tom", "jimmy")
func filterIn(db *gorm.Statement, field, value string) {
	var values []interface{}
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return
	}
	db.Where(field+" IN ? ", values)
}

// filterNotIn NOT IN操作
// SQL: WHERE name NOT IN ("tom", "jimmy")`
func filterNotIn(db *gorm.Statement, field, value string) {
	var values []interface{}
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return
	}
	db.Where(field+" NOT IN ? ", values)
}

// filterGTE GTE (Greater Than or Equal) 大于等于 >=操作
// SQL: WHERE "create_time" >= "2023-10-25"
func filterGTE(db *gorm.Statement, field, value string) {
	db.Where(field+" >= ? ", value)
}

// filterGT GT (Greater than) 大于 >操作
// SQL: WHERE "create_time" > "2023-10-25"
func filterGT(db *gorm.Statement, field, value string) {
	db.Where(field+" > ? ", value)
}

// filterLTE LTE (Less Than or Equal) 小于等于 <=操作
// SQL: WHERE "create_time" <= "2023-10-25"
func filterLTE(db *gorm.Statement, field, value string) {
	db.Where(field+" <= ? ", value)
}

// filterLT LT (Less than) 小于 <操作
// SQL: WHERE "create_time" < "2023-10-25"
func filterLT(db *gorm.Statement, field, value string) {
	db.Where(field+" < ? ", value)
}

// filterRange 在值域之中 BETWEEN操作
// SQL: WHERE "create_time" BETWEEN "2023-10-25" AND "2024-10-25"
// 或者： WHERE "create_time" >= "2023-10-25" AND "create_time" <= "2024-10-25"
func filterRange(db *gorm.Statement, field, value string) {
	var values []interface{}
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return
	}
	if len(values) != 2 {
		return
	}
	db.Where(field+" BETWEEN ? AND ? ", values[0], values[1])
}

// filterIsNull 为空 IS NULL操作
// SQL: WHERE name IS NULL
func filterIsNull(db *gorm.Statement, field, _ string) {
	db.Where(field + " IS NULL ")
}

// filterIsNotNull 不为空 IS NOT NULL操作
// SQL: WHERE name IS NOT NULL
func filterIsNotNull(db *gorm.Statement, field, _ string) {
	db.Where(field + " IS NOT NULL ")
}

// filterContains LIKE 前后模糊查询
// SQL: WHERE name LIKE '%L%';
func filterContains(db *gorm.Statement, field, value string) {
	db.Where(field+" LIKE ? ", "%"+value+"%")
}

// filterInsensitiveContains ILIKE 前后模糊查询
// SQL: WHERE name ILIKE '%L%';
// only support postgresql
func filterInsensitiveContains(db *gorm.Statement, field, value string) {
	db.Where(field+" ILIKE ? ", "%"+value+"%")
}

// filterStartsWith LIKE 前缀+模糊查询
// SQL: WHERE name LIKE 'La%';
func filterStartsWith(db *gorm.Statement, field, value string) {
	db.Where(field+" LIKE ?", value+"%")
}

// filterInsensitiveStartsWith ILIKE 前缀+模糊查询
// SQL: WHERE name ILIKE 'La%';
func filterInsensitiveStartsWith(db *gorm.Statement, field, value string) {
	db.Where(field+" ILIKE ?", value+"%")
}

// filterEndsWith LIKE 后缀+模糊查询
// SQL: WHERE name LIKE '%a';
func filterEndsWith(db *gorm.Statement, field, value string) {
	db.Where(field+" LIKE ?", "%"+value)
}

// filterInsensitiveEndsWith ILIKE 后缀+模糊查询
// SQL: WHERE name ILIKE '%a';
func filterInsensitiveEndsWith(db *gorm.Statement, field, value string) {
	db.Where(field+" ILIKE ?", "%"+value)
}

// filterExact LIKE 操作 精确比对
// SQL: WHERE name LIKE 'a';
func filterExact(db *gorm.Statement, field, value string) {
	db.Where(field+" LIKE ?", value)
}

// filterInsensitiveExact ILIKE 操作 不区分大小写，精确比对
// SQL: WHERE name ILIKE 'a';
func filterInsensitiveExact(db *gorm.Statement, field, value string) {
	db.Where(field+" ILIKE ?", value)
}

// filterRegex 正则查找
// MySQL: WHERE title REGEXP BINARY '^(An?|The) +'
// Oracle: WHERE REGEXP_LIKE(title, '^(An?|The) +', 'c');
// PostgreSQL: WHERE title ~ '^(An?|The) +';
// SQLite: WHERE title REGEXP '^(An?|The) +';
func filterRegex(db *gorm.Statement, field, value string) {
	switch db.Dialector.Name() {
	case dialect.Postgres:
		db.Where(field+" ~ ?", value)

	case dialect.MySQL:
		db.Where(field+" REGEXP BINARY ?", value)

	case dialectOracle:
		db.Where("REGEXP_LIKE(?, ?, 'c')", field, value)
	default:
	}
}

const (
	dialectOracle = "oracle"
)

// filterInsensitiveRegex 正则查找 不区分大小写
// MySQL: WHERE title REGEXP '^(an?|the) +'
// Oracle: WHERE REGEXP_LIKE(title, '^(an?|the) +', 'i');
// PostgreSQL: WHERE title ~* '^(an?|the) +';
// SQLite: WHERE title REGEXP '(?i)^(an?|the) +';
func filterInsensitiveRegex(db *gorm.Statement, field, value string) {
	value = strings.ToLower(value)
	switch db.Dialector.Name() {
	case dialect.Postgres:
		db.Where(field+" ~* ?", value)

	case dialect.MySQL:
		db.Where(field+" REGEXP ?", value)

	case dialectOracle:
		db.Where("REGEXP_LIKE(?, ?, 'i')", field, value)
	default:
	}
}

// filterSearch 全文搜索
// POSTGRESQL 支持
func filterSearch(db *gorm.Statement, _, _ string) {

}

// filterDatePart 时间戳提取日期
// SQL: select extract(quarter from timestamp '2018-08-15 12:10:10');
//func filterDatePart(db *gorm.Statement, datePart, field string) {
//
//	switch db.Dialector.Name() {
//	case dialect.Postgres:
//		str := fmt.Sprintf("EXTRACT('%s' FROM %s)", strings.ToUpper(datePart), s.C(field))
//		db.Where(str)
//
//	case dialect.MySQL:
//		str := fmt.Sprintf("%s(%s)", strings.ToUpper(datePart), s.C(field))
//		b.WriteString(str)
//		//b.Arg(strings.ToLower(value))
//	case dialectOracle:
//	}
//}

// filterDatePartField 日期
//func filterDatePartField(db *gorm.Statement, datePart, field string) string {
//	p := sql.P()
//	switch s.Builder.Dialect() {
//	case dialect.Postgres:
//		str := fmt.Sprintf("EXTRACT('%s' FROM %s)", strings.ToUpper(datePart), s.C(field))
//		p.WriteString(str)
//		break
//
//	case dialect.MySQL:
//		str := fmt.Sprintf("%s(%s)", strings.ToUpper(datePart), s.C(field))
//		p.WriteString(str)
//		break
//	}
//	return p.String()
//}

// filterJsonb 提取JSONB字段
// Postgresql: WHERE ("app_profile"."preferences" ->> 'daily_email') = 'true'
func filterJsonb(db *gorm.Statement, jsonbField, field string) {
	field = stringcase.ToSnakeCase(field)

	switch db.Dialector.Name() {
	case dialect.Postgres:

	case dialect.MySQL:

	}

}

// filterJsonbField JSONB字段
//func filterJsonbField(db *gorm.Statement, jsonbField, field string) string {
//	field = stringcase.ToSnakeCase(field)
//
//	p := sql.P()
//	switch s.Builder.Dialect() {
//	case dialect.Postgres:
//		p.Ident(s.C(field)).WriteString(" ->> ").WriteString("'" + jsonbField + "'")
//		//b.Arg(strings.ToLower(value))
//		break
//
//	case dialect.MySQL:
//		str := fmt.Sprintf("JSON_EXTRACT(%s, '$.%s')", s.C(field), jsonbField)
//		p.WriteString(str)
//		//b.Arg(strings.ToLower(value))
//		break
//	}
//
//	return p.String()
//}
