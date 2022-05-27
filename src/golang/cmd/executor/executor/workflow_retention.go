package executor

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type WorkflowRetentionExecutor struct {
	*BaseExecutor
}

func NewWorkflowRetentionExecutor(base *BaseExecutor) *WorkflowRetentionExecutor {
	return &WorkflowRetentionExecutor{BaseExecutor: base}
}

func (ex *WorkflowRetentionExecutor) Run(ctx context.Context) error {
	log.Info("Starting workflow retention.")
	txn, err := ex.Database.BeginTx(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to start transaction.")
	}
	defer txn.Rollback(ctx)

	// We first retrieve all relevant records from the database.
	workflowObjects, err := ex.WorkflowReader.GetAllWorkflows(ctx, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving workflow.")
	}

	for _, workflowObject := range workflowObjects {
		err = ex.cleanupOldWorkflows(ctx, txn, workflowObject.Id, workflowObject.RetentionPolicy.KLatestRuns)
		if err != nil {
			return err
		}

	}
	log.Info("Executed workflow retention.")

	return nil
}

func (ex *WorkflowRetentionExecutor) cleanupOldWorkflows(ctx context.Context, txn database.Transaction, workflowObjectId uuid.UUID, kLatestRuns int) error {
	// If kLatestRuns set to -1, we keep all runs.
	if kLatestRuns == -1 {
		return nil
	}

	workflowDagResults, err := ex.WorkflowDagResultReader.GetKOffsetWorkflowDagResultsByWorkflowId(ctx, workflowObjectId, kLatestRuns, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving workflow dags.")
	}
	workflowDagResultIds := make([]uuid.UUID, 0, len(workflowDagResults))
	for _, worklowDagResult := range workflowDagResults {
		workflowDagResultIds = append(workflowDagResultIds, worklowDagResult.Id)
	}

	if len(workflowDagResultIds) == 0 {
		return nil
	}

	operatorResultsToDelete, err := ex.OperatorResultReader.GetOperatorResultsByWorkflowDagResultIds(
		ctx,
		workflowDagResultIds,
		txn,
	)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving operator results.")
	}

	operatorResultIds := make([]uuid.UUID, 0, len(operatorResultsToDelete))
	for _, operatorResult := range operatorResultsToDelete {
		operatorResultIds = append(operatorResultIds, operatorResult.Id)
	}

	artifactResultsToDelete, err := ex.ArtifactResultReader.GetArtifactResultsByWorkflowDagResultIds(
		ctx,
		workflowDagResultIds,
		txn,
	)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving artifact results.")
	}

	artifactResultIds := make([]uuid.UUID, 0, len(artifactResultsToDelete))
	for _, artifactResult := range artifactResultsToDelete {
		artifactResultIds = append(artifactResultIds, artifactResult.Id)
	}

	// Do the deleting
	err = ex.OperatorResultWriter.DeleteOperatorResults(ctx, operatorResultIds, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting operator results.")
	}

	err = ex.ArtifactResultWriter.DeleteArtifactResults(ctx, artifactResultIds, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting artifact results.")
	}

	err = ex.WorkflowDagResultWriter.DeleteWorkflowDagResults(ctx, workflowDagResultIds, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting workflow dag results.")
	}

	if err := txn.Commit(ctx); err != nil {
		return errors.Wrap(err, "Failed to commit retention transaction.")
	}

	return nil
}
