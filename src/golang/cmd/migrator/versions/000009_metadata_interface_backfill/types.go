package _000009_metadata_interface_backfill

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

type OldMetadata []map[string]string

type Metadata struct {
	Schema        []map[string]string // Table Schema from Pandas
	SystemMetrics map[string]string   // Metadata from the system
}

type OldNullMetadata struct {
	OldMetadata
	IsNull bool
}

type NullMetadata struct {
	Metadata
	IsNull bool
}

func (m *Metadata) Value() (driver.Value, error) {
	return utils.ValueJSONB(*m)
}

func (m *Metadata) Scan(value interface{}) error {
	return utils.ScanJSONB(value, m)
}

func (m *OldMetadata) Value() (driver.Value, error) {
	return utils.ValueJSONB(*m)
}

func (m *OldMetadata) Scan(value interface{}) error {
	return utils.ScanJSONB(value, m)
}

func (n *OldNullMetadata) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.OldMetadata).Value()
}

func (n *OldNullMetadata) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	metadata := &OldMetadata{}
	if err := metadata.Scan(value); err != nil {
		return err
	}

	n.OldMetadata, n.IsNull = *metadata, false
	return nil
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
