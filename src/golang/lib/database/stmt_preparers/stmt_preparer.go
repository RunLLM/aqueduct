package stmt_preparers

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
}
