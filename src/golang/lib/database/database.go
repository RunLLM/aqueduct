package database

import (
	"context"

	stmt "github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
)

const (
	// SQL error codes
	ErrCodeTableDoesNotExist = "SQLSTATE 42P01"
)

var (
	ErrNoRows            = errors.New("Query returned no rows.")
	ErrUnsupportedDbType = errors.New("DB Type is not supported")
)

// Database is the interface that must be implemented by any database driver.
type Database interface {
	// Type returns the SQL dialect used by this database.
	Type() Type

	// The entire database config
	Config() *DatabaseConfig

	// Execute executes `query`. The args are for any placeholder params in the query.
	// It returns an error, if any. This method should only be used to execute SQL queries that have no response.
	Execute(ctx context.Context, query string, args ...interface{}) error

	// Query executes `query` and scans the result into `dest`. `dest` should either be a pointer
	// to a struct or a pointer to a slice of structs. The args are for any placeholder params in the query.
	// It returns ErrNoRows if no row is found when scanning to a struct pointer. If no rows are found
	// when scanning to a slice of structs pointer, no error is returned, but the slice will be empty.
	Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// Close should be used to terminate a connection. `Database` is designed to be a long-lived object, but
	// it should be closed at the end of its lifetime to prevent connection failures and concurrency limits.
	Close()

	// BeginTx starts a transaction and returns a Transaction object.
	BeginTx(ctx context.Context) (Transaction, error)

	// StmtPreparer is embedded in order to expose prepared statements.
	stmt.StmtPreparer
}

// Transaction defines an interface for performing Database operations inside of a transaction.
type Transaction interface {
	// Rollback aborts the transaction. If the transaction was already committed,
	// Rollback acts as a no-op, so it is safe to use as a deferred statement. See TxnRollbackIgnoreErr().
	Rollback(ctx context.Context) error

	// Commit commits the transaction.
	Commit(ctx context.Context) error

	// A Transaction should be able to perform all Database operations.
	Database
}

// Usage: `defer TxnRollbackIgnoreErr(ctx, txn)`
// Because its bad practice for a defer to return an error.
func TxnRollbackIgnoreErr(ctx context.Context, txn Transaction) {
	_ = txn.Rollback(ctx)
}

func NewDatabase(conf *DatabaseConfig) (Database, error) {
	if conf.Type == PostgresType {
		if conf.Postgres == nil {
			return nil, errors.New("Invalid database config, expected `Postgres` field to be set.")
		}

		return NewPostgresDatabase(conf.Postgres)
	}

	if conf.Type == SqliteType {
		sqliteConf := conf.Sqlite
		if sqliteConf == nil {
			sqliteConf = &SqliteConfig{}
		}

		return NewSqliteDatabase(sqliteConf)
	}

	return nil, ErrUnsupportedDbType
}
