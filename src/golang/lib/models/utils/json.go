package utils

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/errors"
)

// ValueJSONB JSON encodes value to store into a SQL Database.
func ValueJSONB(value interface{}) (driver.Value, error) {
	return json.Marshal(value)
}

// ScanJSONB scans value from a SQL Database into dest.
func ScanJSONB(value interface{}, dest interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion to []byte failed")
	}

	return json.Unmarshal(data, dest)
}
