package server

import "github.com/aqueducthq/aqueduct/internal/server/routes"

func (s *AqServer) Handlers() map[string]Handler {
	return map[string]Handler{
		routes.ArchiveNotificationRoute: &ArchiveNotificationHandler{
			NotificationReader: s.NotificationReader,
			NotificationWriter: s.NotificationWriter,
			Database:           s.Database,
		},
		routes.ConnectIntegrationRoute: &ConnectIntegrationHandler{
			Database:          s.Database,
			IntegrationWriter: s.IntegrationWriter,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
			StorageConfig:     s.StorageConfig,
		},
		routes.DeleteWorkflowRoute: &DeleteWorkflowHandler{
			Database:                s.Database,
			JobManager:              s.JobManager,
			WorkflowReader:          s.WorkflowReader,
			WorkflowDagReader:       s.WorkflowDagReader,
			WorkflowDagEdgeReader:   s.WorkflowDagEdgeReader,
			WorkflowDagResultReader: s.WorkflowDagResultReader,
			OperatorReader:          s.OperatorReader,
			OperatorResultReader:    s.OperatorResultReader,
			ArtifactResultReader:    s.ArtifactResultReader,

			WorkflowWriter:          s.WorkflowWriter,
			WorkflowDagWriter:       s.WorkflowDagWriter,
			WorkflowDagEdgeWriter:   s.WorkflowDagEdgeWriter,
			WorkflowDagResultWriter: s.WorkflowDagResultWriter,
			WorkflowWatcherWriter:   s.WorkflowWatcherWriter,
			OperatorWriter:          s.OperatorWriter,
			OperatorResultWriter:    s.OperatorResultWriter,
			ArtifactWriter:          s.ArtifactWriter,
			ArtifactResultWriter:    s.ArtifactResultWriter,
		},
		routes.EditWorkflowRoute: &EditWorkflowHandler{
			Database:       s.Database,
			WorkflowReader: s.WorkflowReader,
			WorkflowWriter: s.WorkflowWriter,
			JobManager:     s.JobManager,
			Vault:          s.Vault,
			GithubManager:  s.GithubManager,
		},
		routes.ExportFunctionRoute: &ExportFunctionHandler{
			Database:          s.Database,
			OperatorReader:    s.OperatorReader,
			WorkflowDagReader: s.WorkflowDagReader,
		},
		routes.GetArtifactResultRoute: &GetArtifactResultHandler{
			Database:             s.Database,
			ArtifactReader:       s.ArtifactReader,
			ArtifactResultReader: s.ArtifactResultReader,
			WorkflowDagReader:    s.WorkflowDagReader,
		},
		routes.GetArtifactVersionsRoute: &GetArtifactVersionsHandler{
			Database:     s.Database,
			CustomReader: s.CustomReader,
		},
		routes.GetNodePositionsRoute: &GetNodePositionsHandler{},
		routes.GetOperatorResultRoute: &GetOperatorResultHandler{
			Database:             s.Database,
			OperatorReader:       s.OperatorReader,
			OperatorResultReader: s.OperatorResultReader,
		},
		routes.GetUserProfileRoute: &GetUserProfileHandler{},
		routes.GetWorkflowRoute: &GetWorkflowHandler{
			Database:                s.Database,
			ArtifactReader:          s.ArtifactReader,
			OperatorReader:          s.OperatorReader,
			UserReader:              s.UserReader,
			WorkflowReader:          s.WorkflowReader,
			WorkflowDagReader:       s.WorkflowDagReader,
			WorkflowDagEdgeReader:   s.WorkflowDagEdgeReader,
			WorkflowDagResultReader: s.WorkflowDagResultReader,
		},
		routes.ListBuiltinFunctionsRoute: &ListBuiltinFunctionsHandler{
			StorageConfig: s.StorageConfig,
		},
		routes.ListIntegrationsRoute: &ListIntegrationsHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
		},
		routes.ListNotificationsRoute: &ListNotificationsHandler{
			Database:           s.Database,
			NotificationReader: s.NotificationReader,
			WorkflowReader:     s.WorkflowReader,
		},
		routes.ListWorkflowsRoute: &ListWorkflowsHandler{
			Database:       s.Database,
			WorkflowReader: s.WorkflowReader,
		},
		routes.PreviewTableRoute: &PreviewTableHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			StorageConfig:     s.StorageConfig,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.PreviewRoute: &PreviewHandler{
			Database:          s.Database,
			IntegrationReader: s.IntegrationReader,
			StorageConfig:     s.StorageConfig,
			GithubManager:     s.GithubManager,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.DiscoverRoute: &DiscoverHandler{
			Database:          s.Database,
			CustomReader:      s.CustomReader,
			IntegrationReader: s.IntegrationReader,
			StorageConfig:     s.StorageConfig,
			JobManager:        s.JobManager,
			Vault:             s.Vault,
		},
		routes.RefreshWorkflowRoute: &RefreshWorkflowHandler{
			Database:       s.Database,
			JobManager:     s.JobManager,
			GithubManager:  s.GithubManager,
			Vault:          s.Vault,
			WorkflowReader: s.WorkflowReader,
		},
		routes.RegisterWorkflowRoute: &RegisterWorkflowHandler{
			Database:      s.Database,
			JobManager:    s.JobManager,
			GithubManager: s.GithubManager,
			Vault:         s.Vault,
			StorageConfig: s.StorageConfig,

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
		routes.ResetApiKeyRoute: &ResetApiKeyHandler{
			Database:   s.Database,
			UserWriter: s.UserWriter,
		},
	}
}
