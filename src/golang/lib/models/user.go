package models

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// This is a test

const (
	UserTable = "app_user"

	// User column names
	UserID      = "id"
	UserEmail   = "email"
	UserOrgID   = "organization_id"
	UserRole    = "role"
	UserAPIKey  = "api_key"
	UserAuth0ID = "auth0_id"
)

// A User maps to the app_user table.
type User struct {
	ID      uuid.UUID `db:"id" json:"id"`
	Email   string    `db:"email" json:"email"`
	OrgID   string    `db:"organization_id" json:"organization_id"`
	Role    string    `db:"role" json:"role"`
	APIKey  string    `db:"api_key" json:"api_key"`
	Auth0ID string    `db:"auth0_id" json:"auth0_id"`
}

// UserCols returns a comma-separated string of all User columns.
func UserCols() string {
	return strings.Join(allUserCols(), ",")
}

// UserColsWithPrefix returns a comma-separated string of all
// User columns prefixed by the table name.
func UserColsWithPrefix() string {
	cols := allUserCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", UserTable, col)
	}

	return strings.Join(cols, ",")
}

func allUserCols() []string {
	return []string{
		UserID,
		UserEmail,
		UserOrgID,
		UserRole,
		UserAPIKey,
		UserAuth0ID,
	}
}
