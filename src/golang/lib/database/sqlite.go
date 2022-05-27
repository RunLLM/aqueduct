package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"

	stmt "github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// SQLite3 driver from: https://github.com/mattn/go-sqlite3
	sqliteDriver             = "sqlite3"
	sqliteConnectionTemplate = `file:%s?mode=%s&cache=%s`

	defaultSqliteMode  = "rwc"
	defaultSqliteCache = "shared"

	SqliteDatabasePath = "db/aqueduct.db"
)

var DefaultSqliteFile = path.Join(os.Getenv("HOME"), ".aqueduct", "server", SqliteDatabasePath)

type sqliteDatabase struct {
	conf *SqliteConfig
	standardDatabase
	stmt.StandardPreparer
}

type sqliteTransaction struct {
	conf *SqliteConfig
	standardTransaction
	stmt.StandardPreparer
}

// NewSqliteDatabase returns a Database that uses the sqlite3 driver.
func NewSqliteDatabase(conf *SqliteConfig) (Database, error) {
	file := conf.File
	if file == "" {
		file = DefaultSqliteFile
	}

	dsn := fmt.Sprintf(sqliteConnectionTemplate, file, defaultSqliteMode, defaultSqliteCache)
	return newSqliteDatabase(conf, dsn)
}

func NewSqliteInMemoryDatabase(conf *SqliteConfig) (Database, error) {
	dsn := fmt.Sprintf(
		sqliteConnectionTemplate,
		DefaultSqliteFile,
		"memory",
		defaultSqliteCache,
	)
	return newSqliteDatabase(conf, dsn)
}

// newSqliteDatabase returns a sqlite3 driver Database using the DSN provided.
func newSqliteDatabase(conf *SqliteConfig, dsn string) (Database, error) {
	driver, err := sql.Open(sqliteDriver, dsn)
	if err != nil {
		return nil, err
	}

	if err := driver.Ping(); err != nil {
		return nil, err
	}

	return &sqliteDatabase{
		conf: conf,
		standardDatabase: standardDatabase{
			db: driver,
		},
	}, nil
}

func (*sqliteDatabase) Type() Type {
	return SqliteType
}

func (db *sqliteDatabase) Config() *DatabaseConfig {
	return &DatabaseConfig{Type: db.Type(), Sqlite: db.conf}
}

func (sdb *sqliteDatabase) BeginTx(ctx context.Context) (Transaction, error) {
	logQuery("Transaction BEGIN")
	tx, err := sdb.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &sqliteTransaction{
		standardTransaction: standardTransaction{
			tx:     tx,
			nested: false,
		},
	}, nil
}

func (*sqliteTransaction) Type() Type {
	return SqliteType
}

func (tx *sqliteTransaction) Config() *DatabaseConfig {
	return &DatabaseConfig{Type: tx.Type(), Sqlite: tx.conf}
}

func (stx *sqliteTransaction) BeginTx(ctx context.Context) (Transaction, error) {
	// This is already a transaction, so we just return a copy of the receiver.
	// A copy is created so the parent transaction does not get modified.
	tx := &sqliteTransaction{
		standardTransaction: standardTransaction{
			tx:     stx.tx,
			nested: true, // This is a nested transaction.
		},
	}
	return tx, nil
}
