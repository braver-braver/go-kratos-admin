# gormgen Command

This command drives gorm/gen code generation for the admin service data layer. Launch it via the top-level Make targets:

```bash
make gorm-gen GORM_DRIVER=postgres GORM_DSN="host=... sslmode=disable"
```

## Configuration File Support

You can provide configuration via JSON or TOML files using the `--config` flag:

```bash
go run cmd/gormgen/main.go --config path/to/config.toml
go run cmd/gormgen/main.go --config path/to/config.json
```

Configuration file support takes precedence over environment variables but can be overridden by command-line flags.

## Available Configuration Options

### Command-line Flags:

- `--config`: Path to configuration file (JSON or TOML)
- `--driver` / `GORM_DRIVER`: database driver (`postgres`, `mysql`, `oracle`, `sqlite` for local tests).
- `--dsn` / `GORM_DSN`: DSN for schema introspection.
- `--out` / `GORM_OUT`: output directory (defaults to `app/admin/service/internal/data/gorm`).
- `--with-json-tag`: generate json tags in lower_snake_case (default: true).

### Configuration File Fields:

- `driver`: database driver (`postgres`, `mysql`, `oracle`, `sqlite`)
- `dsn`: DSN for schema introspection
- `out`: output directory (defaults to `app/admin/service/internal/data/gorm`)
- `with_json_tag`: boolean to enable json tags in lower_snake_case

## Examples

### Using Configuration File:
```bash
go run cmd/gormgen/main.go --config config.toml
```

### Configuration File Override:
```bash
go run cmd/gormgen/main.go --config config.toml --driver mysql
```

### Using Environment Variables:
```bash
GORM_DRIVER=postgres GORM_DSN="host=localhost user=gorm password=gorm dbname=gorm" go run cmd/gormgen/main.go
```

Oracle-specific handling automatically uppercases table names to avoid case-sensitivity issues.
