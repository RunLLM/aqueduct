package server

import (
	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
)

func (s *AqServer) Handlers() map[string]handler.Handler {
	return map[string]handler.Handler{
		routes.ArchiveNotificationRoute: &handler.ArchiveNotificationHandler{
			Database: s.Database,

			NotificationRepo: s.NotificationRepo,
		},
		routes.ConnectIntegrationRoute: &handler.ConnectIntegrationHandler{
			Database:   s.Database,
			JobManager: s.JobManager,
			Vault:      s.Vault,

			IntegrationReader: s.IntegrationReader,
			IntegrationWriter: s.IntegrationWriter,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
			OperatorRepo:       s.OperatorRepo,

			PauseServer:   s.Pause,
			RestartServer: s.Restart,
		},
		routes.DeleteIntegrationRoute: &handler.DeleteIntegrationHandler{
			Database: s.Database,
			Vault:    s.Vault,

			CustomReader:      s.CustomReader,
			IntegrationReader: s.IntegrationReader,
			OperatorReader:    s.OperatorReader,

			IntegrationWriter:          s.IntegrationWriter,
			ExecutionEnvironmentReader: s.ExecutionEnvironmentReader,
			ExecutionEnvironmentWriter: s.ExecutionEnvironmentWriter,

			OperatorRepo: s.OperatorRepo,
		},
		routes.DeleteWorkflowRoute: &handler.DeleteWorkflowHandler{
			Database:   s.Database,
			Engine:     s.AqEngine,
			JobManager: s.JobManager,
			Vault:      s.Vault,

			IntegrationReader:          s.IntegrationReader,
			ExecutionEnvironmentReader: s.ExecutionEnvironmentReader,
			ExecutionEnvironmentWriter: s.ExecutionEnvironmentWriter,

			OperatorRepo: s.OperatorRepo,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.EditIntegrationRoute: &handler.EditIntegrationHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			IntegrationWriter: s.IntegrationWriter,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.EditWorkflowRoute: &handler.EditWorkflowHandler{
			Database: s.Database,
			Engine:   s.AqEngine,

			WorkflowRepo: s.WorkflowRepo,
		},
		routes.ExportFunctionRoute: &handler.ExportFunctionHandler{
			Database: s.Database,

			DAGRepo:      s.DAGRepo,
			OperatorRepo: s.OperatorRepo,
		},
		routes.GetArtifactResultRoute: &handler.GetArtifactResultHandler{
			Database: s.Database,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
			DAGResultRepo:      s.DAGResultRepo,
		},
		routes.GetArtifactVersionsRoute: &handler.GetArtifactVersionsHandler{
			Database:     s.Database,
			CustomReader: s.CustomReader,
		},
		routes.GetNodePositionsRoute: &handler.GetNodePositionsHandler{},
		routes.GetOperatorResultRoute: &handler.GetOperatorResultHandler{
			Database: s.Database,

			DAGResultRepo:      s.DAGResultRepo,
			OperatorRepo:       s.OperatorRepo,
			OperatorResultRepo: s.OperatorResultRepo,
		},
		routes.GetUserProfileRoute: &handler.GetUserProfileHandler{},
		routes.ListWorkflowObjectsRoute: &handler.ListWorkflowObjectsHandler{
			Database: s.Database,

			OperatorRepo: s.OperatorRepo,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.GetWorkflowRoute: &handler.GetWorkflowHandler{
			Database: s.Database,
			Vault:    s.Vault,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
			DAGEdgeRepo:        s.DAGEdgeRepo,
			DAGResultRepo:      s.DAGResultRepo,
			OperatorRepo:       s.OperatorRepo,
			OperatorResultRepo: s.OperatorResultRepo,
			WorkflowRepo:       s.WorkflowRepo,
		},
		routes.GetWorkflowDagResultRoute: &handler.GetWorkflowDagResultHandler{
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
		routes.ListArtifactResultsRoute: &handler.ListArtifactResultsHandler{
			Database: s.Database,

			ArtifactRepo:       s.ArtifactRepo,
			ArtifactResultRepo: s.ArtifactResultRepo,
			DAGRepo:            s.DAGRepo,
		},
		routes.ListIntegrationsRoute: &handler.ListIntegrationsHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
		},
		routes.ListNotificationsRoute: &handler.ListNotificationsHandler{
			Database: s.Database,

			DAGResultRepo:    s.DAGResultRepo,
			NotificationRepo: s.NotificationRepo,
		},
		routes.ListOperatorsForIntegrationRoute: &handler.ListOperatorsForIntegrationHandler{
			Database:          s.Database,
			OperatorReader:    s.OperatorReader,
			CustomReader:      s.CustomReader,
			IntegrationReader: s.IntegrationReader,
		},
		routes.ListWorkflowsRoute: &handler.ListWorkflowsHandler{
			Database:     s.Database,
			Vault:        s.Vault,
			CustomReader: s.CustomReader,

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
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.PreviewRoute: &handler.PreviewHandler{
			Database:                   s.Database,
			IntegrationReader:          s.IntegrationReader,
			ExecutionEnvironmentReader: s.ExecutionEnvironmentReader,
			ExecutionEnvironmentWriter: s.ExecutionEnvironmentWriter,
			GithubManager:              s.GithubManager,
			AqEngine:                   s.AqEngine,
		},
		routes.DiscoverRoute: &handler.DiscoverHandler{
			Database:          s.Database,
			CustomReader:      s.CustomReader,
			IntegrationReader: s.IntegrationReader,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.ListIntegrationObjectsRoute: &handler.ListIntegrationObjectsHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.CreateTableRoute: &handler.CreateTableHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.RefreshWorkflowRoute: &handler.RefreshWorkflowHandler{
			Database: s.Database,
			Engine:   s.AqEngine,
			Vault:    s.Vault,

			ArtifactRepo: s.ArtifactRepo,
			DAGRepo:      s.DAGRepo,
			DAGEdgeRepo:  s.DAGEdgeRepo,
			OperatorRepo: s.OperatorRepo,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.RegisterWorkflowRoute: &handler.RegisterWorkflowHandler{
			Database:      s.Database,
			JobManager:    s.JobManager,
			GithubManager: s.GithubManager,
			Vault:         s.Vault,
			Engine:        s.AqEngine,

			IntegrationReader:          s.IntegrationReader,
			ExecutionEnvironmentReader: s.ExecutionEnvironmentReader,

			ExecutionEnvironmentWriter: s.ExecutionEnvironmentWriter,

			ArtifactRepo: s.ArtifactRepo,
			DAGRepo:      s.DAGRepo,
			DAGEdgeRepo:  s.DAGEdgeRepo,
			OperatorRepo: s.OperatorRepo,
			WatcherRepo:  s.WatcherRepo,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.RegisterAirflowWorkflowRoute: &handler.RegisterAirflowWorkflowHandler{
			RegisterWorkflowHandler: handler.RegisterWorkflowHandler{
				Database:      s.Database,
				JobManager:    s.JobManager,
				GithubManager: s.GithubManager,
				Vault:         s.Vault,

				IntegrationReader: s.IntegrationReader,

				ArtifactRepo: s.ArtifactRepo,
				DAGRepo:      s.DAGRepo,
				DAGEdgeRepo:  s.DAGEdgeRepo,
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
		routes.TestIntegrationRoute: &handler.TestIntegrationHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.GetServerVersionRoute:     &handler.GetServerVersionHandler{},
		routes.GetServerEnvironmentRoute: &handler.GetServerEnvironmentHandler{},
	}
}
