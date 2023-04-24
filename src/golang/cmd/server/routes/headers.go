package routes

const (
	// Strings used in HTTP REST headers.
	ContentTypeHeader      = "content-type"
	ApiKeyHeader           = "api-key"
	SdkClientVersionHeader = "sdk-client-version"

	// Integration headers
	IntegrationNameHeader    = "integration-name"
	IntegrationServiceHeader = "integration-service"
	IntegrationConfigHeader  = "integration-config"
	IntegrationIDsHeader     = "integration-ids"

	// Storage Migration Headers
	StorageMigrationFilterStatusHeader   = "status"
	StorageMigrationLimitHeader          = "limit"
	StorageMigrationCompletedSinceHeader = "completed-since"

	// Dag Result Get Header
	DagResultGetOrderByHeader = "order_by"
	DagResultGetLimitHeader   = "limit"

	// Export Function headers
	ExportFnUserFriendlyHeader = "user-friendly"

	TableNameHeader = "table-name"

	MetadataOnlyHeader = "metadata-only"

	RunNowHeader              = "run-now"
	DynamicEngineActionHeader = "action"
)
