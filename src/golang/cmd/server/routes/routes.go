package routes

// Please sort the routes by their VALUEs
const (
	// V2 routes
	ResourceOperatorsRoute         = "/api/v2/resource/{resourceID}/nodes/operators"
	ResourcesWorkflowsRoute        = "/api/v2/resources/workflows"
	ResourceWorkflowsRoute         = "/api/v2/resource/{resourceID}/workflows"
	ListStorageMigrationRoute      = "/api/v2/storage-migrations"
	WorkflowsRoute                 = "/api/v2/workflows"
	WorkflowRoute                  = "/api/v2/workflow/{workflowID}"
	WorkflowObjectsRoute           = "/api/v2/workflow/{workflowID}/objects"
	DAGsRoute                      = "/api/v2/workflow/{workflowID}/dags"
	DAGRoute                       = "/api/v2/workflow/{workflowID}/dag/{dagID}"
	DAGResultsRoute                = "/api/v2/workflow/{workflowID}/results"
	DAGResultRoute                 = "/api/v2/workflow/{workflowID}/result/{dagResultID}"
	NodesRoute                     = "/api/v2/workflow/{workflowID}/dag/{dagID}/nodes"
	NodeArtifactRoute              = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/artifact/{nodeID}"
	NodeArtifactResultContentRoute = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/artifact/{nodeID}/result/{nodeResultID}/content"
	NodeArtifactResultsRoute       = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/artifact/{nodeID}/results"
	NodeMetricRoute                = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/metric/{nodeID}"
	NodeMetricResultContentRoute   = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/metric/{nodeID}/result/{nodeResultID}/content"
	NodeCheckRoute                 = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/check/{nodeID}"
	NodeCheckResultContentRoute    = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/check/{nodeID}/result/{nodeResultID}/content"
	NodeOperatorRoute              = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/operator/{nodeID}"
	NodeDagOperatorsRoute          = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/operators"
	NodeOperatorContentRoute       = "/api/v2/workflow/{workflowID}/dag/{dagID}/node/operator/{nodeID}/content"
	NodesResultsRoute              = "/api/v2/workflow/{workflowID}/result/{dagResultID}/nodes/results"
	EnvironmentRoute               = "/api/v2/environment"

	// V2 hacky routes
	// These routes are supposed to be `v2/workflow/{workflowId}`
	// with PATCH (edit) / POST (trigger) / DELETE method.
	// However, it requires significant refactor of handler interfaces as we assumed
	// routes is unique per handler.
	// For now, we simply use the same handler for this route and v1 workflow edit route.
	WorkflowEditPostRoute    = "/api/v2/workflow/{workflowId}/edit"
	WorkflowTriggerPostRoute = "/api/v2/workflow/{workflowId}/trigger"
	WorkflowDeletePostRoute  = "/api/v2/workflow/{workflowId}/delete"

	// V1 routes
	GetArtifactVersionsRoute = "/api/artifact/versions"
	GetArtifactResultRoute   = "/api/artifact/{workflowDagResultId}/{artifactId}/result"

	GetConfigRoute        = "/api/config"
	ConfigureStorageRoute = "/api/config/storage/{resourceID}"

	ExportFunctionRoute = "/api/function/{operatorId}/export"

	ListResourcesRoute            = "/api/resources"
	ConnectResourceRoute          = "/api/resource/connect"
	CreateTableRoute              = "/api/resource/{resourceID}/create"
	DeleteResourceRoute           = "/api/resource/{resourceID}/delete"
	DiscoverRoute                 = "/api/resource/{resourceID}/discover"
	EditResourceRoute             = "/api/resource/{resourceID}/edit"
	ListResourceObjectsRoute      = "/api/resource/{resourceID}/objects"
	PreviewTableRoute             = "/api/resource/{resourceID}/preview"
	ListOperatorsForResourceRoute = "/api/resource/{resourceID}/operators"
	TestResourceRoute             = "/api/resource/{resourceID}/test"
	GetDynamicEngineStatusRoute   = "/api/resource/dynamic-engine/status"
	EditDynamicEngineRoute        = "/api/resource/dynamic-engine/{resourceID}/edit"
	GetImageURLRoute              = "/api/resource/container-registry/url"

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

	GetServerVersionRoute = "/api/version"
)
