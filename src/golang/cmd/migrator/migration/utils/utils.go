package utils

import (
	"context"
	"io/ioutil"

	"github.com/aqueducthq/aqueduct/lib/database"
)

// InvokeSqlScript takes in a file path pointing to a .sql file and executes it
// using the database.Database object provided.
func InvokeSqlScript(ctx context.Context, filePath string, db database.Database) error {
	statements, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return db.Execute(ctx, string(statements))
}
