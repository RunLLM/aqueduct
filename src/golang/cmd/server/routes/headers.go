package routes

const (
	// Strings used in HTTP REST headers.
	ContentTypeHeader      = "content-type"
	ApiKeyHeader           = "api-key"
	SdkClientVersionHeader = "sdk-client-version"

	// BaseResource headers
	IntegrationNameHeader    = "integration-name"
	IntegrationServiceHeader = "integration-service"
	IntegrationConfigHeader  = "integration-config"
	IntegrationIDHeader      = "integration-id"
	IntegrationIDsHeader     = "integration-ids"
	ServiceHeader            = "service"
	ImageNameHeader          = "image-name"

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
