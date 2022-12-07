package tests

import (
	"math/rand"
	"time"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	// Defaults used for seeding database records
	testOrgID = "aqueduct-test"
	testIntegrationService = shared.DemoDbIntegrationName
)

// seedIntegration creates count integration records.
func (ts *TestSuite) seedIntegration(count int) []models.Integration {
	integrations := make([]models.Integration, 0, count)
	users := ts.seedUser(1)

	for i := 0; i < count; i++ {
		name := randString(10)
		config := make(shared.IntegrationConfig)
		config[randString(10)] = randString(10)
		validated := true
		integration, err := ts.integration.CreateForUser(ts.ctx, testOrgID, users[0].ID, testIntegrationService, name, &config, validated, ts.DB)
		require.Nil(ts.T(), err)

		integrations = append(integrations, *integration)
	}

	return integrations
}

// seedNotification creates count notification records for a generated user.
func (ts *TestSuite) seedNotification(count int) []models.Notification {
	notifications := make([]models.Notification, 0, count)
	users := ts.seedUser(1)
	receiverID := users[0].ID

	for i := 0; i < count; i++ {
		content := randString(10)
		level := shared.SuccessNotificationLevel
		association := &shared.NotificationAssociation{
			Object: shared.OrgNotificationObject,
			ID: uuid.New(),
		}
		notification, err := ts.notification.Create(ts.ctx, receiverID, content, level, association, ts.DB)
		require.Nil(ts.T(), err)

		notifications = append(notifications, *notification)
	}

	return notifications
}
	
// seedArtifact creates count artifact records.
func (ts *TestSuite) seedArtifact(count int) []models.Artifact {
	artifacts := make([]models.Artifact, 0, count)

	for i := 0; i < count; i++ {
		name := randString(10)
		description := randString(15)
		artifactType := randArtifactType()
		artifact, err := ts.artifact.Create(ts.ctx, name, description, artifactType, ts.DB)
		require.Nil(ts.T(), err)

		artifacts = append(artifacts, *artifact)
	}

	return artifacts
}

// seedArtifactWithContext creates an artifact record in the context of a newly created workflow DAG.
func (ts *TestSuite) seedArtifactInWorkflow() (models.Artifact, models.DAG, models.Workflow, models.User) {
	artifacts := ts.seedArtifact(1)

	users := ts.seedUser(1)
	userIDs := sampleUserIDs(1, users)

	workflows := ts.seedWorkflowWithUser(1, userIDs)
	workflowIDs := sampleWorkflowIDs(1, workflows)

	dags := ts.seedDAGWithWorkflow(1, workflowIDs)
	dagID := dags[0].ID
	artifactID := artifacts[0].ID
	operatorID := uuid.New()

	_, _ = ts.dagEdge.Create(
		ts.ctx,
		dagID,
		shared.ArtifactToOperatorDAGEdge,
		artifactID,
		operatorID,
		0,
		ts.DB,
	)
	if rand.Intn(2) > 0 {
		_, _ = ts.dagEdge.Create(
			ts.ctx,
			dagID,
			shared.OperatorToArtifactDAGEdge,
			operatorID,
			artifactID,
			0,
			ts.DB,
		)
	}
	return artifacts[0], dags[0], workflows[0], users[0]
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

// seedDAGResult creates count DAGResult records.
// It also creates a new DAG to associate with the DAGResults created.
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

// seedDAGEdgeWith creates count DAGEdge records.
// It creates a new DAG to associate with the DAGEdges.
// For each DAGEdge, it randomly chooses the fromID, toID, and
// type of edge (e.g. Operator to Artifact).
func (ts *TestSuite) seedDAGEdge(count int) []models.DAGEdge {
	dags := ts.seedDAG(1)
	return ts.seedDAGEdgeWithDAG(count, dags[0].ID)
}

// seedDAGEdgeWith creates count DAGEdge records for the DAG specified.
// For each DAGEdge, it randomly chooses the fromID, toID, and
// type of edge (e.g. Operator to Artifact).
func (ts *TestSuite) seedDAGEdgeWithDAG(count int, dagID uuid.UUID) []models.DAGEdge {
	edges := make([]models.DAGEdge, 0, count)

	for i := 0; i < count; i++ {
		edgeType := shared.ArtifactToOperatorDAGEdge
		if rand.Intn(2) > 0 {
			edgeType = shared.OperatorToArtifactDAGEdge
		}

		edge, err := ts.dagEdge.Create(
			ts.ctx,
			dagID,
			edgeType,
			uuid.New(),
			uuid.New(),
			int16(i),
			ts.DB,
		)
		require.Nil(ts.T(), err)

		edges = append(edges, *edge)
	}

	return edges
}

// seedWatcher creates a Watcher record. It creates a new Workflow
// and User to use for the Watcher.
func (ts *TestSuite) seedWatcher() *models.Watcher {
	workflows := ts.seedWorkflow(1)
	workflow := workflows[0]

	watcher, err := ts.watcher.Create(
		ts.ctx,
		workflow.ID,
		workflow.UserID,
		ts.DB,
	)
	require.Nil(ts.T(), err)

	return watcher
}

// seedArtifactResult creates a workflow with 1 DAG and count artifact_result records
// belonging to the same workflow DAG.
func (ts *TestSuite) seedArtifactResult(count int) ([]models.ArtifactResult, models.Artifact, models.DAG, models.Workflow) {
	artifactResults := make([]models.ArtifactResult, 0, count)

	artifact, dag, workflow, _ := ts.seedArtifactInWorkflow()

	for i := 0; i < count; i++ {
		contentPath := randString(10)
		artifactResult, err := ts.artifactResult.Create(ts.ctx, dag.ID, artifact.ID, contentPath, ts.DB)
		require.Nil(ts.T(), err)

		artifactResults = append(artifactResults, *artifactResult)
	}

	return artifactResults, artifact, dag, workflow
}