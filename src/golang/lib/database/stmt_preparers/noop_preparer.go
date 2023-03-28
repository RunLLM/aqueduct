package stmt_preparers

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
)

// StmtPreparer implementation for the standard ANSI version of SQL.
type NoopPreparer struct{}

func (*NoopPreparer) PrepareCreateTableFromQueryStmt(table, query string) string {
	return ""
}

func (*NoopPreparer) PrepareCreateViewStmt(view, query string) string {
	return ""
}

func (*NoopPreparer) PrepareCreateViewFromTableStmt(view, table string) string {
	return ""
}

func (*NoopPreparer) PrepareDropTableStmt(table string, checkExists bool) string {
	return ""
}

func (*NoopPreparer) PrepareDropTableCascadeStmt(table string, checkExists bool) string {
	return ""
}

func (*NoopPreparer) PrepareDropViewStmt(view string, checkExists bool) string {
	return ""
}

func (*NoopPreparer) PrepareSelectAllStmt(table string) string {
	return ""
}

func (*NoopPreparer) PrepareCountRowsStmt(table string) string {
	return ""
}

func (*NoopPreparer) PrepareQueryWithLimitStmt(query string, limit int) string {
	return ""
}

func (*NoopPreparer) PrepareInsertStmt(table string, columns []string) string {
	return ""
}

func (*NoopPreparer) PrepareInsertWithReturnAllStmt(table string, columns []string, allColumns string) string {
	return ""
}

func (*NoopPreparer) PrepareUpdateWhereStmt(table string, columns []string, predicateColumn string) string {
	return ""
}

func (*NoopPreparer) PrepareUpdateWhereWithReturnAllStmt(
	table string,
	columns []string,
	predicateColumn string,
	allColumns string,
) string {
	return ""
}

func (s *NoopPreparer) PrepareUpdateExecStateStmt(
	columnAccessPath string,
	status shared.ExecutionStatus,
	timestamp time.Time,
	offset int,
) (fragment string, args []interface{}, err error) {
	return "", nil, nil
}
