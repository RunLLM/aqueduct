package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

type ArtifactResultMetadata struct {
	Schema []map[string]string // Table Schema from Pandas
	// Metrics from the system regarding the op used to create the artifact result.
	// A key/value pair of [metricname]metricvalue e.g. SystemMetric["runtime"] -> "3.65"
	SystemMetrics     map[string]string         `json:"system_metadata,omitempty"`
	SerializationType ArtifactSerializationType `json:"serialization_type,omitempty"`
	ArtifactType      ArtifactType              `json:"artifact_type,omitempty"`
	PythonType        string                    `json:"python_type,omitempty"`
}

type NullArtifactResultMetadata struct {
	ArtifactResultMetadata
	IsNull bool
}

func (m *ArtifactResultMetadata) Value() (driver.Value, error) {
	return utils.ValueJSONB(*m)
}

func (m *ArtifactResultMetadata) Scan(value interface{}) error {
	return utils.ScanJSONB(value, m)
}

func (n *NullArtifactResultMetadata) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.ArtifactResultMetadata).Value()
}

func (n *NullArtifactResultMetadata) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	metadata := &ArtifactResultMetadata{}
	if err := metadata.Scan(value); err != nil {
		return err
	}

	n.ArtifactResultMetadata, n.IsNull = *metadata, false
	return nil
}
