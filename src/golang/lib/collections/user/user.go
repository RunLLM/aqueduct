package user

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

const (
	apiKeyLength = 60
)

type User struct {
	Id             uuid.UUID `db:"id"`
	Email          string    `db:"email"`
	OrganizationId string    `db:"organization_id"`
	Role           string    `db:"role"`
	ApiKey         string    `db:"api_key"`
	Auth0Id        string    `db:"auth0_id"`
}

type Reader interface {
	GetUser(ctx context.Context, id uuid.UUID, db database.Database) (*User, error)
	GetUserFromApiKey(ctx context.Context, apiKey string, db database.Database) (*User, error)
	GetUserFromEmail(ctx context.Context, email string, db database.Database) (*User, error)
	GetUserFromAuth0Id(ctx context.Context, auth0Id string, db database.Database) (*User, error)
	GetOrganizationAdmin(ctx context.Context, organizationId string, db database.Database) (*User, error)
	GetUsersInOrganization(ctx context.Context, organizationId string, db database.Database) ([]User, error)
	GetWatchersByWorkflowId(ctx context.Context, workflowId uuid.UUID, db database.Database) ([]User, error)
}

type Writer interface {
	CreateUser(
		ctx context.Context,
		email, organizationId, role, auth0Id string,
		db database.Database,
	) (*User, error)
	CreateUserWithApiKey(
		ctx context.Context,
		email, organizationId, role, auth0Id, apiKey string,
		db database.Database,
	) (*User, error)
	CreateOrganizationAdmin(
		ctx context.Context,
		organizationId, organizationName string,
		db database.Database,
	) (*User, error)
	ResetApiKey(ctx context.Context, id uuid.UUID, db database.Database) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID, db database.Database) error
	DeleteUserWithAuth0Id(ctx context.Context, auth0Id string, db database.Database) error
}

func NewReader(dbConf *database.DatabaseConfig) (Reader, error) {
	if dbConf.Type == database.PostgresType {
		return newPostgresReader(), nil
	}

	if dbConf.Type == database.SqliteType {
		return newSqliteReader(), nil
	}

	return nil, database.ErrUnsupportedDbType
}

func NewWriter(dbConf *database.DatabaseConfig) (Writer, error) {
	if dbConf.Type == database.PostgresType {
		return newPostgresWriter(), nil
	}

	if dbConf.Type == database.SqliteType {
		return newSqliteWriter(), nil
	}

	return nil, database.ErrUnsupportedDbType
}
