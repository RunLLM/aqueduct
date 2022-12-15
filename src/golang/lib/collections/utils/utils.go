package utils

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// CountResult is useful when querying for `COUNT(*)`
type CountResult struct {
	Count int `db:"count"`
}

func prepareUpdateRecord(
	changedColumns map[string]interface{},
	predicateColumnVal interface{},
) ([]string, []interface{}) {
	updateColumns := make([]string, 0, len(changedColumns))
	args := make([]interface{}, 0, len(changedColumns))

	for column, arg := range changedColumns {
		updateColumns = append(updateColumns, column)
		args = append(args, arg)
	}

	args = append(args, predicateColumnVal) // The final placeholder is for the WHERE clause parameter

	return updateColumns, args
}

// UpdateRecordToDest execute an UPDATE statement to modify the specified columns in `table`.
// It writes the modified record to `dest`, which should be a pointer to a struct.
// For now, the function only performs UPDATE statements with a single WHERE clause for the column
// `predicateColumn` that has value `predicateColumnVal`.
func UpdateRecordToDest(
	ctx context.Context,
	dest interface{},
	changedColumns map[string]interface{},
	table string,
	predicateColumn string,
	predicateColumnVal interface{},
	allColumns string,
	db database.Database,
) error {
	updateColumns, args := prepareUpdateRecord(changedColumns, predicateColumnVal)
	updateStmt := db.PrepareUpdateWhereWithReturnAllStmt(table, updateColumns, predicateColumn, allColumns)

	return db.Query(ctx, dest, updateStmt, args...)
}

// UpdateRecord execute an UPDATE statement to modify the specified columns in `table`.
// For now, the function only performs UPDATE statements with a single WHERE clause for the column
// `predicateColumn` that has value `predicateColumnVal`.
func UpdateRecord(
	ctx context.Context,
	changedColumns map[string]interface{},
	table string,
	predicateColumn string,
	predicateColumnVal interface{},
	db database.Database,
) error {
	updateColumns, args := prepareUpdateRecord(changedColumns, predicateColumnVal)
	updateStmt := db.PrepareUpdateWhereStmt(table, updateColumns, predicateColumn)

	return db.Execute(ctx, updateStmt, args...)
}

func NoopInterfaceErrorHandling(throwError bool) error {
	var err errors.DropboxError
	if throwError {
		return errors.New("The noop database interface is being used, which should not happen.")
	}
	return nil
}

// Checks if the given UUID exists in the given table
func IdExistsInTable(
	ctx context.Context,
	id uuid.UUID,
	tableName string,
	db database.Database,
) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(1) AS count FROM %s WHERE id = $1", tableName)
	var count CountResult
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
		exists, err := IdExistsInTable(ctx, id, tableName, db)
		if err != nil {
			return uuid.Nil, err
		}

		if !exists {
			return id, nil
		}
	}
}

func ValidateNodeOwnership(
	ctx context.Context,
	organizationId string,
	nodeId uuid.UUID,
	db database.Database,
) (bool, error) {
	// Get the count of rows where the nodeId has an edge (is the `from_id` or `to_id`)
	// in a workflow DAG belonging to a workflow owned by the the user's organization.
	query := `
		SELECT COUNT(*) AS count 
		FROM workflow_dag_edge, workflow_dag, workflow, app_user 
		WHERE workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
		workflow_dag.workflow_id = workflow.id AND workflow.user_id = app_user.id AND 
		app_user.organization_id = $1 AND (workflow_dag_edge.from_id = $2 OR workflow_dag_edge.to_id = $2)`
	var count CountResult

	err := db.Query(ctx, &count, query, organizationId, nodeId)
	if err != nil {
		return false, err
	}

	return count.Count >= 1, nil
}
