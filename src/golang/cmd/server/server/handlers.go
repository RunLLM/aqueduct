package server

import (
	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	v2 "github.com/aqueducthq/aqueduct/cmd/server/handler/v2"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
)

func (s *AqServer) Handlers() map[string]handler.Handler {
	return map[string]handler.Handler{
		// V2 Handlers
		routes.WorkflowRoute: &v2.WorkflowGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.DAGRoute: &v2.DAGGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
			DAGRepo:      s.DAGRepo,
		},
		routes.NodeDagOperatorsRoute: &v2.DagOperatorsGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.DAGsRoute: &v2.DAGsGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
			DAGRepo:      s.DAGRepo,
		},
		routes.DAGResultRoute: &v2.DAGResultGetHandler{
			Database:      s.Database,
			WorkflowRepo:  s.WorkflowRepo,
			DAGResultRepo: s.DAGResultRepo,
		},
		routes.DAGResultsRoute: &v2.DAGResultsGetHandler{
			Database:      s.Database,
			WorkflowRepo:  s.WorkflowRepo,
			DAGResultRepo: s.DAGResultRepo,
		},
		routes.ListStorageMigrationRoute: &v2.ListStorageMigrationsHandler{
			Database:             s.Database,
			StorageMigrationRepo: s.StorageMigrationRepo,
		},
		routes.WorkflowsRoute: &v2.WorkflowsGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.NodesRoute: &v2.NodesGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
			OperatorRepo: s.OperatorRepo,
			ArtifactRepo: s.ArtifactRepo,
		},
		routes.NodeArtifactRoute: &v2.NodeArtifactGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
			ArtifactRepo: s.ArtifactRepo,
		},
		routes.NodeArtifactResultContentRoute: &v2.NodeArtifactResultContentGetHandler{
			Database:           s.Database,
			WorkflowRepo:       s.WorkflowRepo,
			DAGRepo:            s.DAGRepo,
			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
		},
		routes.NodeArtifactResultsRoute: &v2.NodeArtifactResultsGetHandler{
			Database:           s.Database,
			WorkflowRepo:       s.WorkflowRepo,
			DAGRepo:            s.DAGRepo,
			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
		},
		routes.NodeMetricRoute: &v2.NodeMetricGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.NodeMetricResultContentRoute: &v2.NodeMetricResultContentGetHandler{
			Database:           s.Database,
			WorkflowRepo:       s.WorkflowRepo,
			DAGRepo:            s.DAGRepo,
			OperatorRepo:       s.OperatorRepo,
			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
		},
		routes.NodeCheckRoute: &v2.NodeCheckGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.NodeCheckResultContentRoute: &v2.NodeCheckResultContentGetHandler{
			Database:           s.Database,
			WorkflowRepo:       s.WorkflowRepo,
			DAGRepo:            s.DAGRepo,
			OperatorRepo:       s.OperatorRepo,
			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
		},
		routes.NodeOperatorContentRoute: &v2.NodeOperatorContentGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
			DAGRepo:      s.DAGRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.NodeOperatorRoute: &v2.NodeOperatorGetHandler{
			Database:     s.Database,
			WorkflowRepo: s.WorkflowRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.NodesResultsRoute: &v2.NodesResultsGetHandler{
			Database:           s.Database,
			WorkflowRepo:       s.WorkflowRepo,
			DAGRepo:            s.DAGRepo,
			ArtifactRepo:       s.ArtifactRepo,
			OperatorResultRepo: s.OperatorResultRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
		},
		routes.ResourceOperatorsRoute: &v2.ResourceOperatorsGetHandler{
			Database:     s.Database,
			ResourceRepo: s.ResourceRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.ResourcesWorkflowsRoute: &v2.ResourcesWorkflowsGetHandler{
			Database:      s.Database,
			ResourceRepo:  s.ResourceRepo,
			WorkflowRepo:  s.WorkflowRepo,
			DAGRepo:       s.DAGRepo,
			DAGResultRepo: s.DAGResultRepo,
			OperatorRepo:  s.OperatorRepo,
		},
		routes.ResourceWorkflowsRoute: &v2.ResourceWorkflowsGetHandler{
			Database:      s.Database,
			ResourceRepo:  s.ResourceRepo,
			WorkflowRepo:  s.WorkflowRepo,
			DAGRepo:       s.DAGRepo,
			DAGResultRepo: s.DAGResultRepo,
			OperatorRepo:  s.OperatorRepo,
		},
		routes.EnvironmentRoute: &v2.EnvironmentHandler{},

		// V1 Handlers
		// (ENG-2715) Remove deprecated ones
		routes.ArchiveNotificationRoute: &handler.ArchiveNotificationHandler{
			Database: s.Database,

			NotificationRepo: s.NotificationRepo,
		},
		routes.ConnectResourceRoute: &handler.ConnectResourceHandler{
			Database:   s.Database,
			JobManager: s.JobManager,

			ArtifactRepo:         s.ArtifactRepo,
			ArtifactResultRepo:   s.ArtifactResultRepo,
			DAGRepo:              s.DAGRepo,
			ResourceRepo:         s.ResourceRepo,
			StorageMigrationRepo: s.StorageMigrationRepo,
			OperatorRepo:         s.OperatorRepo,

			PauseServer:   s.Pause,
			RestartServer: s.Restart,
		},
		routes.DeleteResourceRoute: &handler.DeleteResourceHandler{
			Database: s.Database,

			DAGRepo:                  s.DAGRepo,
			ExecutionEnvironmentRepo: s.ExecutionEnvironmentRepo,
			ResourceRepo:             s.ResourceRepo,
			OperatorRepo:             s.OperatorRepo,
			StorageMigrationRepo:     s.StorageMigrationRepo,
			WorkflowRepo:             s.WorkflowRepo,
		},
		routes.DeleteWorkflowRoute: &v2.WorkflowDeleteHandler{
			Database:   s.Database,
			Engine:     s.AqEngine,
			JobManager: s.JobManager,

			ResourceRepo:             s.ResourceRepo,
			ExecutionEnvironmentRepo: s.ExecutionEnvironmentRepo,
			OperatorRepo:             s.OperatorRepo,
			WorkflowRepo:             s.WorkflowRepo,
			DagRepo:                  s.DAGRepo,
			ArtifactResultRepo:       s.ArtifactResultRepo,
		},
		routes.WorkflowDeletePostRoute: &v2.WorkflowDeleteHandler{
			Database:   s.Database,
			Engine:     s.AqEngine,
			JobManager: s.JobManager,

			ResourceRepo:             s.ResourceRepo,
			ExecutionEnvironmentRepo: s.ExecutionEnvironmentRepo,
			OperatorRepo:             s.OperatorRepo,
			WorkflowRepo:             s.WorkflowRepo,
			DagRepo:                  s.DAGRepo,
			ArtifactResultRepo:       s.ArtifactResultRepo,
		},
		routes.EditResourceRoute: &handler.EditResourceHandler{
			Database:   s.Database,
			JobManager: s.JobManager,

			ResourceRepo: s.ResourceRepo,
		},
		routes.EditWorkflowRoute: &v2.WorkflowPatchHandler{
			Database: s.Database,
			Engine:   s.AqEngine,

			ArtifactRepo: s.ArtifactRepo,
			DAGRepo:      s.DAGRepo,
			DAGEdgeRepo:  s.DAGEdgeRepo,
			OperatorRepo: s.OperatorRepo,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.WorkflowEditPostRoute: &v2.WorkflowPatchHandler{
			Database: s.Database,
			Engine:   s.AqEngine,

			ArtifactRepo: s.ArtifactRepo,
			DAGRepo:      s.DAGRepo,
			DAGEdgeRepo:  s.DAGEdgeRepo,
			OperatorRepo: s.OperatorRepo,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.ExportFunctionRoute: &handler.ExportFunctionHandlerDeprecated{
			Database: s.Database,

			DAGRepo:      s.DAGRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.GetArtifactResultRoute: &handler.GetArtifactResultHandlerDeprecated{
			Database: s.Database,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
			DAGResultRepo:      s.DAGResultRepo,
		},
		routes.GetArtifactVersionsRoute: &handler.GetArtifactVersionsHandler{
			Database: s.Database,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
			DAGResultRepo:      s.DAGResultRepo,
			OperatorRepo:       s.OperatorRepo,
			OperatorResultRepo: s.OperatorResultRepo,
		},
		routes.GetConfigRoute: &handler.GetConfigHandler{
			ResourceRepo:         s.ResourceRepo,
			StorageMigrationRepo: s.StorageMigrationRepo,
			Database:             s.Database,
		},
		routes.ConfigureStorageRoute: &handler.ConfigureStorageHandler{
			Database: s.Database,

			ArtifactRepo:         s.ArtifactRepo,
			ArtifactResultRepo:   s.ArtifactResultRepo,
			DAGRepo:              s.DAGRepo,
			ResourceRepo:         s.ResourceRepo,
			OperatorRepo:         s.OperatorRepo,
			StorageMigrationRepo: s.StorageMigrationRepo,

			PauseServerFn:   s.Pause,
			RestartServerFn: s.Restart,
		},
		routes.GetNodePositionsRoute: &handler.GetNodePositionsHandler{},
		routes.GetOperatorResultRoute: &handler.GetOperatorResultHandlerDeprecated{
			Database: s.Database,

			DAGResultRepo:      s.DAGResultRepo,
			OperatorRepo:       s.OperatorRepo,
			OperatorResultRepo: s.OperatorResultRepo,
		},
		routes.GetUserProfileRoute: &handler.GetUserProfileHandler{},
		routes.ListWorkflowObjectsRoute: &v2.WorkflowObjectsGetHandler{
			Database: s.Database,

			OperatorRepo:       s.OperatorRepo,
			WorkflowRepo:       s.WorkflowRepo,
			WorkflowDagRepo:    s.DAGRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
		},
		routes.WorkflowObjectsRoute: &v2.WorkflowObjectsGetHandler{
			Database: s.Database,

			OperatorRepo:       s.OperatorRepo,
			WorkflowRepo:       s.WorkflowRepo,
			WorkflowDagRepo:    s.DAGRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
		},
		routes.GetWorkflowRouteV1: &handler.GetWorkflowHandler{
			Database: s.Database,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
			DAGEdgeRepo:        s.DAGEdgeRepo,
			DAGResultRepo:      s.DAGResultRepo,
			OperatorRepo:       s.OperatorRepo,
			OperatorResultRepo: s.OperatorResultRepo,
			WorkflowRepo:       s.WorkflowRepo,
		},
		routes.GetWorkflowDAGRoute: &handler.GetWorkflowDAGHandler{
			Database: s.Database,

			ArtifactRepo: s.ArtifactRepo,
			DAGRepo:      s.DAGRepo,
			DAGEdgeRepo:  s.DAGEdgeRepo,
			OperatorRepo: s.OperatorRepo,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.GetWorkflowDagResultRoute: &handler.GetWorkflowDagResultHandlerDeprecated{
			Database: s.Database,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
			DAGEdgeRepo:        s.DAGEdgeRepo,
			DAGResultRepo:      s.DAGResultRepo,
			OperatorRepo:       s.OperatorRepo,
			OperatorResultRepo: s.OperatorResultRepo,
			WorkflowRepo:       s.WorkflowRepo,
		},
		routes.GetWorkflowHistoryRoute: &handler.GetWorkflowHistoryHandler{
			Database: s.Database,

			DAGResultRepo: s.DAGResultRepo,
			WorkflowRepo:  s.WorkflowRepo,
		},
		routes.ListArtifactResultsRoute: &handler.ListArtifactResultsHandlerDeprecated{
			Database: s.Database,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
		},
		routes.ListResourcesRoute: &handler.ListResourcesHandler{
			Database: s.Database,

			ResourceRepo: s.ResourceRepo,
		},
		routes.GetDynamicEngineStatusRoute: &handler.GetDynamicEngineStatusHandler{
			Database: s.Database,

			ResourceRepo: s.ResourceRepo,
		},
		routes.EditDynamicEngineRoute: &handler.EditDynamicEngineHandler{
			Database: s.Database,

			ResourceRepo: s.ResourceRepo,
		},
		routes.GetImageURLRoute: &handler.GetImageURLHandler{
			Database: s.Database,

			ResourceRepo: s.ResourceRepo,
		},
		routes.ListNotificationsRoute: &handler.ListNotificationsHandler{
			Database: s.Database,

			DAGResultRepo:    s.DAGResultRepo,
			NotificationRepo: s.NotificationRepo,
		},
		routes.ListOperatorsForResourceRoute: &handler.ListOperatorsResourecHandler{
			Database: s.Database,

			DAGRepo:      s.DAGRepo,
			ResourceRepo: s.ResourceRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.ListWorkflowsRoute: &handler.ListWorkflowsHandler{
			Database: s.Database,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
			DAGEdgeRepo:        s.DAGEdgeRepo,
			DAGResultRepo:      s.DAGResultRepo,
			OperatorRepo:       s.OperatorRepo,
			OperatorResultRepo: s.OperatorResultRepo,
			WorkflowRepo:       s.WorkflowRepo,
		},
		routes.PreviewTableRoute: &handler.PreviewTableHandler{
			Database:   s.Database,
			JobManager: s.JobManager,

			ResourceRepo: s.ResourceRepo,
		},
		routes.PreviewRoute: &handler.PreviewHandler{
			Database:      s.Database,
			GithubManager: s.GithubManager,
			AqEngine:      s.AqEngine,

			ExecutionEnvironmentRepo: s.ExecutionEnvironmentRepo,
			ResourceRepo:             s.ResourceRepo,
		},
		routes.DiscoverRoute: &handler.DiscoverHandler{
			Database:   s.Database,
			JobManager: s.JobManager,

			ResourceRepo: s.ResourceRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.ListResourceObjectsRoute: &handler.ListResourceObjectsHandler{
			Database:   s.Database,
			JobManager: s.JobManager,

			ResourceRepo: s.ResourceRepo,
		},
		routes.CreateTableRoute: &handler.CreateTableHandler{
			Database:   s.Database,
			JobManager: s.JobManager,

			ResourceRepo: s.ResourceRepo,
		},
		routes.RefreshWorkflowRoute: &v2.WorkflowPostHandler{
			Database: s.Database,
			Engine:   s.AqEngine,

			WorkflowRepo: s.WorkflowRepo,
		},
		routes.WorkflowTriggerPostRoute: &v2.WorkflowPostHandler{
			Database: s.Database,
			Engine:   s.AqEngine,

			WorkflowRepo: s.WorkflowRepo,
		},
		routes.RegisterWorkflowRoute: &handler.RegisterWorkflowHandler{
			Database:      s.Database,
			JobManager:    s.JobManager,
			GithubManager: s.GithubManager,
			Engine:        s.AqEngine,

			ArtifactRepo:             s.ArtifactRepo,
			DAGRepo:                  s.DAGRepo,
			DAGEdgeRepo:              s.DAGEdgeRepo,
			ExecutionEnvironmentRepo: s.ExecutionEnvironmentRepo,
			ResourceRepo:             s.ResourceRepo,
			OperatorRepo:             s.OperatorRepo,
			WatcherRepo:              s.WatcherRepo,
			WorkflowRepo:             s.WorkflowRepo,
		},
		routes.RegisterAirflowWorkflowRoute: &handler.RegisterAirflowWorkflowHandler{
			RegisterWorkflowHandler: handler.RegisterWorkflowHandler{
				Database:      s.Database,
				JobManager:    s.JobManager,
				GithubManager: s.GithubManager,

				ArtifactRepo: s.ArtifactRepo,
				DAGRepo:      s.DAGRepo,
				DAGEdgeRepo:  s.DAGEdgeRepo,
				ResourceRepo: s.ResourceRepo,
				OperatorRepo: s.OperatorRepo,
				WatcherRepo:  s.WatcherRepo,
				WorkflowRepo: s.WorkflowRepo,
			},

			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGResultRepo:      s.DAGResultRepo,
			OperatorResultRepo: s.OperatorResultRepo,
		},
		routes.ResetApiKeyRoute: &handler.ResetApiKeyHandler{
			Database: s.Database,
			UserRepo: s.UserRepo,
		},
		routes.TestResourceRoute: &handler.TestResourceHandler{
			Database:   s.Database,
			JobManager: s.JobManager,

			ResourceRepo: s.ResourceRepo,
		},
		routes.GetServerVersionRoute: &handler.GetServerVersionHandler{},
	}
}
