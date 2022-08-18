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
	sqliteDriver = "sqlite3"

	SqliteDatabasePath = "db/aqueduct.db"
)

var DefaultSqliteFile = path.Join(os.Getenv("HOME"), ".aqueduct", "server", SqliteDatabasePath)

var defaultSqliteOptions = map[string]string{
	"mode":          "rwc",
	"cache":         "shared",
	"_journal_mode": "WAL",  // Enable Write-Ahead logging.
	"_busy_timeout": "3000", // Wait for a bit on database locks before giving up.
}

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

// Create Data Source String with which to configure this Sqlite driver.
func createDsn(file string, sqliteOptions map[string]string) string {
	dsn := fmt.Sprintf("file:%s?", file)
	for k, v := range sqliteOptions {
		dsn += fmt.Sprintf("%s=%s&", k, v)
	}

	// Remove the hanging '&'
	if len(sqliteOptions) > 0 {
		dsn = dsn[:len(dsn)-1]
	}
	return dsn
}

// NewSqliteDatabase returns a Database that uses the sqlite3 driver.
func NewSqliteDatabase(conf *SqliteConfig) (Database, error) {
	file := conf.File
	if file == "" {
		file = DefaultSqliteFile
	}

	return newSqliteDatabase(conf, createDsn(file, defaultSqliteOptions))
}

func NewSqliteInMemoryDatabase(conf *SqliteConfig) (Database, error) {
	dsn := createDsn(DefaultSqliteFile, map[string]string{
		"mode":  "memory",
		"cache": "shared",
	})
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
