package routes

// Please sort the routes by their VALUEs
const (
	// V2 routes
	ListStorageMigrationRoute = "/api/v2/storage-migrations"
	WorkflowsRoute            = "/api/v2/workflows"

	WorkflowRoute                  = "/api/v2/workflow/{workflowID}"
	DAGRoute                       = "/api/v2/workflow/{workflowID}/dag/{dagID}"
	DAGResultsRoute                = "/api/v2/workflow/{workflowID}/results"
	DAGResultRoute                 = "/api/v2/workflow/{workflowID}/result/{dagResultID}"
	NodesRoute                     = "/api/v2/workflow/{workflowID}/dag/{dagID}/nodes"
	NodeArtifactRoute              = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/artifact/{nodeID}"
	NodeArtifactResultContentRoute = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/artifact/{nodeID}/result/{nodeResultID}/content"
	NodeArtifactResultsRoute       = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/artifact/{nodeID}/results"
	NodeOperatorRoute              = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/operator/{nodeID}"
	NodeOperatorContentRoute       = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/operator/{nodeID}/content"
	NodesResultsRoute              = "/api/v2/workflow/{workflowID}/result/{dagResultID}/nodes/results"

	IntegrationsWorkflowsRoute = "/api/v2/integrations/workflows"

	// V1 routes
	GetArtifactVersionsRoute = "/api/artifact/versions"
	GetArtifactResultRoute   = "/api/artifact/{workflowDagResultId}/{artifactId}/result"

	GetConfigRoute        = "/api/config"
	ConfigureStorageRoute = "/api/config/storage/{integrationId}"

	ExportFunctionRoute = "/api/function/{operatorId}/export"

	ListIntegrationsRoute            = "/api/integrations"
	ConnectIntegrationRoute          = "/api/integration/connect"
	CreateTableRoute                 = "/api/integration/{integrationId}/create"
	DeleteIntegrationRoute           = "/api/integration/{integrationId}/delete"
	DiscoverRoute                    = "/api/integration/{integrationId}/discover"
	EditIntegrationRoute             = "/api/integration/{integrationId}/edit"
	ListIntegrationObjectsRoute      = "/api/integration/{integrationId}/objects"
	PreviewTableRoute                = "/api/integration/{integrationId}/preview"
	ListOperatorsForIntegrationRoute = "/api/integration/{integrationId}/operators"
	TestIntegrationRoute             = "/api/integration/{integrationId}/test"
	GetDynamicEngineStatusRoute      = "/api/integration/dynamic-engine/status"
	EditDynamicEngineRoute           = "/api/integration/dynamic-engine/{integrationId}/edit"

	ResetApiKeyRoute = "/api/keys/reset" // nolint:gosec

	ListNotificationsRoute   = "/api/notifications"
	ArchiveNotificationRoute = "/api/notifications/{notificationId}/archive"

	GetOperatorResultRoute = "/api/operator/{workflowDagResultId}/{operatorId}/result"

	GetNodePositionsRoute = "/api/positioning"
	PreviewRoute          = "/api/preview"

	GetUserProfileRoute = "/api/user"

	ListWorkflowsRoute           = "/api/workflows"
	RegisterWorkflowRoute        = "/api/workflow/register"
	RegisterAirflowWorkflowRoute = "/api/workflow/register/airflow"
	GetWorkflowRouteV1           = "/api/workflow/{workflowId}"
	ListArtifactResultsRoute     = "/api/workflow/{workflowId}/artifact/{artifactId}/results"
	GetWorkflowDAGRoute          = "/api/workflow/{workflowId}/dag/{workflowDagID}"
	ListWorkflowObjectsRoute     = "/api/workflow/{workflowId}/objects"
	DeleteWorkflowRoute          = "/api/workflow/{workflowId}/delete"
	EditWorkflowRoute            = "/api/workflow/{workflowId}/edit"
	RefreshWorkflowRoute         = "/api/workflow/{workflowId}/refresh"
	GetWorkflowDagResultRoute    = "/api/workflow/{workflowId}/result/{workflowDagResultId}"
	GetWorkflowHistoryRoute      = "/api/workflow/{workflowId}/history"

	GetServerVersionRoute     = "/api/version"
	GetServerEnvironmentRoute = "/api/environment"
)
