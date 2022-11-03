package shared

type SerializationType string

const (
	String    SerializationType = "string"
	Table     SerializationType = "table"
	Json      SerializationType = "json"
	Bytes     SerializationType = "bytes"
	Image     SerializationType = "image"
	Picklable SerializationType = "picklable"
)
