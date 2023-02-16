package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

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
	DB database.Database,
) error {
	updateColumns, args := prepareUpdateRecord(changedColumns, predicateColumnVal)
	updateStmt := DB.PrepareUpdateWhereWithReturnAllStmt(table, updateColumns, predicateColumn, allColumns)

	return DB.Query(ctx, dest, updateStmt, args...)
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
	DB database.Database,
) error {
	updateColumns, args := prepareUpdateRecord(changedColumns, predicateColumnVal)
	updateStmt := DB.PrepareUpdateWhereStmt(table, updateColumns, predicateColumn)

	return DB.Execute(ctx, updateStmt, args...)
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
