package executor

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
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

	engineReaders := GetEngineReaders(base.Readers)
	engineWriters := GetEngineWriters(base.Writers)

	engine, err := engine.NewAqEngine(
		base.Database,
		githubManager,
		base.Vault,
		spec.AqPath,
		spec.StorageConfig,
		engineReaders,
		engineWriters,
	)
	if err != nil {
		return nil, err
	}

	return &WorkflowExecutor{
		BaseExecutor:  base,
		WorkflowId:    workflowId,
		GithubManager: githubManager,
		Engine:        engine,
		Parameters:    spec.Parameters,
	}, nil
}

func (ex *WorkflowExecutor) Run(ctx context.Context) error {
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
