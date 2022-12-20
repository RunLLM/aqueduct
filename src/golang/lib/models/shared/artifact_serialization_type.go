package shared

type ArtifactSerializationType string

const (
	StringSerialization    ArtifactSerializationType = "string"
	TableSerialization     ArtifactSerializationType = "table"
	BsonTableSerialization ArtifactSerializationType = "bson_table"
	JsonSerialization      ArtifactSerializationType = "json"
	BytesSerialization     ArtifactSerializationType = "bytes"
	ImageSerialization     ArtifactSerializationType = "image"
	PicklableSerialization ArtifactSerializationType = "picklable"
)
