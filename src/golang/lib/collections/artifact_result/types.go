package artifact_result

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
)

type SerializationType string

const (
	String    SerializationType = "string"
	Table     SerializationType = "table"
	Json      SerializationType = "json"
	Bytes     SerializationType = "bytes"
	Image     SerializationType = "image"
	Picklable SerializationType = "picklable"
)

type Metadata struct {
	Schema []map[string]string // Table Schema from Pandas
	// Metrics from the system regarding the op used to create the artifact result.
	// A key/value pair of [metricname]metricvalue e.g. SystemMetric["runtime"] -> "3.65"
	SystemMetrics     map[string]string `json:"system_metadata,omitempty"`
	SerializationType SerializationType `json:"serialization_type,omitempty"`
	// TODO: Change this to shared.ArtifactType when ArtifactResult refactor occurs
	ArtifactType artifact.Type `json:"artifact_type,omitempty"`
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
