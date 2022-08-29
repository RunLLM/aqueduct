package artifact

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
	None      Type = "none"
)
