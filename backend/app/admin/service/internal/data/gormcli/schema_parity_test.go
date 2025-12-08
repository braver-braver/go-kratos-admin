package gormcli

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"kratos-admin/app/admin/service/internal/data/gormcli/model"
)

// Ensures all generated tables exist in the target database.
func TestSchemaParity_TablesExist(t *testing.T) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	tables := []string{
		model.TableNameSysAdminOperationLog,
		model.TableNameSysAdminLoginLog,
		model.TableNameSysAdminLoginRestriction,
		model.TableNameSysAPIResource,
		model.TableNameSysDepartment,
		model.TableNameSysDictEntry,
		model.TableNameSysDictType,
		model.TableNameSysLanguage,
		model.TableNameSysMenu,
		model.TableNameSysOrganization,
		model.TableNameSysPosition,
		model.TableNameSysRole,
		model.TableNameSysRoleAPI,
		model.TableNameSysRoleDept,
		model.TableNameSysRoleMenu,
		model.TableNameSysRoleOrg,
		model.TableNameSysRolePosition,
		model.TableNameSysTask,
		model.TableNameSysTenant,
		model.TableNameSysUser,
		model.TableNameSysUserCredential,
		model.TableNameSysUserPosition,
		model.TableNameSysUserRole,
		model.TableNameInternalMessage,
		model.TableNameInternalMessageCategory,
		model.TableNameInternalMessageRecipient,
		model.TableNameFile,
	}

	migrator := db.Migrator()
	for _, table := range tables {
		if !migrator.HasTable(table) {
			t.Fatalf("missing table %s", table)
		}
		cols, err := migrator.ColumnTypes(table)
		if err != nil {
			t.Fatalf("read columns for %s: %v", table, err)
		}
		if len(cols) == 0 {
			t.Fatalf("table %s has no columns detected", table)
		}
	}
}
