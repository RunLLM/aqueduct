package engine

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/notification"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
)

func (eng *aqEngine) getNotifications(ctx context.Context, wfDag dag.WorkflowDag) ([]notification.Notification, error) {
	return notification.GetNotificationsFromUser(
		ctx,
		wfDag.UserID(),
		eng.IntegrationRepo,
		eng.Vault,
		eng.Database,
	)
}

func notificationMsg(wfDag dag.WorkflowDag, level shared.NotificationLevel, contextMsg string) string {
	statusMsg := ""
	contextSuffix := "."
	if len(contextMsg) > 0 {
		contextSuffix = fmt.Sprintf(": %s .", contextMsg)
	}
	if level == shared.SuccessNotificationLevel {
		statusMsg = fmt.Sprintf("succeeded%s", contextSuffix)
	} else if level == shared.WarningNotificationLevel {
		statusMsg = fmt.Sprintf("succeeded with warning%s", contextSuffix)
	} else if level == shared.ErrorNotificationLevel {
		statusMsg = fmt.Sprintf("failed%s", contextSuffix)
	} else {
		// For now, no caller will send message other than success, warning, or error.
		// This line is in case of future use cases.
		statusMsg = fmt.Sprintf("has a message: %s", contextMsg)
	}

	// Full message will look like "Workflow my_churn succeeded with warning: some context ."
	return fmt.Sprintf("Workflow %s %s", wfDag.Name(), statusMsg)
}

func (eng *aqEngine) sendNotifications(
	ctx context.Context,
	wfDag dag.WorkflowDag,
	level shared.NotificationLevel,
	contextMsg string,
) error {
	notifications, err := eng.getNotifications(ctx, wfDag)
	if err != nil {
		return err
	}

	msg := notificationMsg(wfDag, level, contextMsg)
	workflowSettings := wfDag.NotificationSettings().Settings
	for _, notificationObj := range notifications {
		if len(workflowSettings) > 0 {
			// send based on settings
			thresholdLevel, ok := workflowSettings[notificationObj.ID()]
			if ok {
				if notification.ShouldSend(thresholdLevel, level) {
					err = notificationObj.Send(ctx, msg)
					if err != nil {
						return err
					}
				}
			}
		} else {
			// Otherwise we send based on global settings.
			// ENG-2341 will allow user to configure if a notification applies to all workflows.
			if notification.ShouldSend(notificationObj.Level(), level) {
				err = notificationObj.Send(ctx, msg)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
