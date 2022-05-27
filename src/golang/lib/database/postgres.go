package database

import (
	"context"
	"database/sql"
	"fmt"

	stmt "github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	// Postgres driver from: https://github.com/jackc/pgx
	postgresDriver             = "pgx"
	postgresConnectionTemplate = `postgres://%s:%s@%s:%s/%s`

	defaultPostgresPort = "5432"
)

type postgresDatabase struct {
	conf *PostgresConfig
	standardDatabase
	stmt.StandardPreparer
}

type postgresTransaction struct {
	conf *PostgresConfig
	standardTransaction
	stmt.StandardPreparer
}

// NewPostgresDatabase returns a Database that uses the pgx driver.
// It uses the default Postgres port of 5432.
func NewPostgresDatabase(conf *PostgresConfig) (Database, error) {
	dsn := fmt.Sprintf(
		postgresConnectionTemplate,
		conf.UserName,
		conf.Password,
		conf.Address,
		defaultPostgresPort,
		conf.Database,
	)
	return newPostgresDatabase(conf, dsn)
}

// NewPostgresDatabaseWithPort returns a Database that uses the pgx driver.
// It uses the port specified.
func NewPostgresDatabaseWithPort(conf *PostgresConfig) (Database, error) {
	dsn := fmt.Sprintf(
		postgresConnectionTemplate,
		conf.UserName,
		conf.Password,
		conf.Address,
		conf.Port,
		conf.Database,
	)
	return newPostgresDatabase(conf, dsn)
}

// newPostgresDatabase returns a pgx driver Database using the DSN provided.
func newPostgresDatabase(conf *PostgresConfig, dsn string) (Database, error) {
	driver, err := sql.Open(postgresDriver, dsn)
	if err != nil {
		return nil, err
	}

	err = driver.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresDatabase{
		conf: conf,
		standardDatabase: standardDatabase{
			db: driver,
		},
	}, nil
}

func (*postgresDatabase) Type() Type {
	return PostgresType
}

func (db *postgresDatabase) Config() *DatabaseConfig {
	return &DatabaseConfig{Type: db.Type(), Postgres: db.conf}
}

func (pdb *postgresDatabase) BeginTx(ctx context.Context) (Transaction, error) {
	logQuery("Transaction BEGIN")
	tx, err := pdb.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &postgresTransaction{
		conf: pdb.conf,
		standardTransaction: standardTransaction{
			tx:     tx,
			nested: false,
		},
	}, nil
}

func (*postgresTransaction) Type() Type {
	return PostgresType
}

func (tx *postgresTransaction) Config() *DatabaseConfig {
	return &DatabaseConfig{Type: tx.Type(), Postgres: tx.conf}
}

func (ptx *postgresTransaction) BeginTx(ctx context.Context) (Transaction, error) {
	// This is already a transaction, so we just return a copy of the receiver.
	// A copy is created so the parent transaction does not get modified.
	tx := &postgresTransaction{
		standardTransaction: standardTransaction{
			tx:     ptx.tx,
			nested: true, // This is a nested transaction.
		},
	}
	return tx, nil
}
