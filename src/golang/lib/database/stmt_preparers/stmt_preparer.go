package stmt_preparers

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
)

// StmtPreparer is the interface that must be implemented in order to provide prepared statements.
type StmtPreparer interface {
	PrepareCreateTableFromQueryStmt(table, query string) string
	PrepareCreateViewStmt(view, query string) string
	PrepareCreateViewFromTableStmt(view, table string) string
	PrepareDropTableStmt(table string, checkExists bool) string
	PrepareDropTableCascadeStmt(table string, checkExists bool) string
	PrepareDropViewStmt(view string, checkExists bool) string
	PrepareSelectAllStmt(table string) string
	PrepareCountRowsStmt(table string) string
	PrepareQueryWithLimitStmt(query string, limit int) string
	PrepareInsertStmt(table string, columns []string) string
	PrepareInsertWithReturnAllStmt(table string, columns []string, allColumns string) string
	PrepareUpdateWhereStmt(table string, columns []string, predicateColumn string) string
	PrepareUpdateWhereWithReturnAllStmt(table string, columns []string, predicateColumn string, allColumns string) string
	// This helper method returns a query fragment that updates exec state blob
	// with the given status and timestamp.
	// This is useful to update the state without deserializing the content.
	// Example: PrepareUpdateExecStateStmt('integration.execution_state', 'succeeded', time.Now())
	// -> '`integration.execution_state = CAST(
	//  json_set(json_set(integration.execution_state, '$.status', 'succeeded'),
	//    '$.timestamps.finished_at', '2023-03-27 14:13PM'
	//  ) AS BLOB)`
	PrepareUpdateExecStateStmt(
		columnAccessPath string,
		status shared.ExecutionStatus,
		timestamp time.Time,
		offset int,
	) (fragment string, args []interface{}, err error)
}
