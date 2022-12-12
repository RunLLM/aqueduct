package executor

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/operator/param"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const pollingIntervalMS = time.Millisecond * 500

type WorkflowExecutor struct {
	*BaseExecutor

	WorkflowId    uuid.UUID
	GithubManager github.Manager
	Engine        engine.Engine

	// The parameters to execute this workflow job with. If nil, then only default parameters
	// will be used. These values not persisted to the db.
	Parameters map[string]param.Param
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

	engineReaders := getEngineReaders(base.Readers)
	engineRepos := getEngineRepos(base.Repos)

	eng, err := engine.NewAqEngine(
		base.Database,
		githubManager,
		nil, /* PreviewCacheManager */
		base.Vault,
		spec.AqPath,
		engineReaders,
		engineRepos,
	)
	if err != nil {
		return nil, err
	}

	return &WorkflowExecutor{
		BaseExecutor:  base,
		WorkflowId:    workflowId,
		GithubManager: githubManager,
		Engine:        eng,
		Parameters:    spec.Parameters,
	}, nil
}

func (ex *WorkflowExecutor) Run(ctx context.Context) error {
	// First, ensure that workflow execution is not paused
	lock := utils.NewExecutionLock()
	// The following will block until the RLock can be acquired
	if err := lock.RLock(); err != nil {
		return err
	}
	defer func() {
		unlockErr := lock.RUnlock()
		if unlockErr != nil {
			log.Errorf("Unexpected error when unlocking execution lock: %v", unlockErr)
		}
	}()

	status, err := ex.Engine.ExecuteWorkflow(
		ctx,
		ex.WorkflowId,
		&engine.AqueductTimeConfig{
			OperatorPollInterval: pollingIntervalMS,
			ExecTimeout:          engine.DefaultExecutionTimeout,
			CleanupTimeout:       engine.DefaultCleanupTimeout,
		},
		ex.Parameters,
	)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"WorkflowId": ex.WorkflowId,
		"Parameters": ex.Parameters,
	}).Infof("Workflow run completed with status: %v", status)

	return nil
}
