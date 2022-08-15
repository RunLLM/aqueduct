package artifact

type Type string

const (
	Untyped       Type = "untyped"
	StringType    Type = "string"
	BoolType      Type = "bool"
	NumericType   Type = "numeric"
	DictType      Type = "dictionary"
	TupleType     Type = "tuple"
	TableType     Type = "table"
	JsonType      Type = "json"
	BytesType     Type = "bytes"
	ImageType     Type = "image"
	PicklableType Type = "picklable"
	NoneType 	  Type = "none"
)
