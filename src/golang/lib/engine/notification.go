package engine

import (
	"github.com/aqueducthq/aqueduct/lib/notification"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
)

func (eng *aqEngine) getNotifications(ctx, wfDag dag.WorkflowDag) ([]notification.Notification, error) {
	return notification.GetNotificationsFromUser(
		ctx,
		wfDag.UserID(),
		eng.IntegrationRepo,
		eng.Vault,
		eng.Database,
	)
}
