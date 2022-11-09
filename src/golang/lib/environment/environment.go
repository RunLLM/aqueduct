package environment

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type Environment struct {
	Id            uuid.UUID `json:"id"`
	PythonVersion string    `json:"python_version"`
	Dependencies  string    `json:"dependencies"`

	// TODO: Add environment object DB reader
}

// `CreateDB()` should initialize the DB object.
func (e *Environment) CreateDB() error {
	return nil
}

// `DeleteFromDB()` should remove this environment from DB
func (e *Environment) DeleteFromDB() error {
	return nil
}

// `Hash()` generates a hash based on the environment's
// dependency set and python version.
func (e *Environment) Hash() string {
	return ""
}

// `NewFromDB` constructs an env object from DB.
func NewFromDB(
	ctx context.Context,
	// dbEnv DBEnvironment,
	// reader db_environment.Reader,
	db database.Database,
) (*Environment, error) {
	return nil, nil
}

// `Exists()` check if a given environment has a match in DB.
// 'match' is defined as having the same python version and dependency set.
//
// Returns the environment object from the DB row if exists.
// Otherwise, returns nil.
func Exists(ctx context.Context, env *Environment) (*Environment, error) {
	return nil, nil
}
