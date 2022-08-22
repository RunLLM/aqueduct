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
			Database:          s.Database,
			IntegrationWriter: s.IntegrationWriter,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.DeleteWorkflowRoute: &handler.DeleteWorkflowHandler{
			Database:   s.Database,
			Engine:     s.AqEngine,
			JobManager: s.JobManager,
			Vault:      s.Vault,

			OperatorReader:    s.OperatorReader,
			IntegrationReader: s.IntegrationReader,
			WorkflowReader:    s.WorkflowReader,
		},
		routes.EditIntegrationRoute: &handler.EditIntegrationHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			IntegrationWriter: s.IntegrationWriter,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.EditWorkflowRoute: &handler.EditWorkflowHandler{
			Database:       s.Database,
			WorkflowReader: s.WorkflowReader,
			Engine:         s.AqEngine,
		},
		routes.ExportFunctionRoute: &handler.ExportFunctionHandler{
			Database:          s.Database,
			OperatorReader:    s.OperatorReader,
			WorkflowDagReader: s.WorkflowDagReader,
		},
		routes.GetArtifactResultRoute: &handler.GetArtifactResultHandler{
			Database:             s.Database,
			ArtifactReader:       s.ArtifactReader,
			ArtifactResultReader: s.ArtifactResultReader,
			WorkflowDagReader:    s.WorkflowDagReader,
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
		},
		routes.GetUserProfileRoute: &handler.GetUserProfileHandler{},
		routes.ListWorkflowObjectsRoute: &handler.ListWorkflowObjectsHandler{
			Database:       s.Database,
			OperatorReader: s.OperatorReader,
			WorkflowReader: s.WorkflowReader,
		},
		routes.GetWorkflowRoute: &handler.GetWorkflowHandler{
			Database:                s.Database,
			ArtifactReader:          s.ArtifactReader,
			OperatorReader:          s.OperatorReader,
			UserReader:              s.UserReader,
			WorkflowReader:          s.WorkflowReader,
			WorkflowDagReader:       s.WorkflowDagReader,
			WorkflowDagEdgeReader:   s.WorkflowDagEdgeReader,
			WorkflowDagResultReader: s.WorkflowDagResultReader,
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
			Database:       s.Database,
			OperatorReader: s.OperatorReader,
			CustomReader:   s.CustomReader,
		},
		routes.ListWorkflowsRoute: &handler.ListWorkflowsHandler{
			Database:                s.Database,
			Vault:                   s.Vault,
			UserReader:              s.UserReader,
			ArtifactReader:          s.ArtifactReader,
			OperatorReader:          s.OperatorReader,
			WorkflowReader:          s.WorkflowReader,
			WorkflowDagReader:       s.WorkflowDagReader,
			WorkflowDagEdgeReader:   s.WorkflowDagEdgeReader,
			WorkflowDagResultReader: s.WorkflowDagResultReader,
			CustomReader:            s.CustomReader,
			ArtifactWriter:          s.ArtifactWriter,
			OperatorWriter:          s.OperatorWriter,
			WorkflowWriter:          s.WorkflowWriter,
			WorkflowDagWriter:       s.WorkflowDagWriter,
			WorkflowDagEdgeWriter:   s.WorkflowDagEdgeWriter,
			WorkflowDagResultWriter: s.WorkflowDagResultWriter,
			OperatorResultWriter:    s.OperatorResultWriter,
			ArtifactResultWriter:    s.ArtifactResultWriter,
			NotificationWriter:      s.NotificationWriter,
		},
		routes.PreviewTableRoute: &handler.PreviewTableHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.PreviewRoute: &handler.PreviewHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			GithubManager:     s.GithubManager,
			AqEngine:          s.AqEngine,
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
			Database:       s.Database,
			WorkflowReader: s.WorkflowReader,
			Engine:         s.AqEngine,
		},
		routes.RegisterWorkflowRoute: &handler.RegisterWorkflowHandler{
			Database:      s.Database,
			JobManager:    s.JobManager,
			GithubManager: s.GithubManager,
			Vault:         s.Vault,
			Engine:        s.AqEngine,

			ArtifactReader:    s.ArtifactReader,
			IntegrationReader: s.IntegrationReader,
			OperatorReader:    s.OperatorReader,
			WorkflowReader:    s.WorkflowReader,

			ArtifactWriter:        s.ArtifactWriter,
			OperatorWriter:        s.OperatorWriter,
			WorkflowWriter:        s.WorkflowWriter,
			WorkflowDagWriter:     s.WorkflowDagWriter,
			WorkflowDagEdgeWriter: s.WorkflowDagEdgeWriter,
			WorkflowWatcherWriter: s.WorkflowWatcherWriter,
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
				WorkflowReader:    s.WorkflowReader,

				ArtifactWriter:        s.ArtifactWriter,
				OperatorWriter:        s.OperatorWriter,
				WorkflowWriter:        s.WorkflowWriter,
				WorkflowDagWriter:     s.WorkflowDagWriter,
				WorkflowDagEdgeWriter: s.WorkflowDagEdgeWriter,
				WorkflowWatcherWriter: s.WorkflowWatcherWriter,
			},
			WorkflowDagReader:     s.WorkflowDagReader,
			WorkflowDagEdgeReader: s.WorkflowDagEdgeReader,
		},
		routes.ResetApiKeyRoute: &handler.ResetApiKeyHandler{
			Database:   s.Database,
			UserWriter: s.UserWriter,
		},
		routes.TestIntegrationRoute: &handler.TestIntegrationHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
	}
}
