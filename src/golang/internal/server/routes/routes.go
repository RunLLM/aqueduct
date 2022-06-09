package routes

// Please sort the route by their VALUEs
const (
	GetArtifactVersionsRoute = "/artifact_versions"
	GetArtifactResultRoute   = "/artifact_result/{workflowDagResultId}/{artifactId}"

	ListBuiltinFunctionsRoute = "/builtinFunctions"
	GetFunctionRoute          = "/function/{functionId}"
	ExportFunctionRoute       = "/function/{operatorId}/export"

	ListIntegrationsRoute   = "/integrations"
	ConnectIntegrationRoute = "/integration/connect"
	DeleteIntegrationRoute  = "/integration/{integrationId}/delete"
	CreateTableRoute        = "/integration/{integrationId}/create"
	PreviewTableRoute       = "/integration/{integrationId}/preview_table"
	DiscoverRoute           = "/integration/{integrationId}/discover"

	ResetApiKeyRoute = "/keys/reset"

	ListNotificationsRoute   = "/notifications"
	ArchiveNotificationRoute = "/notifications/{notificationId}/archive"

	GetOperatorResultRoute = "/operator_result/{workflowDagResultId}/{operatorId}"

	GetNodePositionsRoute = "/positioning"
	PreviewRoute          = "/preview"

	GetUserProfileRoute = "/user"

	ListWorkflowsRoute    = "/workflows"
	RegisterWorkflowRoute = "/workflow/register"
	GetWorkflowRoute      = "/workflow/{workflowId}"
	DeleteWorkflowRoute   = "/workflow/{workflowId}/delete"
	EditWorkflowRoute     = "/workflow/{workflowId}/edit"
	RefreshWorkflowRoute  = "/workflow/{workflowId}/refresh"
	UnwatchWorkflowRoute  = "/workflow/{workflowId}/unwatch"
	WatchWorkflowRoute    = "/workflow/{workflowId}/watch"
)
