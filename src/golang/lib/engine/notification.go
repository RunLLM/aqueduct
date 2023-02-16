package engine

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/notification"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
)

type notificationContentStruct struct {
	level shared.NotificationLevel
	// Additional system error that the user should be notified.
	// Execution related user errors should already captured by
	// dag.WorkflowDag object using `OperatorsWithError()` and
	// `OperatorsWithWarning()` interfaces.
	systemErrContext string
}

func getNotifications(
	ctx context.Context,
	wfDag dag.WorkflowDag,
	vaultObject vault.Vault,
	integrationRepo repos.Integration,
	DB database.Database,
) ([]notification.Notification, error) {
	return notification.GetNotificationsFromUser(
		ctx,
		wfDag.UserID(),
		integrationRepo,
		vaultObject,
		DB,
	)
}

func sendNotifications(
	ctx context.Context,
	wfDag dag.WorkflowDag,
	content *notificationContentStruct,
	vaultObject vault.Vault,
	integrationRepo repos.Integration,
	DB database.Database,
) error {
	if content == nil {
		return nil
	}

	notifications, err := getNotifications(ctx, wfDag, vaultObject, integrationRepo, DB)
	if err != nil {
		return err
	}

	workflowSettings := wfDag.NotificationSettings().Settings
	for _, notificationObj := range notifications {
		if len(workflowSettings) > 0 {
			// send based on workflow settings
			thresholdLevel, ok := workflowSettings[notificationObj.ID()]
			if ok {
				if notification.ShouldSend(thresholdLevel, content.level) {
					err = notificationObj.SendForDag(
						ctx,
						wfDag,
						content.level,
						content.systemErrContext,
					)
					if err != nil {
						return err
					}
				}
			}
		} else {
			// Otherwise we send based on global settings.
			if notificationObj.Enabled() && notification.ShouldSend(
				notificationObj.Level(), content.level,
			) {
				err = notificationObj.SendForDag(
					ctx,
					wfDag,
					content.level,
					content.systemErrContext,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
