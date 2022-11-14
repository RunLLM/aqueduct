package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
)

// Watcher defines all of the database operations that can be performed for a Watcher.
type Watcher interface {
	watcherReader
	watcherWriter
}

type watcherReader interface{}

type watcherWriter interface {
	// Create inserts a new Watcher with the specified fields.
	Create(
		ctx context.Context,
		workflowID uuid.UUID,
		userID uuid.UUID,
		DB database.Database,
	) (*models.Watcher, error)

	// Delete deletes the Watcher for the User and Workflow specified.
	Delete(
		ctx context.Context,
		workflowID uuid.UUID,
		userID uuid.UUID,
		DB database.Database,
	) error

	// DeleteByWorkflow deletes all Watchers for the Workflow specified.
	DeleteByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) error
}
