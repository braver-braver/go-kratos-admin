# test_gorm Command

A throwaway playground used while validating gorm migrations. It provisions a minimal `User` model, runs auto-migration and exercises CRUD operations against a live PostgreSQL instance.

Usage:

```bash
go run ./cmd/test_gorm
```

Update the DSN inside `main.go` before running. This command is not wired into the build and can be removed once the production repositories are fully verified.
