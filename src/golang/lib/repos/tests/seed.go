package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	// Defaults used for seeding database records
	testOrgID = "aqueduct-test"
)

// seedArtifactResult creates count artifact_result records.
func (ts *TestSuite) seedArtifactResult(count int) []models.ArtifactResult {
	artifact_results := make([]models.ArtifactResult, 0, count)

	for i := 0; i < count; i++ {
		dagResultID := uuid.New()
		artifactID := uuid.New()
		contentPath := randString(10)
		artifact_result, err := ts.artifact_result.Create(ts.ctx, dagResultID, artifactID, contentPath, ts.DB)
		require.Nil(ts.T(), err)

		artifact_results = append(artifact_results, *artifact_result)
	}

	return artifact_results
}

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
