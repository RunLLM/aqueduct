package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

// countResult is useful when querying for `COUNT(*)`
type countResult struct {
	Count int `db:"count"`
}

// Checks if the given UUID exists in the given table
func IDExistsInTable(
	ctx context.Context,
	id uuid.UUID,
	tableName string,
	db database.Database,
) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(1) AS count FROM %s WHERE id = $1", tableName)
	var count countResult
	err := db.Query(ctx, &count, query, id)
	return count.Count > 0, err
}

// GenerateUniqueUUID generates a unique UUID for the `id`
// column in the table specified. It also returns an error, if any.
func GenerateUniqueUUID(
	ctx context.Context,
	tableName string,
	db database.Database,
) (uuid.UUID, error) {
	for {
		id := uuid.New()
		exists, err := IDExistsInTable(ctx, id, tableName, db)
		if err != nil {
			return uuid.Nil, err
		}

		if !exists {
			return id, nil
		}
	}
}

func validateNodeOwnership(
	ctx context.Context,
	orgID string,
	nodeID uuid.UUID,
	DB database.Database,
) (bool, error) {
	// Get the count of rows where the nodeId has an edge (is the `from_id` or `to_id`)
	// in a workflow DAG belonging to a workflow owned by the the user's organization.
	query := `
		SELECT COUNT(*) AS count 
		FROM workflow_dag_edge, workflow_dag, workflow, app_user 
		WHERE workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
		workflow_dag.workflow_id = workflow.id AND workflow.user_id = app_user.id AND 
		app_user.organization_id = $1 AND (workflow_dag_edge.from_id = $2 OR workflow_dag_edge.to_id = $2)`
	var count countResult

	err := DB.Query(ctx, &count, query, orgID, nodeID)
	if err != nil {
		return false, err
	}

	return count.Count >= 1, nil
}
