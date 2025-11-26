package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	oracle "github.com/godoes/gorm-oracle"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	defaultOut = filepath.Join("app", "admin", "service", "internal", "data", "gorm")
	tableNames = []string{
		"admin_operation_logs",
		"admin_login_logs",
		"admin_login_restrictions",
		"sys_api_resources",
		"sys_menus",
		"sys_dicts",
		"files",
		"sys_api_resources",
		"notification_messages",
		"departments",
		"notification_message_categories",
		"notification_message_recipients",
		"positions",
		"organizations",
		"private_messages",
		"tenants",
		"user_credentials",
		"sys_roles",
		"users",
		"sys_tasks",
	}
)

type Config struct {
	Driver   string `json:"driver" toml:"driver"`
	DSN      string `json:"dsn" toml:"dsn"`
	Out      string `json:"out" toml:"out"`
	WithJSON bool   `json:"with_json_tag" toml:"with_json_tag"`
}

func main() {
	var (
		configFile = flag.String("config", "", "path to configuration file (JSON or TOML)")
		driver     = flag.String(
			"driver", envOrDefault("GORM_DRIVER", ""), "database driver (postgres/mysql/oracle/sqlite)",
		)
		dsn     = flag.String("dsn", envOrDefault("GORM_DSN", ""), "database DSN")
		outRoot = flag.String(
			"out", envOrDefault("GORM_OUT", defaultOut), "output root directory for generated code",
		)
		withJSON = flag.Bool("with-json-tag", true, "generate json tags in lower_snake_case")
	)

	flag.Parse()

	// Load configuration from file if provided
	var config Config
	if *configFile != "" {
		if err := loadConfig(*configFile, &config); err != nil {
			log.Fatalf("gormgen: load config file failed: %v", err)
		}
		// Override flag values with config values if flags weren't explicitly set
		if !flagIsSet("driver") && config.Driver != "" {
			*driver = config.Driver
		}
		if !flagIsSet("dsn") && config.DSN != "" {
			*dsn = config.DSN
		}
		if !flagIsSet("out") && config.Out != "" {
			*outRoot = config.Out
		}
		if !flagIsSet("with-json-tag") {
			*withJSON = config.WithJSON
		}
	} else {
		// Use environment variables or defaults as fallback
		if *driver == "" {
			*driver = envOrDefault("GORM_DRIVER", "")
		}
		if *dsn == "" {
			*dsn = envOrDefault("GORM_DSN", "")
		}
		if *outRoot == defaultOut {
			*outRoot = envOrDefault("GORM_OUT", defaultOut)
		}
	}

	if strings.TrimSpace(*driver) == "" {
		log.Fatal("gormgen: missing database driver, pass --driver or set GORM_DRIVER or use config file")
	}
	if strings.TrimSpace(*dsn) == "" {
		log.Fatal("gormgen: missing database DSN, pass --dsn or set GORM_DSN or use config file")
	}

	cfg := buildGeneratorConfig(*outRoot, *driver, *withJSON)

	g := gen.NewGenerator(cfg)
	db, err := openDatabase(*driver, *dsn)
	if err != nil {
		log.Fatalf("gormgen: open database failed: %v", err)
	}

	g.UseDB(db)

	if err = applyModels(g, *driver); err != nil {
		log.Fatalf("gormgen: apply models failed: %v", err)
	}

	g.Execute()
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadConfig loads configuration from a JSON or TOML file
func loadConfig(filePath string, config *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Check file extension to determine format
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return json.Unmarshal(data, config)
	case ".toml":
		return toml.Unmarshal(data, config)
	default:
		// Try JSON first, then TOML
		if err := json.Unmarshal(data, config); err == nil {
			return nil
		}
		return toml.Unmarshal(data, config)
	}
}

// flagIsSet checks if a flag was explicitly set
func flagIsSet(name string) bool {
	found := false
	flag.Visit(
		func(f *flag.Flag) {
			if f.Name == name {
				found = true
			}
		},
	)
	return found
}

func buildGeneratorConfig(outRoot, driver string, withJSON bool) gen.Config {
	outQuery := filepath.Join(outRoot, "query")
	outModel := filepath.Join(outRoot, "model")

	cfg := gen.Config{
		OutPath:           outQuery,
		ModelPkgPath:      outModel,
		FieldNullable:     true,
		FieldCoverable:    true,
		FieldSignable:     true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
	}

	if withJSON {
		cfg.WithJSONTagNameStrategy(
			func(columnName string) string {
				return toLowerSnake(columnName)
			},
		)
	}

	// When generating from Oracle, ensure tables are treated in uppercase to match driver behaviour.
	if strings.EqualFold(driver, "oracle") {
		cfg.WithTableNameStrategy(
			func(tableName string) string {
				return strings.ToUpper(tableName)
			},
		)
		cfg.WithModelNameStrategy(
			func(tableName string) string {
				return toCamel(strings.ToLower(tableName))
			},
		)
		cfg.WithFileNameStrategy(
			func(tableName string) string {
				return toLowerSnake(tableName)
			},
		)
	}

	return cfg
}

func openDatabase(driver, dsn string) (*gorm.DB, error) {
	var (
		dialector gorm.Dialector
		namer     schema.NamingStrategy
	)

	switch strings.ToLower(driver) {
	case "postgres", "postgresql":
		dialector = postgres.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	case "oracle":
		dialector = oracle.New(
			oracle.Config{
				DSN:                 dsn,
				NamingCaseSensitive: false,
			},
		)
		// Oracle defaults to uppercase identifiers.
		namer = schema.NamingStrategy{
			SingularTable: false,
			NameReplacer:  strings.NewReplacer(`"`, ""),
		}
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	cfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	if dialector != nil && namer != (schema.NamingStrategy{}) {
		cfg.NamingStrategy = namer
	}

	return gorm.Open(dialector, cfg)
}

func applyModels(g *gen.Generator, driver string) error {
	models := make([]interface{}, 0, len(tableNames))
	for _, table := range tableNames {
		models = append(models, g.GenerateModel(table))
	}

	g.ApplyBasic(models...)

	if strings.EqualFold(driver, "oracle") {
		// Oracle driver returns upper-case identifiers; ensure json tags remain consistent.
		g.WithJSONTagNameStrategy(
			func(columnName string) string {
				return toLowerSnake(columnName)
			},
		)
	}

	return nil
}

func toLowerSnake(input string) string {
	if input == "" {
		return input
	}

	var builder strings.Builder
	builder.Grow(len(input) + 4)

	for i, r := range input {
		if isUpper(r) {
			if i > 0 {
				builder.WriteByte('_')
			}
			builder.WriteRune(toLower(r))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func toCamel(input string) string {
	if input == "" {
		return input
	}

	segments := strings.FieldsFunc(input, func(r rune) bool { return r == '_' || r == '-' })
	for i := range segments {
		segments[i] = capitalize(segments[i])
	}
	return strings.Join(segments, "")
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	if len(s) == 1 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
