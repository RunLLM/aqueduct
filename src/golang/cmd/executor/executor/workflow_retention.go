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
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	// We first retrieve all relevant records from the database.
	workflows, err := ex.WorkflowRepo.List(ctx, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving workflow.")
	}

	for _, workflow := range workflows {
		err = ex.cleanupOldWorkflows(ctx, txn, workflow.ID, workflow.RetentionPolicy.KLatestRuns)
		if err != nil {
			return err
		}

	}
	log.Info("Executed workflow retention.")

	return nil
}

func (ex *WorkflowRetentionExecutor) cleanupOldWorkflows(
	ctx context.Context,
	txn database.Transaction,
	workflowObjectID uuid.UUID,
	kLatestRuns int,
) error {
	// If kLatestRuns set to -1, we keep all runs.
	if kLatestRuns == -1 {
		return nil
	}

	dagResults, err := ex.DAGResultRepo.GetKOffsetByWorkflow(
		ctx,
		workflowObjectID,
		kLatestRuns,
		txn,
	)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving workflow dags.")
	}
	dagResultIDs := make([]uuid.UUID, 0, len(dagResults))
	for _, dagResult := range dagResults {
		dagResultIDs = append(dagResultIDs, dagResult.ID)
	}

	if len(dagResultIDs) == 0 {
		return nil
	}

	operatorResultsToDelete, err := ex.OperatorResultReader.GetOperatorResultsByWorkflowDagResultIds(
		ctx,
		dagResultIDs,
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
		dagResultIDs,
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

	err = ex.DAGResultRepo.DeleteBatch(ctx, dagResultIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting workflow dag results.")
	}

	if err := txn.Commit(ctx); err != nil {
		return errors.Wrap(err, "Failed to commit retention transaction.")
	}

	return nil
}
