package database

import (
	"context"

	stmt "github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
)

type noopDatabase struct {
	stmt.NoopPreparer
}

type noopTransaction struct {
	noopDatabase
}

func NewNoopDatabase() Database {
	return &noopDatabase{}
}

func (*noopDatabase) Config() *DatabaseConfig {
	return &DatabaseConfig{Type: NoopType}
}

func (db *noopDatabase) Type() Type {
	return db.Config().Type
}

func (*noopDatabase) Execute(ctx context.Context, query string, args ...interface{}) error {
	return nil
}

func (*noopDatabase) Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (*noopDatabase) Close() {}

func (*noopDatabase) BeginTx(ctx context.Context) (Transaction, error) {
	return &noopTransaction{}, nil
}

func (*noopTransaction) Rollback(ctx context.Context) error {
	return nil
}

func (*noopTransaction) Commit(ctx context.Context) error {
	return nil
}
