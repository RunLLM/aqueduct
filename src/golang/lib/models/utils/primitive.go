package utils

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
)

// NullString represents a string that may be NULL.
type NullString struct {
	String string
	IsNull bool
}

func (n *NullString) Scan(value interface{}) error {
	sqlNullString := &sql.NullString{}
	if err := sqlNullString.Scan(value); err != nil {
		return err
	}

	n.String, n.IsNull = sqlNullString.String, !sqlNullString.Valid
	return nil
}

// NullInt64 represents an int64 that may be NULL.
type NullInt64 struct {
	Int64  int64
	IsNull bool
}

func (n *NullInt64) Scan(value interface{}) error {
	sqlNullInt64 := &sql.NullInt64{}
	if err := sqlNullInt64.Scan(value); err != nil {
		return err
	}

	n.Int64, n.IsNull = sqlNullInt64.Int64, !sqlNullInt64.Valid
	return nil
}

// NullInt represents an int that may be NULL.
type NullInt struct {
	Int    int
	IsNull bool
}

func (n *NullInt) Scan(value interface{}) error {
	nullInt64 := &NullInt64{}
	if err := nullInt64.Scan(value); err != nil {
		return err
	}

	n.Int, n.IsNull = int(nullInt64.Int64), nullInt64.IsNull
	return nil
}

// NullFloat64 represents a float64 that may be NULL.
type NullFloat64 struct {
	Float64 float64
	IsNull  bool
}

func (n *NullFloat64) Scan(value interface{}) error {
	sqlNullFloat64 := &sql.NullFloat64{}
	if err := sqlNullFloat64.Scan(value); err != nil {
		return err
	}

	n.Float64, n.IsNull = sqlNullFloat64.Float64, !sqlNullFloat64.Valid
	return nil
}

// NullBool represents a bool that may be NULL.
type NullBool struct {
	Bool   bool
	IsNull bool
}

func (n *NullBool) Scan(value interface{}) error {
	sqlNullBool := &sql.NullBool{}
	if err := sqlNullBool.Scan(value); err != nil {
		return err
	}

	n.Bool, n.IsNull = sqlNullBool.Bool, !sqlNullBool.Valid
	return nil
}

// NullTime represents a time.Time that may be NULL.
type NullTime struct {
	Time   time.Time
	IsNull bool
}

func (n *NullTime) Scan(value interface{}) error {
	sqlNullTime := &sql.NullTime{}
	if err := sqlNullTime.Scan(value); err != nil {
		return err
	}

	n.Time, n.IsNull = sqlNullTime.Time, !sqlNullTime.Valid
	return nil
}

type UUIDSlice []uuid.UUID

func (u *UUIDSlice) Value() (driver.Value, error) {
	return ValueJSONB(*u)
}

func (u *UUIDSlice) Scan(value interface{}) error {
	return ScanJSONB(value, u)
}

// NullUUID represents a uuid.UUID that may be NULL.
type NullUUID struct {
	UUID   uuid.UUID
	IsNull bool
}

func (n *NullUUID) Scan(value interface{}) error {
	if value == nil {
		// UUID is NULL
		n.IsNull = true
		return nil
	}

	id := &uuid.UUID{}
	if err := id.Scan(value); err != nil {
		return err
	}

	n.UUID, n.IsNull = *id, false
	return nil
}

// NullUUIDSlice represents a UUIDSlice that may be NULL.
type NullUUIDSlice struct {
	UUIDSlice UUIDSlice
	IsNull    bool
}

func (n *NullUUIDSlice) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.UUIDSlice).Value()
}

func (n *NullUUIDSlice) Scan(value interface{}) error {
	if value == nil {
		// UUIDSlice is NULL
		n.IsNull = true
		return nil
	}

	uuidSlice := &UUIDSlice{}
	if err := uuidSlice.Scan(value); err != nil {
		return err
	}

	n.UUIDSlice, n.IsNull = *uuidSlice, false
	return nil
}

type Metadata struct {
	Schema []map[string]string // Table Schema from Pandas
	// Metrics from the system regarding the op used to create the artifact result.
	// A key/value pair of [metricname]metricvalue e.g. SystemMetric["runtime"] -> "3.65"
	SystemMetrics     map[string]string `json:"system_metadata,omitempty"`
	SerializationType SerializationType `json:"serialization_type,omitempty"`
	ArtifactType      artifact.Type     `json:"artifact_type,omitempty"`
}

type NullMetadata struct {
	Metadata
	IsNull bool
}

func (m *Metadata) Value() (driver.Value, error) {
	return utils.ValueJsonB(*m)
}

func (m *Metadata) Scan(value interface{}) error {
	return utils.ScanJsonB(value, m)
}

func (n *NullMetadata) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.Metadata).Value()
}

func (n *NullMetadata) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	metadata := &Metadata{}
	if err := metadata.Scan(value); err != nil {
		return err
	}

	n.Metadata, n.IsNull = *metadata, false
	return nil
}
