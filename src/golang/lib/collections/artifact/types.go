package artifact

type Type string

const (
	Untyped       Type = "untyped"
	StringType    Type = "string"
	BoolType      Type = "bool"
	NumericType   Type = "numeric"
	DictType      Type = "dictionary"
	TupleType     Type = "tuple"
	TabularType   Type = "tabular"
	JsonType      Type = "json"
	BytesType     Type = "bytes"
	ImageType     Type = "image"
	PicklableType Type = "picklable"
)
