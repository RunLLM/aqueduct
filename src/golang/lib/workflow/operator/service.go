package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// GetOperatorsOnIntegraiton will return an empty list for notification resources.
func GetOperatorsOnResource(
	ctx context.Context,
	orgID string,
	resource *models.Resource,
	resourceRepo repos.Resource,
	operatorRepo repos.Operator,
	DB database.Database,
) ([]models.Operator, error) {
	if shared.IsNotificationResource(resource.Service) {
		return []models.Operator{}, nil
	}

	resourceID := resource.ID

	// If the requested resource is a cloud resource, substitute the cloud resource ID
	// with the ID of the dynamic k8s resource.
	if resource.Service == shared.AWS {
		k8sResource, err := resourceRepo.GetByNameAndUser(
			ctx,
			fmt.Sprintf("%s:%s", resource.Name, dynamic.K8sResourceNameSuffix),
			uuid.Nil,
			orgID,
			DB,
		)
		if err != nil {
			return nil, err
		}

		resourceID = k8sResource.ID
	}

	resourceObject, err := resourceRepo.Get(ctx, resourceID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve resource.")
	}

	if shared.IsDataResource(resourceObject.Service) {
		return operatorRepo.GetExtractAndLoadOPsByResource(ctx, resourceID, DB)
	}

	// If the resource is the native Aqueduct compute engine, we need a separate query.
	if resourceObject.Service == shared.Aqueduct {
		return operatorRepo.GetForAqueductEngine(ctx, DB)
	}

	if _, ok := shared.ServiceToEngineConfigField[resourceObject.Service]; ok {
		return operatorRepo.GetByEngineResourceID(ctx, resourceID, DB)
	}

	// Other eligible cases
	if resourceObject.Service == shared.Conda {
		return operatorRepo.GetByEngineType(ctx, shared.AqueductCondaEngineType, DB)
	}

	// This feature is not supported for the given service.
	return nil, nil
}

// GetWorkflowIDsUsingNotification returns the list of all workflow IDs using the given notification resource.
func GetWorkflowIDsUsingNotification(
	ctx context.Context,
	resourceObject *models.Resource,
	workflowRepo repos.Workflow,
	DB database.Database,
) ([]uuid.UUID, error) {
	// First we look at the globally set level on the notification resource.
	// A nil level means that the default notification setting is disabled.
	defaultLevel, err := lib_utils.ExtractNotificationLevel(resourceObject)
	if err != nil {
		return nil, err
	}

	// Second, we look at all the workflows that have a custom notification setting pointing
	// to this notification resource.
	workflowObjects, err := workflowRepo.List(ctx, DB)
	if err != nil {
		return nil, err
	}

	// This notification is considered disabled for a workflow if the custom settings dict is
	// non-empty, but does not reference this notification.
	// These two lists are disjoint.
	customNotificationWorkflowIDs := make([]uuid.UUID, 0, len(workflowObjects))
	disabledWorkflowIDs := make(map[uuid.UUID]bool, 1)
	for _, workflowObj := range workflowObjects {
		if workflowObj.NotificationSettings.Settings != nil &&
			len(workflowObj.NotificationSettings.Settings) > 0 {
			if _, ok := workflowObj.NotificationSettings.Settings[resourceObject.ID]; ok {
				customNotificationWorkflowIDs = append(customNotificationWorkflowIDs, workflowObj.ID)
			} else {
				disabledWorkflowIDs[workflowObj.ID] = true
			}
		}
	}

	// If the notification is disabled globally, then we only count the custom workflows using this notification.
	if defaultLevel == nil {
		return customNotificationWorkflowIDs, nil
	}

	// Otherwise, we count all the workflows, except for those explicitly disabled. Note that there is
	// currently no way of disabling all notifications for a given workflow until ENG-2944 is fixed.
	enabledForWorkflowIDs := make([]uuid.UUID, 0, len(workflowObjects))
	allWorkflowIDs := slices.Map(workflowObjects, func(w models.Workflow) uuid.UUID {
		return w.ID
	})
	for _, workflowID := range allWorkflowIDs {
		if _, ok := disabledWorkflowIDs[workflowID]; !ok {
			enabledForWorkflowIDs = append(enabledForWorkflowIDs, workflowID)
		}
	}
	return enabledForWorkflowIDs, nil
}
