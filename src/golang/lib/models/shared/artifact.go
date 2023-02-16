package shared

type ArtifactType string

const (
	UntypedArtifact   ArtifactType = "untyped"
	StringArtifact    ArtifactType = "string"
	BoolArtifact      ArtifactType = "boolean"
	NumericArtifact   ArtifactType = "numeric"
	DictArtifact      ArtifactType = "dictionary"
	TupleArtifact     ArtifactType = "tuple"
	TableArtifact     ArtifactType = "table"
	JsonArtifact      ArtifactType = "json"
	BytesArtifact     ArtifactType = "bytes"
	ImageArtifact     ArtifactType = "image"
	PicklableArtifact ArtifactType = "picklable"
)

// IsCompact indicates if the value is 'small' enough to pass around in-memory.
// Otherwise, the value may not fit memory and should be passed around as storage pointers
// This is typically used for request / response handling.
// TODO (ENG-1687): persist compact values directly to DB.
func (t ArtifactType) IsCompact() bool {
	return t == BoolArtifact || t == NumericArtifact || t == StringArtifact
}
