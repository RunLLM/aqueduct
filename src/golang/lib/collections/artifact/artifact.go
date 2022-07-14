package artifact

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type DBArtifact struct {
	Id          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Spec        Spec      `db:"spec" json:"spec"`
}

type Reader interface {
	Exists(ctx context.Context, id uuid.UUID, db database.Database) (bool, error)
	GetArtifact(ctx context.Context, id uuid.UUID, db database.Database) (*DBArtifact, error)
	GetArtifacts(ctx context.Context, ids []uuid.UUID, db database.Database) ([]DBArtifact, error)
	GetArtifactsByWorkflowDagId(
		ctx context.Context,
		workflowDagId uuid.UUID,
		db database.Database,
	) ([]DBArtifact, error)
	ValidateArtifactOwnership(
		ctx context.Context,
		organizationId string,
		artifactId uuid.UUID,
		db database.Database,
	) (bool, error)
}

type Writer interface {
	CreateArtifact(
		ctx context.Context,
		name string,
		description string,
		spec *Spec,
		db database.Database,
	) (*DBArtifact, error)
	UpdateArtifact(
		ctx context.Context,
		id uuid.UUID,
		changes map[string]interface{},
		db database.Database,
	) (*DBArtifact, error)
	DeleteArtifact(ctx context.Context, id uuid.UUID, db database.Database) error
	DeleteArtifacts(ctx context.Context, ids []uuid.UUID, db database.Database) error
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
