package routes

// Please sort the route by their VALUEs
const (
	ListArtifactResultsRoute = "/api/artifact/{artifactId}/results"
	GetArtifactVersionsRoute = "/api/artifact_versions"
	GetArtifactResultRoute   = "/api/artifact_result/{workflowDagResultId}/{artifactId}"

	GetFunctionRoute    = "/api/function/{functionId}"
	ExportFunctionRoute = "/api/function/{operatorId}/export"

	ListIntegrationsRoute            = "/api/integrations"
	ConnectIntegrationRoute          = "/api/integration/connect"
	CreateTableRoute                 = "/api/integration/{integrationId}/create"
	DeleteIntegrationRoute           = "/api/integration/{integrationId}/delete"
	DiscoverRoute                    = "/api/integration/{integrationId}/discover"
	EditIntegrationRoute             = "/api/integration/{integrationId}/edit"
	ListIntegrationObjectsRoute      = "/api/integration/{integrationId}/objects"
	PreviewTableRoute                = "/api/integration/{integrationId}/preview_table"
	ListOperatorsForIntegrationRoute = "/api/integration/{integrationId}/operators"
	TestIntegrationRoute             = "/api/integration/{integrationId}/test"

	ResetApiKeyRoute = "/api/keys/reset" // nolint:gosec

	ListNotificationsRoute   = "/api/notifications"
	ArchiveNotificationRoute = "/api/notifications/{notificationId}/archive"

	GetOperatorResultRoute = "/api/operator_result/{workflowDagResultId}/{operatorId}"

	GetNodePositionsRoute = "/api/positioning"
	PreviewRoute          = "/api/preview"

	GetUserProfileRoute = "/api/user"

	ListWorkflowsRoute           = "/api/workflows"
	RegisterWorkflowRoute        = "/api/workflow/register"
	RegisterAirflowWorkflowRoute = "/api/workflow/register_airflow"
	GetWorkflowRoute             = "/api/workflow/{workflowId}"
	ListWorkflowObjectsRoute     = "/api/workflow/{workflowId}/objects"
	DeleteWorkflowRoute          = "/api/workflow/{workflowId}/delete"
	EditWorkflowRoute            = "/api/workflow/{workflowId}/edit"
	RefreshWorkflowRoute         = "/api/workflow/{workflowId}/refresh"
	GetWorkflowDagResultRoute    = "/api/workflow/{workflowId}/result/{workflowDagResultId}"
	UnwatchWorkflowRoute         = "/api/workflow/{workflowId}/unwatch"
	WatchWorkflowRoute           = "/api/workflow/{workflowId}/watch"
)
