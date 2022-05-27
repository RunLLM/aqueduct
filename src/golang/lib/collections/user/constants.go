package user

import (
	"fmt"
	"strings"
)

const (
	tableName = "app_user"

	// User table column names
	IdColumn             = "id"
	EmailColumn          = "email"
	OrganizationIdColumn = "organization_id"
	RoleColumn           = "role"
	ApiKeyColumn         = "api_key"
	Auth0IdColumn        = "auth0_id"
)

// Returns a joined string of all User columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			EmailColumn,
			OrganizationIdColumn,
			RoleColumn,
			ApiKeyColumn,
			Auth0IdColumn,
		},
		",",
	)
}

// Returns a joined string of all User columns prefixed by table name.
func allColumnsWithPrefix() string {
	joinedColumns := allColumns()
	columns := strings.Split(joinedColumns, ",")

	columnsWithPrefix := make([]string, 0, len(columns))
	for _, col := range columns {
		columnsWithPrefix = append(
			columnsWithPrefix,
			fmt.Sprintf(
				"%s.%s",
				tableName,
				col,
			))
	}
	return strings.Join(columnsWithPrefix, ",")
}

type Role string

const (
	AdminRole Role = "admin"
)
