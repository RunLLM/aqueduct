package utils

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/dropbox/godropbox/errors"
)

// This file defines various types that should be stored as a JSONB in Postgres.
// Each type should override the following methods:
// - Value() (driver.Value, error)
// - Scan(value interface{}) error
// Value prepares the type to be inserted as a JSONB.
// Scan parses the JSONB into the type.

type Config map[string]string

func (c *Config) Value() (driver.Value, error) {
	return ValueJsonB(*c)
}

func (c *Config) Scan(value interface{}) error {
	return ScanJsonB(value, c)
}

// Helper function that JSON encodes any object that should be stored as
// a JSONB in Postgres.
func ValueJsonB(value interface{}) (driver.Value, error) {
	return json.Marshal(value)
}

// Helper function that parses a JSONB object into `dest`.
func ScanJsonB(value interface{}, dest interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion to []byte failed")
	}

	return json.Unmarshal(data, dest)
}
