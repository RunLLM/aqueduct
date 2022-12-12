package stmt_preparers

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// This helper function generates placeholder for arg list.
// It takes a length and generates a query template $2,$3,...,$n based on length and startIdx.
//
// Reference: https://stackoverflow.com/questions/45351644/golang-slice-in-mysql-query-with-where-in-clause
// Example usage, fetching from a list of ids:
//
//	query := "SELECT * FROM workflow WHERE id IN (" + GenerateArgsList(len(ids)) + ")"
//	var workflows []Workflow
//	args := CastIdsListToInterfaceList(ids)
//	err := db.QueryToDest(ctx, &workflows, query, args...)
func GenerateArgsList(length int, startIdx int) string {
	args := ""

	for idx := startIdx; idx < startIdx+length; idx++ {
		args += fmt.Sprintf("$%d", idx)
		if idx != startIdx+length-1 {
			args += ","
		}
	}

	return args
}

// Explicitly convert an id list to interface list,
// so that we can use them as QueryToDest args
func CastIdsListToInterfaceList(ids []uuid.UUID) []interface{} {
	args := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	return args
}

// Removes trailing semicolon if it exists
func removeTrailingSemiColon(s string) string {
	if s[len(s)-1] == ';' {
		return s[:len(s)-1]
	}
	return s
}

// Generates a temporary table name
func generateTempName() string {
	uid, _ := uuid.NewUUID()
	s := uid.String()
	s = strings.ReplaceAll(s, "-", "") // Remove dashes to avoid SQL naming issues
	return fmt.Sprintf("temp%s", s)
}

// DoubleQuoteIdentifier wraps the provided identifier in double quotes.
// This util function is helpful when the identifier is case sensitive.
// WARNING: If a table is created with a double quoted identifier, each subsequent
// query that references the table name must use a doubled quoted identifier.
func DoubleQuoteIdentifier(identifier string) string {
	return fmt.Sprintf(`"%s"`, identifier)
}
