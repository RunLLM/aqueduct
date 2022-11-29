package tests

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	// Defaults used for seeding database records
	testOrgID = "aqueduct-test"
)

// seedUser creates count user records.
func (ts *TestSuite) seedUser(count int) []models.User {
	users := make([]models.User, 0, count)

	for i := 0; i < count; i++ {
		user, err := ts.user.Create(ts.ctx, testOrgID, randAPIKey(), ts.DB)
		require.Nil(ts.T(), err)

		users = append(users, *user)
	}

	return users
}

// seedWorkflow creates count workflow records.
// It creates a new user as the workflows' owner.
func (ts *TestSuite) seedWorkflow(count int) []models.Workflow {
	users := ts.seedUser(1)
	userIDs := sampleUserIDs(count, users)
	return ts.seedWorkflowWithUser(count, userIDs)
}

// seedWorkflowWithUser creates count workflow records. It uses userIDs as the
// owner of each workflow.
func (ts *TestSuite) seedWorkflowWithUser(count int, userIDs []uuid.UUID) []models.Workflow {
	require.Len(ts.T(), userIDs, count)

	workflows := make([]models.Workflow, 0, count)

	for i := 0; i < count; i++ {
		userID := userIDs[i]
		name := randString(10)
		description := randString(15)
		schedule := &shared.Schedule{
			Trigger:              shared.PeriodicUpdateTrigger,
			CronSchedule:         "* * * * *",
			DisableManualTrigger: false,
			Paused:               false,
		}
		retentionPolicy := &shared.RetentionPolicy{
			KLatestRuns: 10,
		}

		workflow, err := ts.workflow.Create(
			ts.ctx,
			userID,
			name,
			description,
			schedule,
			retentionPolicy,
			ts.DB,
		)
		require.Nil(ts.T(), err)

		workflows = append(workflows, *workflow)
	}

	return workflows
}

// seedDAG creates count DAG records.
// It also creates a new Workflow to associate with the DAG.
func (ts *TestSuite) seedDAG(count int) []models.DAG {
	workflows := ts.seedWorkflow(1)
	workflowIDs := sampleWorkflowIDs(count, workflows)
	return ts.seedDAGWithWorkflow(count, workflowIDs)
}

// seedDAGWithWorkflow creates count DAG records. It uses workflowIDs as the Workflow
// associated with each DAG.
func (ts *TestSuite) seedDAGWithWorkflow(count int, workflowIDs []uuid.UUID) []models.DAG {
	require.Len(ts.T(), workflowIDs, count)

	dags := make([]models.DAG, 0, count)

	for i := 0; i < count; i++ {
		workflowID := workflowIDs[i]
		storageConfig := &shared.StorageConfig{
			Type: shared.S3StorageType,
			S3Config: &shared.S3Config{
				Region: "us-east-2",
				Bucket: "test",
			},
		}
		engineConfig := &shared.EngineConfig{
			Type:           shared.AqueductEngineType,
			AqueductConfig: &shared.AqueductConfig{},
		}

		dag, err := ts.dag.Create(
			ts.ctx,
			workflowID,
			storageConfig,
			engineConfig,
			ts.DB,
		)
		require.Nil(ts.T(), err)

		dags = append(dags, *dag)
	}

	return dags
}

func (ts *TestSuite) seedDAGResult(count int) []models.DAGResult {
	dags := ts.seedDAG(1)
	dagIDs := sampleDagIDs(count, dags)
	return ts.seedDAGResultWithDAG(count, dagIDs)
}

// seedDAGResultWithDAG creates count DAGResult records. It uses dagIDs as the
// DAG associated with each DAGResult.
func (ts *TestSuite) seedDAGResultWithDAG(count int, dagIDs []uuid.UUID) []models.DAGResult {
	require.Len(ts.T(), dagIDs, count)

	dagResults := make([]models.DAGResult, 0, count)

	for i := 0; i < count; i++ {
		now := time.Now()
		execState := &shared.ExecutionState{
			Status: shared.PendingExecutionStatus,
			Timestamps: &shared.ExecutionTimestamps{
				PendingAt: &now,
			},
		}

		dagResult, err := ts.dagResult.Create(
			ts.ctx,
			dagIDs[i],
			execState,
			ts.DB,
		)
		require.Nil(ts.T(), err)

		dagResults = append(dagResults, *dagResult)
	}

	return dagResults
}
