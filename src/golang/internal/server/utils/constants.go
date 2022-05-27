package utils

const (
	// For logging purposes only
	Server = "Server"
)

const (
	// Strings used in HTTP REST headers.
	EmailHeader            = "email"
	Auth0IdHeader          = "user-id"
	UserEmailHeader        = "user-email"
	OrganizationIdHeader   = "organization-id"
	UserRoleHeader         = "user-role"
	ContentTypeHeader      = "content-type"
	ApiKeyHeader           = "api-key"
	LogsHeader             = "logs"
	SdkClientVersionHeader = "sdk-client-version"

	// Integration headers
	IntegrationNameHeader    = "integration-name"
	IntegrationServiceHeader = "integration-service"
	IntegrationConfigHeader  = "integration-config"

	TableNameHeader = "table-name"

	WorkflowIdUrlParam          = "workflowId"
	WorkflowDagResultIdUrlParam = "workflowDagResultId"
	OperatorIdUrlParam          = "operatorId"
	ArtifactIdUrlParam          = "artifactId"
	NotificationIdUrlParam      = "notificationId"
	IntegrationIdUrlParam       = "integrationId"

	CsvExportType   = "csv"
	ExcelExportType = "excel"

	ExportContentInput  = "input"
	ExportContentOutput = "output"

	// Allowed sdk client version
	// Only sdk client version equal and above will be accepted by server
	// ***** WARNING *****
	// the version number must be the same as what is in: sdk/aqueduct/_version.py
	// *******************
	AllowedSdkClientVersion = 3
)

type contextKeyType string

const (
	UserIdKey         contextKeyType = "userId"
	OrganizationIdKey contextKeyType = "organizationId"
	UserRequestIdKey  contextKeyType = "userRequestId"
	UserAuth0IdKey    contextKeyType = "userAuth0Id"
)
