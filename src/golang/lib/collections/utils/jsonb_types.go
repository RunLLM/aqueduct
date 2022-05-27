package utils

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// This file defines various types that should be stored as a JSONB in Postgres.
// Each type should override the following methods:
// - Value() (driver.Value, error)
// - Scan(value interface{}) error
// Value prepares the type to be inserted as a JSONB.
// Scan parses the JSONB into the type.

type Logs map[string]interface{}

func (l *Logs) Value() (driver.Value, error) {
	return ValueJsonB(*l)
}

func (l *Logs) Scan(value interface{}) error {
	return ScanJsonB(value, l)
}

type Config map[string]string

func (c *Config) Value() (driver.Value, error) {
	return ValueJsonB(*c)
}

func (c *Config) Scan(value interface{}) error {
	return ScanJsonB(value, c)
}

type ServicePorts map[uint32]uint32

func (s *ServicePorts) Value() (driver.Value, error) {
	return ValueJsonB(*s)
}

func (s *ServicePorts) Scan(value interface{}) error {
	return ScanJsonB(value, s)
}

type UUIDSlice []uuid.UUID

func (u *UUIDSlice) Value() (driver.Value, error) {
	return ValueJsonB(*u)
}

func (u *UUIDSlice) Scan(value interface{}) error {
	return ScanJsonB(value, u)
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
