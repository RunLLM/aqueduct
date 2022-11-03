package shared

type Type string

const (
	Untyped   Type = "untyped"
	String    Type = "string"
	Bool      Type = "boolean"
	Numeric   Type = "numeric"
	Dict      Type = "dictionary"
	Tuple     Type = "tuple"
	Table     Type = "table"
	Json      Type = "json"
	Bytes     Type = "bytes"
	Image     Type = "image"
	Picklable Type = "picklable"
)

// `IsCompact` indicates if the value is 'small' enough to pass around in-memory.
// Otherwise, the value may not fit memory and should be passed around as storage pointers
// This is typically used for request / response handling.
// TODO (ENG-1687): persist compact values directly to DB.
func (t Type) IsCompact() bool {
	return t == Bool || t == Numeric || t == String
}
