package database

import (
	"context"
	"database/sql"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

// Database implementation for any database that has a driver that implements the
// database/sql interface defined in: https://pkg.go.dev/database/sql.
// The full list of database drivers is here: https://github.com/golang/go/wiki/SQLDrivers.
type standardDatabase struct {
	db *sql.DB
}

// standardTransaction is the Transaction implementation associated with standardDatabase.
type standardTransaction struct {
	tx     *sql.Tx
	nested bool // True if this transaction was created inside of a transaction.
}

func (sdb *standardDatabase) Execute(ctx context.Context, query string, args ...interface{}) error {
	logQuery(query, args...)
	_, err := sdb.db.ExecContext(ctx, query, args...)
	return err
}

func (sdb *standardDatabase) Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	logQuery(query, args...)
	rows, err := sdb.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanRows(rows, dest)
}

func (sdb *standardDatabase) Close() {
	sdb.db.Close()
}

func (stx *standardTransaction) Execute(ctx context.Context, query string, args ...interface{}) error {
	logQuery(query, args...)
	_, err := stx.tx.ExecContext(ctx, query, args...)
	return err
}

func (stx *standardTransaction) Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	logQuery(query, args...)
	rows, err := stx.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanRows(rows, dest)
}

func (stx *standardTransaction) Rollback(ctx context.Context) error {
	if stx.nested {
		// Aborting a nested transaction should have no effect
		return nil
	}

	err := stx.tx.Rollback()
	if !errors.IsError(err, sql.ErrTxDone) {
		// Transaction was not already committed or aborted
		logQuery("Transaction ROLLBACK")
	}
	log.Errorf("Rollback failed: %v.", err)

	return err
}

func (stx *standardTransaction) Commit(ctx context.Context) error {
	if stx.nested {
		// Committing a nested transaction should have no effect
		return nil
	}

	logQuery("Transaction COMMIT")
	return stx.tx.Commit()
}

func (stx *standardTransaction) Close() {
	// Calling close on a transaction does nothing
}
