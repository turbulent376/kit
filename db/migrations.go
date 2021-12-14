package db

import (
	"database/sql"
	"fmt"
	"git.jetbrains.space/orbi/fcsd/kit/log"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"os"
	"path/filepath"
)

const pgMigrationAdvisoryLockId = 789654123

type Migration interface {
	Up() error
}

type migImpl struct {
	db     *sql.DB
	source string
	logger log.CLoggerFunc
}

func NewMigration(db *sql.DB, source string, logger log.CLoggerFunc) Migration {
	return &migImpl{
		db:     db,
		source: source,
		logger: logger,
	}
}

func (m *migImpl) Up() error {
	l := m.logger().Cmp("db-migration").Mth("up").InfF("applying from %s ...", m.source)

	absPath, _ := filepath.Abs(m.source)
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return ErrGooseFolderNotFound(absPath)
		}
		return ErrGooseFolderOpen(err)
	}

	// lock before migration (applying advisory lock) to guaranty exclusive migration execution
	_, err := m.db.Exec(fmt.Sprintf("select pg_advisory_lock(%d)", pgMigrationAdvisoryLockId))
	if err != nil {
		m.logger().E(ErrGooseMigrationLock(err)).Err()
	}
	// unlock after migration
	defer func() {
		if _, err := m.db.Exec(fmt.Sprintf("select pg_advisory_unlock(%d)", pgMigrationAdvisoryLockId)); err != nil {
			m.logger().E(ErrGooseMigrationUnLock(err)).Err()
		}
	}()

	err = goose.Up(m.db, m.source)
	if err != nil {
		return ErrGooseMigrationUp(err)
	}
	version, err := goose.GetDBVersion(m.db)
	if err != nil {
		return ErrGooseMigrationGetVer(err)
	}
	l.InfF("ok, version: %d", version)
	return nil
}
