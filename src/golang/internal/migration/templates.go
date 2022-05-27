package migration

const (
	sqlScriptHeader = `package _{{.Dir}}

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

`
	sqlScriptBody = `
const (
	sqliteScript = ""
	upPostgresScript = ""
	downPostgresScript = ""
)

func UpPostgres(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, upPostgresScript)
}

func DownPostgres(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, downPostgresScript)
}

func UpSqlite(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, sqliteScript)
}

`

	goScriptHeader = `package _{{.Dir}}

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

`

	goScriptBody = `
func Up(ctx context.Context, db database.Database) error {
	panic("TODO: Implement me")
}

func Down(ctx context.Context, db database.Database) error {
	panic("TODO: Implement me")
}
`
)

func getSqlTemplate() string {
	return sqlScriptHeader + sqlScriptBody
}

func getGoTemplate() string {
	return goScriptHeader + goScriptBody
}
