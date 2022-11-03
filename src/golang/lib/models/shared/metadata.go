package shared

type Metadata struct {
	Schema []map[string]string // Table Schema from Pandas
	// Metrics from the system regarding the op used to create the artifact result.
	// A key/value pair of [metricname]metricvalue e.g. SystemMetric["runtime"] -> "3.65"
	SystemMetrics     map[string]string `json:"system_metadata,omitempty"`
	SerializationType SerializationType `json:"serialization_type,omitempty"`
	ArtifactType      Type     `json:"artifact_type,omitempty"`
}