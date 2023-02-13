package _000017_update_to_canceled_status

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
)

func Up(ctx context.Context, db database.Database) error {
	// First, get the list of all artifacts that are marked as pending and mark
	// them as canceled instead.
	pendingArtifactStatuses, err := getPendingArtifactResultStatuses(ctx, db)
	if err != nil {
		return err
	}

	for _, artifactStatusInfo := range pendingArtifactStatuses {
		artifactStatusInfo.ExecState.Status = CanceledExecutionStatus
		changes := map[string]interface{}{
			"status":          CanceledExecutionStatus,
			"execution_state": &artifactStatusInfo.ExecState,
		}

		err = repos.UpdateRecord(ctx, changes, "artifact_result", "id", artifactStatusInfo.ArtifactResultID, db)
		if err != nil {
			return err
		}
	}

	// Then do the same for all pending operators.
	pendingOperatorStatuses, err := getPendingOperatorResultStatuses(ctx, db)
	if err != nil {
		return err
	}

	for _, operatorStatusInfo := range pendingOperatorStatuses {
		operatorStatusInfo.ExecState.Status = CanceledExecutionStatus
		changes := map[string]interface{}{
			"status":          CanceledExecutionStatus,
			"execution_state": &operatorStatusInfo.ExecState,
		}

		err = repos.UpdateRecord(ctx, changes, "operator_result", "id", operatorStatusInfo.OperatorResultID, db)
		if err != nil {
			return err
		}
	}

	// Finally, update all failed artifacts to be canceled. In the new status
	// setting, failed operators lead to canceled artifacts because the artifacts
	// themselves are never created.
	failedArtifactStatuses, err := getFailedArtifactResultStatuses(ctx, db)
	if err != nil {
		return err
	}

	for _, artifactStatusInfo := range failedArtifactStatuses {
		artifactStatusInfo.ExecState.Status = CanceledExecutionStatus
		changes := map[string]interface{}{
			"status":          CanceledExecutionStatus,
			"execution_state": &artifactStatusInfo.ExecState,
		}

		err = repos.UpdateRecord(ctx, changes, "artifact_result", "id", artifactStatusInfo.ArtifactResultID, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func Down(ctx context.Context, db database.Database) error {
	// First, mark as previously canceled operators as pending.
	canceledOperatorStatuses, err := getCanceledOperatorResultStatuses(ctx, db)
	if err != nil {
		return err
	}

	for _, operatorStatusInfo := range canceledOperatorStatuses {
		operatorStatusInfo.ExecState.Status = PendingExecutionStatus
		changes := map[string]interface{}{
			"status":          PendingExecutionStatus,
			"execution_state": &operatorStatusInfo.ExecState,
		}

		err = repos.UpdateRecord(ctx, changes, "operator_result", "id", operatorStatusInfo.OperatorResultID, db)
		if err != nil {
			return err
		}
	}

	// Then do the same for all artifact statuses. However, there's a small hitch
	// here because prior to this update, some artifacts that are now canceled
	// were failed and some were pending. If the status has logs, had an error
	// set, it was failed, and otherwise it was pending.
	canceledArtifactStatuses, err := getCanceledArtifactResultStatuses(ctx, db)
	if err != nil {
		return err
	}

	for _, artifactStatusInfo := range canceledArtifactStatuses {
		var oldStatus ExecutionStatus
		if artifactStatusInfo.ExecState.Error != nil {
			oldStatus = FailedExecutionStatus
		} else {
			oldStatus = PendingExecutionStatus
		}

		artifactStatusInfo.ExecState.Status = oldStatus
		changes := map[string]interface{}{
			"status":          oldStatus,
			"execution_state": &artifactStatusInfo.ExecState,
		}

		err = repos.UpdateRecord(ctx, changes, "artifact_result", "id", artifactStatusInfo.ArtifactResultID, db)
		if err != nil {
			return err
		}
	}

	return nil
}
