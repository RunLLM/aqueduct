package executor

import (
	"context"
	"github.com/dropbox/godropbox/errors"
	"time"

	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/aqueducthq/aqueduct/lib/workflow/orchestrator"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const pollingIntervalMS = time.Millisecond * 500

type WorkflowExecutor struct {
	*BaseExecutor

	WorkflowId    uuid.UUID
	GithubManager github.Manager

	// The parameters to execute this workflow job with. If nil, then only default parameters
	// will be used. These values not persisted to the db.
	Parameters map[string]string
}

func NewWorkflowExecutor(spec *job.WorkflowSpec, base *BaseExecutor) (*WorkflowExecutor, error) {
	workflowId, err := uuid.Parse(spec.WorkflowId)
	if err != nil {
		return nil, err
	}

	githubManager, err := github.NewManager(spec.GithubManager)
	if err != nil {
		return nil, err
	}

	return &WorkflowExecutor{
		BaseExecutor:  base,
		WorkflowId:    workflowId,
		GithubManager: githubManager,
		Parameters:    spec.Parameters,
	}, nil
}

func (ex *WorkflowExecutor) Run(ctx context.Context) error {
	workflowDag, err := utils.ReadLatestWorkflowDagFromDatabase(
		ctx,
		ex.WorkflowId,
		ex.WorkflowReader,
		ex.WorkflowDagReader,
		ex.OperatorReader,
		ex.ArtifactReader,
		ex.WorkflowDagEdgeReader,
		ex.Database,
	)
	if err != nil {
		return err
	}

	githubClient, err := ex.GithubManager.GetClient(ctx, workflowDag.Metadata.UserId)
	if err != nil {
		return err
	}

	workflowDag, err = utils.UpdateWorkflowDagToLatest(
		ctx,
		githubClient,
		workflowDag,
		ex.WorkflowReader,
		ex.WorkflowWriter,
		ex.WorkflowDagReader,
		ex.WorkflowDagWriter,
		ex.OperatorReader,
		ex.OperatorWriter,
		ex.WorkflowDagEdgeReader,
		ex.WorkflowDagEdgeWriter,
		ex.ArtifactReader,
		ex.ArtifactWriter,
		ex.Database,
	)
	if err != nil {
		return err
	}

	// Overwrite the "default" values in the operator spec for this workflowDag.
	// Because this workflowDag is never written to the database, we will not contaminate
	// the default in the db.
	if ex.Parameters != nil {
		for name, newVal := range ex.Parameters {
			op := workflowDag.GetOperatorByName(name)
			if op == nil {
				continue
			}

			if !op.Spec.IsParam() {
				return errors.Newf("Cannot set parameters on a non-parameter operator %s", name)
			}
			workflowDag.Operators[op.Id].Spec.Param().Val = newVal
		}
	}

	workflowStoragePaths := utils.GenerateWorkflowStoragePaths(workflowDag)

	// Do not clean up artifact contents.
	defer utils.CleanupWorkflowStorageFiles(ctx, workflowStoragePaths, &workflowDag.StorageConfig, true /* metadataOnly */)

	status, err := orchestrator.Execute(
		ctx,
		workflowDag,
		workflowStoragePaths,
		pollingIntervalMS,
		ex.WorkflowReader,
		ex.WorkflowDagResultWriter,
		ex.OperatorResultWriter,
		ex.ArtifactResultWriter,
		ex.NotificationWriter,
		ex.UserReader,
		ex.Database,
		ex.JobManager,
		ex.Vault,
	)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"WorkflowId":    workflowDag.WorkflowId,
		"WorkflowDagId": workflowDag.Id,
		"Parameters":    ex.Parameters,
	}).Infof("Workflow run completed with status: %v", status)

	return nil
}
