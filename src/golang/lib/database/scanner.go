package database

import (
	"database/sql"
	"reflect"

	"github.com/aqueducthq/aqueduct/lib/errors"
	log "github.com/sirupsen/logrus"
)

const (
	// Specifies the column name of a field struct in order to scan sql.Rows to a struct.
	fieldTag = "db"
)

// Helper function to scan `rows` into `dest`.
// This function requires understanding of Go's reflect package, which allows programs to
// dynamically interact with variables, structs, and functions.
// See the following resources for a quick introduction:
// https://blog.golang.org/laws-of-reflection
// https://medium.com/capital-one-tech/learning-to-use-go-reflection-822a0aed74b7
func scanRows(rows *sql.Rows, dest interface{}) error {
	destVal := reflect.ValueOf(dest) // ValueOf is the entrypoint to reflection. This returns the underlying value of `dest`.
	destKind := destVal.Kind()       // The kind of the type of `dest`, such as Ptr, Struct, Int, etc.

	if destKind != reflect.Ptr {
		return errors.Newf("Dest must be a pointer, but was of kind: %v.", destKind)
	}

	destElem := destVal.Elem()      // The object that `destVal` points to, such as the struct or the slice of structs.
	destElemKind := destElem.Kind() // The kind of `destElem`, such as Struct, Slice, etc.

	switch destElemKind {
	case reflect.Struct:
		// Scan a single row to a struct
		if ok := rows.Next(); !ok {
			if rows.Err() != nil {
				log.Errorf("Error when calling rows.Next(): %s", rows.Err())
			}

			// No rows to scan
			return ErrNoRows
		}
		return scanRow(rows, dest)
	case reflect.Slice:
		// Scan all rows to a slice of structs
		destSliceType := destElem.Type()                 // The type of the slice `destElem`, such as []Foo{}
		destSliceElemKind := destSliceType.Elem().Kind() // The kind of the object that `destElem` stores, such as Struct, Slice, etc.
		if destSliceElemKind != reflect.Struct {
			return errors.Newf("Dest must be a pointer to a slice of structs, but was a pointer to a slice of: %v", destSliceElemKind)
		}
		return scanAllRows(rows, dest)
	default:
		return errors.Newf("Dest must be a pointer to a struct or a slice, but was a pointer to a: %v", destElemKind)
	}
}

// Helper function that scans a single `sql.Row` to `dest`.
// Precondition:
// - `dest` is a pointer to a struct.
// - `rows.Next()` returned true before this function was called.
func scanRow(rows *sql.Rows, dest interface{}) error {
	destVal := reflect.ValueOf(dest)    // Pointer to struct
	destStruct := destVal.Elem()        // The struct itself
	destStructType := destStruct.Type() // The type of the struct.

	cols, err := rows.Columns() // Get column names
	if err != nil {
		return err
	}

	fieldPtrs := make([]interface{}, 0, len(cols))
	for _, col := range cols {
		// Get index of field in struct `destElem` that corresponds to column `col`
		fieldIdx := getMatchingFieldIndex(destStructType, col)
		if fieldIdx < 0 {
			return errors.Newf("No matching struct field found for column: %s", col)
		}

		// Get pointer to field
		fieldPtr := destStruct.Field(fieldIdx).Addr().Interface()
		fieldPtrs = append(fieldPtrs, fieldPtr)
	}

	return rows.Scan(fieldPtrs...)
}

// Returns the field index of `column` in struct of type `typ` based on the
// field tag `db`. Returns -1 if there are no matches.
func getMatchingFieldIndex(typ reflect.Type, column string) int {
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Tag.Get(fieldTag) == column {
			return i
		}
	}
	return -1
}

// Helper function that scans all `sql.Rows` in `rows` to `dest`.
// Precondition:
// - `dest` is a pointer to a slice of structs.
func scanAllRows(rows *sql.Rows, dest interface{}) error {
	destVal := reflect.ValueOf(dest)          // Pointer to slice of structs
	destSlice := destVal.Elem()               // The slice of structs itself
	destSliceType := destSlice.Type()         // The type of the slice, such as []Foo{}
	destSliceElemType := destSliceType.Elem() // The type of the struct that the slice stores, such as Foo{}

	index := 0
	for rows.Next() {
		if destSlice.Len() == index {
			// Slice is full and must be expanded
			destSlice = expandSlice(destSlice, destSliceElemType)
		}

		// Get pointer to struct at `index` in `destSlice`
		destSliceElemPtr := destSlice.Index(index).Addr().Interface()
		err := scanRow(rows, destSliceElemPtr)
		if err != nil {
			return err
		}

		index++
	}

	// Point `dest` to new slice.
	// reflect.Indirect(v) returns the value that `v` points to, and `Set` updates that value.
	// destSlice.Slice(0, index) is used to ensure that the length of the slice `dest` points to is equal
	// to the number of rows that were scanned.
	reflect.Indirect(destVal).Set(destSlice.Slice(0, index))

	return nil
}

// Helper function to double the length (and capacity) of a slice of type `elemType`.
func expandSlice(slice reflect.Value, elemType reflect.Type) reflect.Value {
	newLen := (slice.Len() + 1) * 2
	newCap := (slice.Cap() + 1) * 2
	newSlice := reflect.MakeSlice(reflect.SliceOf(elemType), newLen, newCap)

	// Copy elements from old slice to new slice
	reflect.Copy(newSlice, slice)

	return newSlice
}
