package server

import (
	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
)

func (s *AqServer) Handlers() map[string]handler.Handler {
	return map[string]handler.Handler{
		routes.ArchiveNotificationRoute: &handler.ArchiveNotificationHandler{
			NotificationReader: s.NotificationReader,
			NotificationWriter: s.NotificationWriter,
			Database:           s.Database,
		},
		routes.ConnectIntegrationRoute: &handler.ConnectIntegrationHandler{
			Database:   s.Database,
			JobManager: s.JobManager,
			Vault:      s.Vault,

			ArtifactReader:       s.ArtifactReader,
			ArtifactResultReader: s.ArtifactResultReader,
			OperatorReader:       s.OperatorReader,
			IntegrationReader:    s.IntegrationReader,
			IntegrationWriter:    s.IntegrationWriter,

			DAGRepo: s.DAGRepo,

			PauseServer:   s.Pause,
			RestartServer: s.Restart,
		},
		routes.DeleteIntegrationRoute: &handler.DeleteIntegrationHandler{
			Database:                   s.Database,
			Vault:                      s.Vault,
			CustomReader:               s.CustomReader,
			OperatorReader:             s.OperatorReader,
			IntegrationReader:          s.IntegrationReader,
			IntegrationWriter:          s.IntegrationWriter,
			ExecutionEnvironmentReader: s.ExecutionEnvironmentReader,
			ExecutionEnvironmentWriter: s.ExecutionEnvironmentWriter,
		},
		routes.DeleteWorkflowRoute: &handler.DeleteWorkflowHandler{
			Database:   s.Database,
			Engine:     s.AqEngine,
			JobManager: s.JobManager,
			Vault:      s.Vault,

			OperatorReader:             s.OperatorReader,
			IntegrationReader:          s.IntegrationReader,
			ExecutionEnvironmentReader: s.ExecutionEnvironmentReader,
			ExecutionEnvironmentWriter: s.ExecutionEnvironmentWriter,

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
			Database:       s.Database,
			OperatorReader: s.OperatorReader,

			DAGRepo: s.DAGRepo,
		},
		routes.GetArtifactResultRoute: &handler.GetArtifactResultHandler{
			Database:             s.Database,
			ArtifactReader:       s.ArtifactReader,
			ArtifactResultReader: s.ArtifactResultReader,

			DAGRepo:       s.DAGRepo,
			DAGResultRepo: s.DAGResultRepo,
		},
		routes.GetArtifactVersionsRoute: &handler.GetArtifactVersionsHandler{
			Database:     s.Database,
			CustomReader: s.CustomReader,
		},
		routes.GetNodePositionsRoute: &handler.GetNodePositionsHandler{},
		routes.GetOperatorResultRoute: &handler.GetOperatorResultHandler{
			Database:             s.Database,
			OperatorReader:       s.OperatorReader,
			OperatorResultReader: s.OperatorResultReader,

			DAGResultRepo: s.DAGResultRepo,
		},
		routes.GetUserProfileRoute: &handler.GetUserProfileHandler{},
		routes.ListWorkflowObjectsRoute: &handler.ListWorkflowObjectsHandler{
			Database:       s.Database,
			OperatorReader: s.OperatorReader,

			WorkflowRepo: s.WorkflowRepo,
		},
		routes.GetWorkflowRoute: &handler.GetWorkflowHandler{
			Database: s.Database,
			Vault:    s.Vault,

			ArtifactReader:        s.ArtifactReader,
			OperatorReader:        s.OperatorReader,
			WorkflowDagEdgeReader: s.WorkflowDagEdgeReader,

			OperatorResultWriter: s.OperatorResultWriter,
			ArtifactResultWriter: s.ArtifactResultWriter,

			DAGRepo:       s.DAGRepo,
			DAGResultRepo: s.DAGResultRepo,
			WorkflowRepo:  s.WorkflowRepo,
		},
		routes.GetWorkflowDagResultRoute: &handler.GetWorkflowDagResultHandler{
			Database:              s.Database,
			ArtifactReader:        s.ArtifactReader,
			ArtifactResultReader:  s.ArtifactResultReader,
			OperatorReader:        s.OperatorReader,
			OperatorResultReader:  s.OperatorResultReader,
			WorkflowDagEdgeReader: s.WorkflowDagEdgeReader,

			DAGRepo:       s.DAGRepo,
			DAGResultRepo: s.DAGResultRepo,
			WorkflowRepo:  s.WorkflowRepo,
		},
		routes.ListArtifactResultsRoute: &handler.ListArtifactResultsHandler{
			Database:             s.Database,
			ArtifactReader:       s.ArtifactReader,
			ArtifactResultReader: s.ArtifactResultReader,

			DAGRepo: s.DAGRepo,
		},
		routes.ListIntegrationsRoute: &handler.ListIntegrationsHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
		},
		routes.ListNotificationsRoute: &handler.ListNotificationsHandler{
			Database:           s.Database,
			NotificationReader: s.NotificationReader,
			WorkflowReader:     s.WorkflowReader,
		},
		routes.ListOperatorsForIntegrationRoute: &handler.ListOperatorsForIntegrationHandler{
			Database:          s.Database,
			OperatorReader:    s.OperatorReader,
			CustomReader:      s.CustomReader,
			IntegrationReader: s.IntegrationReader,
		},
		routes.ListWorkflowsRoute: &handler.ListWorkflowsHandler{
			Database:              s.Database,
			Vault:                 s.Vault,
			ArtifactReader:        s.ArtifactReader,
			OperatorReader:        s.OperatorReader,
			WorkflowReader:        s.WorkflowReader,
			WorkflowDagEdgeReader: s.WorkflowDagEdgeReader,
			CustomReader:          s.CustomReader,
			ArtifactWriter:        s.ArtifactWriter,
			OperatorWriter:        s.OperatorWriter,
			WorkflowDagEdgeWriter: s.WorkflowDagEdgeWriter,
			OperatorResultWriter:  s.OperatorResultWriter,
			ArtifactResultWriter:  s.ArtifactResultWriter,
			NotificationWriter:    s.NotificationWriter,

			DAGRepo:       s.DAGRepo,
			DAGResultRepo: s.DAGResultRepo,
			WorkflowRepo:  s.WorkflowRepo,
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

			WorkflowDagEdgeReader: s.WorkflowDagEdgeReader,
			OperatorReader:        s.OperatorReader,
			ArtifactReader:        s.ArtifactReader,

			DAGRepo:      s.DAGRepo,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.RegisterWorkflowRoute: &handler.RegisterWorkflowHandler{
			Database:      s.Database,
			JobManager:    s.JobManager,
			GithubManager: s.GithubManager,
			Vault:         s.Vault,
			Engine:        s.AqEngine,

			ArtifactReader:             s.ArtifactReader,
			IntegrationReader:          s.IntegrationReader,
			OperatorReader:             s.OperatorReader,
			ExecutionEnvironmentReader: s.ExecutionEnvironmentReader,

			ArtifactWriter:             s.ArtifactWriter,
			OperatorWriter:             s.OperatorWriter,
			WorkflowDagEdgeWriter:      s.WorkflowDagEdgeWriter,
			ExecutionEnvironmentWriter: s.ExecutionEnvironmentWriter,

			DAGRepo:      s.DAGRepo,
			WatcherRepo:  s.WatcherRepo,
			WorkflowRepo: s.WorkflowRepo,
		},
		routes.RegisterAirflowWorkflowRoute: &handler.RegisterAirflowWorkflowHandler{
			RegisterWorkflowHandler: handler.RegisterWorkflowHandler{
				Database:      s.Database,
				JobManager:    s.JobManager,
				GithubManager: s.GithubManager,
				Vault:         s.Vault,

				ArtifactReader:    s.ArtifactReader,
				IntegrationReader: s.IntegrationReader,
				OperatorReader:    s.OperatorReader,

				ArtifactWriter:        s.ArtifactWriter,
				OperatorWriter:        s.OperatorWriter,
				WorkflowDagEdgeWriter: s.WorkflowDagEdgeWriter,

				DAGRepo:      s.DAGRepo,
				WatcherRepo:  s.WatcherRepo,
				WorkflowRepo: s.WorkflowRepo,
			},
			WorkflowDagEdgeReader: s.WorkflowDagEdgeReader,

			OperatorResultWriter: s.OperatorResultWriter,
			ArtifactResultWriter: s.ArtifactResultWriter,
			NotificationWriter:   s.NotificationWriter,

			DAGResultRepo: s.DAGResultRepo,
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
