package server

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	shared_utils "github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	"github.com/gorhill/cronexpr"
	log "github.com/sirupsen/logrus"
)

func (s *AqServer) SyncCronJobs() error {
	ctx := context.Background()
	if err := s.backfillKilledJobs(ctx); err != nil {
		return err
	}

	return s.runMissedCronJobs(ctx)
}

func (s *AqServer) triggerMissedCronJobs(
	ctx context.Context,
	workflowId uuid.UUID,
	cronSchedule string,
	referenceTime time.Time,
) {
	nextTwoTriggerTimeStamp := cronexpr.MustParse(cronSchedule).NextN(time.Now(), 2)
	// We subtract the next two trigger timestamp to get the duration between consecutive cron jobs.
	duration := nextTwoTriggerTimeStamp[1].Unix() - nextTwoTriggerTimeStamp[0].Unix()
	// Subtracting the duration from the next trigger timestamp gives us the last expected trigger timestamp.
	lastExpectedTriggerTime := nextTwoTriggerTimeStamp[0].Unix() - duration
	if lastExpectedTriggerTime > referenceTime.Unix() {
		// This means that the workflow should have been triggered, but it wasn't.
		// So we manually trigger the workflow here.
		_, _, err := (&handler.RefreshWorkflowHandler{
			Database: s.Database,
			Engine:   s.AqEngine,
		}).Perform(
			ctx,
			&handler.RefreshWorkflowArgs{
				WorkflowId: workflowId,
			},
		)
		if err != nil {
			log.Errorf("Unable to trigger workflow: %v", err)
		}
	}
}

// backfillKilledJobs backfills all pending and running op/artf/DAG _results
// and mark them as canceled. For non-aqueduct jobs like Airflow, we sync these
// jobs from the remote servers.
func (s *AqServer) backfillKilledJobs(ctx context.Context) error {
	txn, err := s.Database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	if _, err := s.OperatorResultRepo.UpdateBatchStatusByStatus(
		ctx,
		shared.RunningExecutionStatus,
		shared.CanceledExecutionStatus,
		txn,
	); err != nil {
		return err
	}

	if _, err := s.OperatorResultRepo.UpdateBatchStatusByStatus(
		ctx,
		shared.PendingExecutionStatus,
		shared.CanceledExecutionStatus,
		txn,
	); err != nil {
		return err
	}

	if _, err := s.ArtifactResultRepo.UpdateBatchStatusByStatus(
		ctx,
		shared.RunningExecutionStatus,
		shared.CanceledExecutionStatus,
		txn,
	); err != nil {
		return err
	}

	if _, err := s.ArtifactResultRepo.UpdateBatchStatusByStatus(
		ctx,
		shared.PendingExecutionStatus,
		shared.CanceledExecutionStatus,
		txn,
	); err != nil {
		return err
	}

	if _, err := s.DAGResultRepo.UpdateBatchStatusByStatus(
		ctx,
		shared.RunningExecutionStatus,
		shared.CanceledExecutionStatus,
		txn,
	); err != nil {
		return err
	}

	if _, err := s.DAGResultRepo.UpdateBatchStatusByStatus(
		ctx,
		shared.PendingExecutionStatus,
		shared.CanceledExecutionStatus,
		txn,
	); err != nil {
		return err
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return err
	}

	// Backfill self-orchestrated workflows after the above steps and overwrite
	// within the txn.
	if err := engine.SyncSelfOrchestratedWorkflows(
		ctx,
		"", /* orgID */
		s.ArtifactRepo,
		s.ArtifactResultRepo,
		s.DAGRepo,
		s.DAGEdgeRepo,
		s.DAGResultRepo,
		s.OperatorRepo,
		s.OperatorResultRepo,
		s.WorkflowRepo,
		vaultObject,
		txn,
	); err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// runMissedCronJobs first gets the latest workflow run timestamp of all deployed workflows that are
// running on Aqueduct, on a schedule, and are not paused. For each workflow, it compares the latest workflow
// run's timestamp with the expected trigger timestamp calculated based on the cron schedule, and manually
// triggers the workflow if the cron triggering did not happen.
func (s *AqServer) runMissedCronJobs(ctx context.Context) error {
	wfLastRuns, err := s.WorkflowRepo.GetLastRunByEngine(
		ctx,
		shared.AqueductEngineType,
		s.Database,
	)
	if err != nil {
		return errors.Wrap(err, "Unable to get workflow last run data from database.")
	}

	workflowsRan := map[uuid.UUID]bool{}

	for _, wfLastRun := range wfLastRuns {
		if wfLastRun.Schedule.CronSchedule != "" && !wfLastRun.Schedule.Paused {
			s.triggerMissedCronJobs(
				ctx,
				wfLastRun.ID,
				string(wfLastRun.Schedule.CronSchedule),
				wfLastRun.LastRunAt,
			)
		}
		workflowsRan[wfLastRun.ID] = true
	}

	allWorkflows, err := s.WorkflowRepo.List(ctx, s.Database)
	if err != nil {
		return errors.Wrap(err, "Unable to get workflows from database.")
	}

	for _, workflow := range allWorkflows {
		if _, ok := workflowsRan[workflow.ID]; !ok {
			// If we reach here, it means this workflow hasn't produced any run yet.
			if workflow.Schedule.CronSchedule != "" && !workflow.Schedule.Paused {
				s.triggerMissedCronJobs(
					ctx,
					workflow.ID,
					string(workflow.Schedule.CronSchedule),
					workflow.CreatedAt,
				)
			}
		}
	}

	return nil
}

func (s *AqServer) initializeWorkflowCronJobs(ctx context.Context) error {
	workflows, err := s.WorkflowRepo.List(ctx, s.Database)
	if err != nil {
		return err
	}

	for _, wf := range workflows {
		if wf.Schedule.CronSchedule != "" {
			if wf.Schedule.Paused {
				wf.Schedule.CronSchedule = ""
			}
			name := shared_utils.AppendPrefix(wf.ID.String())
			period := string(wf.Schedule.CronSchedule)

			err = s.AqEngine.ScheduleWorkflow(
				ctx,
				wf.ID,
				name,
				period,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
