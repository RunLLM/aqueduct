package routes

const (
	// Strings used in HTTP REST headers.
	ContentTypeHeader      = "content-type"
	ApiKeyHeader           = "api-key"
	SdkClientVersionHeader = "sdk-client-version"

	// Resource headers
	ResourceNameHeader    = "resource-name"
	ResourceServiceHeader = "resource-service"
	ResourceConfigHeader  = "resource-config"
	ResourceIDHeader      = "resource-id"
	ResourceIDsHeader     = "resource-ids"
	ServiceHeader         = "service"
	ImageNameHeader       = "image-name"

	// Storage Migration Headers
	StorageMigrationFilterStatusHeader   = "status"
	StorageMigrationLimitHeader          = "limit"
	StorageMigrationCompletedSinceHeader = "completed-since"

	// Export Function headers
	ExportFnUserFriendlyHeader = "user-friendly"

	TableNameHeader = "table-name"

	MetadataOnlyHeader = "metadata-only"

	RunNowHeader              = "run-now"
	DynamicEngineActionHeader = "action"
)
