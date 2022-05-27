package stmt_preparers

import (
	"fmt"
	"strings"
)

const (
	standardPlaceHolderPrefix = `$`
)

// StmtPreparer implementation for the standard ANSI version of SQL.
type StandardPreparer struct{}

func (s *StandardPreparer) PrepareCreateTableFromQueryStmt(table, query string) string {
	return fmt.Sprintf(`CREATE TABLE %s AS %s;`, table, query)
}

func (s *StandardPreparer) PrepareCreateViewStmt(view, query string) string {
	query = removeTrailingSemiColon(query)
	return fmt.Sprintf(`CREATE VIEW %s AS (%s);`, view, query)
}

func (s *StandardPreparer) PrepareCreateViewFromTableStmt(view, table string) string {
	return s.PrepareCreateViewStmt(view, s.PrepareSelectAllStmt(table))
}

func (s *StandardPreparer) PrepareDropTableStmt(table string, checkExists bool) string {
	stmt := `DROP TABLE %s;`
	if checkExists {
		stmt = `DROP TABLE IF EXISTS %s;`
	}
	return fmt.Sprintf(stmt, table)
}

func (s *StandardPreparer) PrepareDropTableCascadeStmt(table string, checkExists bool) string {
	stmt := `DROP TABLE %s CASCADE;`
	if checkExists {
		stmt = `DROP TABLE IF EXISTS %s CASCADE;`
	}
	return fmt.Sprintf(stmt, table)
}

func (s *StandardPreparer) PrepareDropViewStmt(view string, checkExists bool) string {
	stmt := `DROP VIEW %s;`
	if checkExists {
		stmt = `DROP VIEW IF EXISTS %s;`
	}
	return fmt.Sprintf(stmt, view)
}

func (s *StandardPreparer) PrepareSelectAllStmt(table string) string {
	return fmt.Sprintf(`SELECT * FROM %s;`, table)
}

func (s *StandardPreparer) PrepareCountRowsStmt(table string) string {
	return fmt.Sprintf(`SELECT COUNT(*) FROM %s;`, table)
}

func (s *StandardPreparer) PrepareQueryWithLimitStmt(query string, limit int) string {
	query = removeTrailingSemiColon(query)
	temp := generateTempName()
	return fmt.Sprintf(`WITH %s AS (%s) SELECT * FROM %s LIMIT %d;`, temp, query, temp, limit)
}

func (s *StandardPreparer) PrepareInsertStmt(table string, columns []string) string {
	columnString := strings.Join(columns, ", ")
	placeHolders := make([]string, len(columns))
	for i := range placeHolders {
		placeHolders[i] = fmt.Sprintf("%s%d", standardPlaceHolderPrefix, i+1)
	}
	placeHoldersString := strings.Join(placeHolders, ", ")
	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s);`, table, columnString, placeHoldersString)
}

func (s *StandardPreparer) PrepareInsertWithReturnAllStmt(table string, columns []string, allColumns string) string {
	insertStmt := s.PrepareInsertStmt(table, columns)
	insertStmt = removeTrailingSemiColon(insertStmt)
	return fmt.Sprintf("%s RETURNING %s;", insertStmt, allColumns)
}

func (s *StandardPreparer) PrepareUpdateWhereStmt(table string, columns []string, predicateColumn string) string {
	columnString := strings.Join(columns, ", ")
	placeHolders := make([]string, len(columns))
	for i := range placeHolders {
		placeHolders[i] = fmt.Sprintf("%s%d", standardPlaceHolderPrefix, i+1)
	}
	placeHoldersString := strings.Join(placeHolders, ", ")

	updateTemplate := "UPDATE %s SET (%s) = (%s) WHERE %s = %s%d;"
	if len(columns) == 1 {
		// Cannot use multiple-column update format if there is only 1 column to update
		updateTemplate = "UPDATE %s SET %s = %s WHERE %s = %s%d;"
	}

	return fmt.Sprintf(
		updateTemplate,
		table,
		columnString,
		placeHoldersString,
		predicateColumn,
		standardPlaceHolderPrefix,
		len(columns)+1,
	)
}

func (s *StandardPreparer) PrepareUpdateWhereWithReturnAllStmt(table string, columns []string, predicateColumn string, allColumns string) string {
	updateStmt := s.PrepareUpdateWhereStmt(table, columns, predicateColumn)
	updateStmt = removeTrailingSemiColon(updateStmt)
	return fmt.Sprintf("%s RETURNING %s;", updateStmt, allColumns)
}
