package shared

type SerializationType string

const (
	StringSerialization    SerializationType = "string"
	TableSerialization     SerializationType = "table"
	JsonSerialization      SerializationType = "json"
	BytesSerialization     SerializationType = "bytes"
	ImageSerialization     SerializationType = "image"
	PicklableSerialization SerializationType = "picklable"
)
