package database

import (
	"context"
	"database/sql"

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
	defer func() {
		err = rows.Close()
		if err != nil {
			log.Errorf("Error when closing rows: %s", err)
		}
	}()

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
	defer func() {
		err = rows.Close()
		if err != nil {
			log.Errorf("Error when closing rows: %s", err)
		}
	}()

	return scanRows(rows, dest)
}

func (stx *standardTransaction) Rollback(ctx context.Context) error {
	if stx.nested {
		// Aborting a nested transaction should have no effect
		return nil
	}

	if err := stx.tx.Rollback(); err != nil {
		return err
	}

	// Transaction rollback actually happened
	logQuery("Transaction ROLLBACK")
	return nil
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
