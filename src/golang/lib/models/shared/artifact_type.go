package shared

type ArtifactType string

const (
	Untyped   ArtifactType = "untyped"
	String    ArtifactType = "string"
	Bool      ArtifactType = "boolean"
	Numeric   ArtifactType = "numeric"
	Dict      ArtifactType = "dictionary"
	Tuple     ArtifactType = "tuple"
	Table     ArtifactType = "table"
	Json      ArtifactType = "json"
	Bytes     ArtifactType = "bytes"
	Image     ArtifactType = "image"
	Picklable ArtifactType = "picklable"
)

// `IsCompact` indicates if the value is 'small' enough to pass around in-memory.
// Otherwise, the value may not fit memory and should be passed around as storage pointers
// This is typically used for request / response handling.
// TODO (ENG-1687): persist compact values directly to DB.
func (t ArtifactType) IsCompact() bool {
	return t == Bool || t == Numeric || t == String
}
