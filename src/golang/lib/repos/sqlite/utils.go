package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
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

// generateUpdateExecStateSnippet returns a query fragment that updates exec state blob
// with the given status and timestamp.
// This is useful to update the state without deserializing the content.
// Example: generateUpdateExecStateSnippet('integration.execution_state', 'succeeded', time.Now())
// -> '`integration.execution_state = CAST(
//
//	json_set(json_set(integration.execution_state, '$.status', 'succeeded'),
//	  '$.timestamps.finished_at', '2023-03-27 14:13PM'
//	) AS BLOB)`
func generateUpdateExecStateSnippet(
	columnAccessPath string,
	status shared.ExecutionStatus,
	timestamp time.Time,
	offset int,
) (fragment string, args []interface{}, err error) {
	timestampField, err := shared.ExecutionTimestampsJsonFieldByStatus(status)
	if err != nil {
		return "", nil, err
	}

	// (likawind) If passing the timestamp object directly to query, it somehow bypass the
	// drivers serialization and results in values that couldn't be deserialized back.
	// The reason is not clear, but using `MarshalText()` to pre-serialize the value
	// appears to solve the issue.
	timestampValue, err := timestamp.MarshalText()
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf(`%s = CAST(
		json_set(
			json_set(%s, '$.status', $%d),
			'$.timestamps.%s',
			$%d
		) AS BLOB)`,
			columnAccessPath,
			columnAccessPath,
			offset+1,
			timestampField,
			offset+2,
		), []interface{}{
			status, string(timestampValue),
		}, nil
}
