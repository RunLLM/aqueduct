package tests

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/check"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/function"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/metric"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	// Defaults used for seeding database records
	testOrgID              = "aqueduct-test"
	testIntegrationService = shared.AqueductDemo
)

// seedStorageMigraton creates a 5 storage migration records, alternating between
// real destination IDs and nil ones. Also updates each entry to have a complete set of timestamps.
func (ts *TestSuite) seedStorageMigration() []models.StorageMigration {
	count := 5
	storageMigrations := make([]models.StorageMigration, count)
	for i := 0; i < count; i++ {
		var destIntegrationID *uuid.UUID
		var err error
		if i%2 == 0 {
			rawID, err := uuid.NewUUID()
			destIntegrationID = &rawID
			require.Nil(ts.T(), err)
		}
		entry, err := ts.storageMigration.Create(ts.ctx, destIntegrationID, ts.DB)
		require.Nil(ts.T(), err)

		now := time.Now()
		entry.ExecState.Timestamps.RunningAt = &now
		entry.ExecState.Timestamps.FinishedAt = &now

		// Mark the last entry as current.
		is_current := i == count-1
		entry, err = ts.storageMigration.Update(ts.ctx, entry.ID, map[string]interface{}{
			"execution_state": &entry.ExecState,
			"current":         is_current,
		},
			ts.DB,
		)
		require.Nil(ts.T(), err)

		// Insert into the `storageMigrations` list in reverse order, which
		// simulates the result of calling List().
		storageMigrations[count-1-i] = *entry
	}
	return storageMigrations
}

// seedIntegration creates count integration records for the given user.
func (ts *TestSuite) seedIntegrationWithUser(count int, userID uuid.UUID) []models.Integration {
	integrations := make([]models.Integration, 0, count)

	for i := 0; i < count; i++ {
		name := randString(10)
		config := make(shared.IntegrationConfig)
		config[randString(10)] = randString(10)
		validated := true
		integration, err := ts.integration.CreateForUser(
			ts.ctx,
			testOrgID,
			userID,
			testIntegrationService,
			name,
			&config,
			validated,
			ts.DB,
		)
		require.Nil(ts.T(), err)

		integrations = append(integrations, *integration)
	}

	return integrations
}

// seedIntegration creates count integration records and a new user that owns all of them.
func (ts *TestSuite) seedIntegration(count int) []models.Integration {
	users := ts.seedUser(1)
	return ts.seedIntegrationWithUser(count, users[0].ID)
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
			ID:     uuid.New(),
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
			&shared.NotificationSettings{},
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

// seedDAGWithUser creates count DAG records for the user.
// It also creates a new Workflow to associate with the DAG.
func (ts *TestSuite) seedDAGWithUser(count int, user models.User) []models.DAG {
	userIDs := sampleUserIDs(count, []models.User{user})
	workflows := ts.seedWorkflowWithUser(1, userIDs)
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

// seedOperator creates count Operator records.
// It does not create any other records and only creates Function Operators.
func (ts *TestSuite) seedOperator(count int) []models.Operator {
	operators := make([]models.Operator, 0, count)

	for i := 0; i < count; i++ {
		spec := operator.NewSpecFromFunction(
			function.Function{},
		)

		operator, err := ts.operator.Create(
			ts.ctx,
			randString(10),
			randString(15),
			spec,
			nil,
			ts.DB,
		)
		require.Nil(ts.T(), err)

		operators = append(operators, *operator)
	}

	return operators
}

// seedOperatorResultForDAGAndOperator creates count OperatorResult records.
// It does not create any other records and only creates OperatorResults for the specified DAG and operator.
func (ts *TestSuite) seedOperatorResultForDAGAndOperator(count int, dagResultID uuid.UUID, operatorID uuid.UUID) []models.OperatorResult {
	operatorResults := make([]models.OperatorResult, 0, count)

	for i := 0; i < count; i++ {
		now := time.Now()
		execState := &shared.ExecutionState{
			Status: shared.PendingExecutionStatus,
			Timestamps: &shared.ExecutionTimestamps{
				PendingAt: &now,
			},
		}
		operatorResult, err := ts.operatorResult.Create(
			ts.ctx,
			dagResultID,
			operatorID,
			execState,
			ts.DB,
		)
		require.Nil(ts.T(), err)

		operatorResults = append(operatorResults, *operatorResult)
	}

	return operatorResults
}

// seedOperatorResult creates count OperatorResult records.
// It creates DAGEdges, Operators, OperatorResults.
func (ts *TestSuite) seedOperatorResult(count int, opType operator.Type) ([]models.OperatorResult, *models.Operator, uuid.UUID) {
	artifactID := uuid.New()
	users := ts.seedUser(1)
	userIDs := sampleUserIDs(1, users)
	workflows := ts.seedWorkflowWithUser(1, userIDs)
	workflowIDs := sampleWorkflowIDs(1, workflows)
	dags := ts.seedDAGWithWorkflow(1, workflowIDs)
	dag := dags[0]
	operator := ts.seedOperatorAndDAG(artifactID, dag.ID, users[0].ID, opType)

	return ts.seedOperatorResultForDAGAndOperator(count, dag.ID, operator.ID), operator, artifactID
}

// seedOperatorWithDAG creates count Operator records of Type opType.
// The supported options are Function, Extract, and Load.
// It creates a DAGEdge for each Operator to associate it with the specified DAG.
// The DAGEdge type is randomly chosen (unless it is a check type) and does not connect to an actual Artifact.
func (ts *TestSuite) seedOperatorWithDAG(count int, dagID uuid.UUID, userID uuid.UUID, opType operator.Type) []models.Operator {
	operators := make([]models.Operator, 0, count)

	// A fake Artifact is used for all of the DAGEdges
	artifactID := uuid.New()

	for i := 0; i < count; i++ {
		operator := ts.seedOperatorAndDAG(artifactID, dagID, userID, opType)
		operators = append(operators, *operator)
	}

	return operators
}

// seedOperatorAndDAG creates an Operator records of Type opType.
// The supported options are Function, Extract, and Load.
// It creates a DAGEdge for the Operator to associate it with the specified DAG.
// The DAGEdge type is randomly chosen (unless it is a check type) and does not connect to an actual Artifact.
func (ts *TestSuite) seedOperatorAndDAG(artifactID uuid.UUID, dagID uuid.UUID, userID uuid.UUID, opType operator.Type) *models.Operator {
	var spec *operator.Spec
	switch opType {
	case operator.FunctionType:
		spec = operator.NewSpecFromFunction(
			function.Function{},
		)
	case operator.ExtractType:
		spec = operator.NewSpecFromExtract(
			connector.Extract{
				Service:       shared.Postgres,
				IntegrationId: uuid.New(),
				Parameters:    &connector.PostgresExtractParams{},
			},
		)
	case operator.LoadType:
		loadIntegrations := ts.seedIntegrationWithUser(1, userID)
		loadIntegration := loadIntegrations[0]
		spec = operator.NewSpecFromLoad(
			connector.Load{
				Service:       loadIntegration.Service,
				IntegrationId: loadIntegration.ID,
				Parameters: &connector.PostgresLoadParams{
					RelationalDBLoadParams: connector.RelationalDBLoadParams{
						Table:      randString(10),
						UpdateMode: "replace",
					},
				},
			},
		)
	case operator.CheckType:
		spec = operator.NewSpecFromCheck(
			check.Check{
				Level:    check.ErrorLevel,
				Function: function.Function{},
			},
		)
	default:
		ts.Fail("Seeding an Operator of type %v is not supported", opType)
	}

	newOperator, err := ts.operator.Create(
		ts.ctx,
		randString(10),
		randString(15),
		spec,
		nil,
		ts.DB,
	)
	require.Nil(ts.T(), err)

	edgeType := shared.ArtifactToOperatorDAGEdge
	fromID, toID := artifactID, newOperator.ID
	if rand.Intn(2) > 0 && opType != operator.CheckType {
		// Randomly change the direction of the DAGEdge
		edgeType = shared.OperatorToArtifactDAGEdge
		fromID, toID = toID, fromID
	}

	_, err = ts.dagEdge.Create(
		ts.ctx,
		dagID,
		edgeType,
		fromID,
		toID,
		int16(0),
		ts.DB,
	)
	require.Nil(ts.T(), err)

	return newOperator
}

// seedOperatorAndDAG creates an Operator records of Type opType.
// The supported options are Function, Extract, and Load.
// It creates a DAGEdge for the Operator to associate it with the specified DAG.
// The DAGEdge type is OperatorToArtifactDAGEdge and does not connect to an actual Artifact.
func (ts *TestSuite) seedOperatorAndDAGOperatorToArtifact(artifactID uuid.UUID, dagID uuid.UUID, opType operator.Type) *models.Operator {
	var spec *operator.Spec
	switch opType {
	case operator.FunctionType:
		spec = operator.NewSpecFromFunction(
			function.Function{},
		)
	case operator.ExtractType:
		spec = operator.NewSpecFromExtract(
			connector.Extract{
				Service:       shared.Postgres,
				IntegrationId: uuid.New(),
				Parameters:    &connector.PostgresExtractParams{},
			},
		)
	case operator.LoadType:
		spec = operator.NewSpecFromLoad(
			connector.Load{
				Service:       shared.Postgres,
				IntegrationId: uuid.New(),
				Parameters: &connector.PostgresLoadParams{
					RelationalDBLoadParams: connector.RelationalDBLoadParams{
						Table:      randString(10),
						UpdateMode: "replace",
					},
				},
			},
		)
	case operator.CheckType:
		spec = operator.NewSpecFromCheck(
			check.Check{
				Level:    check.ErrorLevel,
				Function: function.Function{},
			},
		)
	default:
		ts.Fail("Seeding an Operator of type %v is not supported", opType)
	}

	newOperator, err := ts.operator.Create(
		ts.ctx,
		randString(10),
		randString(15),
		spec,
		nil,
		ts.DB,
	)
	require.Nil(ts.T(), err)

	edgeType := shared.ArtifactToOperatorDAGEdge
	fromID, toID := artifactID, newOperator.ID
	edgeType = shared.OperatorToArtifactDAGEdge
	fromID, toID = newOperator.ID, artifactID

	_, err = ts.dagEdge.Create(
		ts.ctx,
		dagID,
		edgeType,
		fromID,
		toID,
		int16(0),
		ts.DB,
	)
	require.Nil(ts.T(), err)

	return newOperator
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

// seedSchemaVersion creates count schema versions versioned from CurrentSchemaVersion + 1  to CurrentSchemaVersion + count.
func (ts *TestSuite) seedSchemaVersion(count int) []models.SchemaVersion {
	schemaVersions := make([]models.SchemaVersion, 0, count)

	for i := 1; i <= count; i++ {
		name := randString(10)
		schemaVersion, err := ts.schemaVersion.Create(ts.ctx, int64(models.CurrentSchemaVersion+i), name, ts.DB)
		require.Nil(ts.T(), err)

		schemaVersions = append(schemaVersions, *schemaVersion)
	}

	return schemaVersions
}

// seedUnusedExecutionEnvironment creates `count` unused execution environments in execution_environment.
func (ts *TestSuite) seedUnusedExecutionEnvironment(count int) []models.ExecutionEnvironment {
	executionEnvironments := make([]models.ExecutionEnvironment, 0, count)

	for i := 0; i < count; i++ {
		spec := shared.ExecutionEnvironmentSpec{
			PythonVersion: randString(10),
			Dependencies:  []string{randString(10), randString(10), randString(10)},
		}
		hash := uuid.New()
		executionEnvironment, err := ts.executionEnvironment.Create(ts.ctx, &spec, hash, ts.DB)
		require.Nil(ts.T(), err)

		executionEnvironments = append(executionEnvironments, *executionEnvironment)
	}

	return executionEnvironments
}

// seedUsedExecutionEnvironment creates `count` used execution environments in execution_environment
// and the workflow and operators that use them.
func (ts *TestSuite) seedUsedExecutionEnvironment(count int) ([]models.ExecutionEnvironment, []models.Operator) {
	operators := make([]models.Operator, 0, count)

	users := ts.seedUser(1)
	userIDs := sampleUserIDs(1, users)

	workflows := ts.seedWorkflowWithUser(1, userIDs)
	workflowIDs := sampleWorkflowIDs(1, workflows)

	dags := ts.seedDAGWithWorkflow(1, workflowIDs)
	dagID := dags[0].ID

	executionEnvironments := ts.seedUnusedExecutionEnvironment(count)
	for i := 0; i < count; i++ {
		artifactID := uuid.New()

		spec := operator.NewSpecFromFunction(
			function.Function{},
		)

		operator, err := ts.operator.Create(
			ts.ctx,
			randString(10),
			randString(15),
			spec,
			&executionEnvironments[i].ID,
			ts.DB,
		)
		require.Nil(ts.T(), err)
		operators = append(operators, *operator)

		_, _ = ts.dagEdge.Create(
			ts.ctx,
			dagID,
			shared.OperatorToArtifactDAGEdge,
			operator.ID,
			artifactID,
			int16(i),
			ts.DB,
		)
	}

	return executionEnvironments, operators
}

// seedComplexWorkflow creates a workflow with multiple operator and artifacts.
// To make expected results easy to find, the map of artifacts and operators are keyed by names.
// The workflow reflects the following DAG:
// extract --> extract_artf --> function_1 --> function_1_artf --> metric_1 --> metric_1_artf --> check --> check_artf
//
//	|                      |=> function_3 --> function_3_artf // function_3 takes artf of function_1 and function_2 as inputs
//	|-> function_2 --> function_2_artf --> metric_2 --> metric_2_artf
func (ts *TestSuite) seedComplexWorkflow() (models.DAG, map[string]models.Operator, map[string]models.Artifact) {
	// this gives all op -> op edges by names. We assume each operator is named by `<type>_<index>` format
	// to deduce the type used to create the operator.
	type simpleDependency struct {
		From string
		To   string
		Idx  int
	}

	dependencies := []simpleDependency{
		{
			From: "extract",
			To:   "function_1",
			Idx:  0,
		},
		{
			From: "extract",
			To:   "function_2",
			Idx:  0,
		},
		{
			From: "function_1",
			To:   "function_3",
			Idx:  0,
		},
		{
			From: "function_2",
			To:   "function_3",
			Idx:  1,
		},
		{
			From: "function_1",
			To:   "metric_1",
			Idx:  0,
		},
		{
			From: "function_2",
			To:   "metric_2",
			Idx:  0,
		},
		{
			From: "metric_1",
			To:   "check",
			Idx:  0,
		},
	}

	dag := ts.seedDAG(1)[0]

	fn_spec := operator.NewSpecFromFunction(
		function.Function{},
	)

	extract_spec := operator.NewSpecFromExtract(
		connector.Extract{
			Service:       shared.Postgres,
			IntegrationId: uuid.New(),
			Parameters:    &connector.PostgresExtractParams{},
		},
	)

	metric_spec := operator.NewSpecFromMetric(metric.Metric{})
	check_spec := operator.NewSpecFromCheck(check.Check{})

	operators := make(map[string]models.Operator, len(dependencies))
	artifacts := make(map[string]models.Artifact, len(dependencies))
	createdOperatorNames := make(map[string]bool, len(dependencies))

	for _, dep := range dependencies {
		for _, opName := range []string{dep.From, dep.To} {
			if _, ok := createdOperatorNames[opName]; !ok {
				// create op, artf, and op -> artf edge
				opType := strings.Split(opName, "_")[0]
				spec := extract_spec
				if opType == string(operator.FunctionType) {
					spec = fn_spec
				} else if opType == string(operator.MetricType) {
					spec = metric_spec
				} else if opType == string(operator.CheckType) {
					spec = check_spec
				} else if opType != string(operator.ExtractType) {
					require.True(ts.T(), false, "Unsupported operator type.")
				}

				op, err := ts.operator.Create(
					ts.ctx,
					opName,
					randString(15),
					spec,
					nil,
					ts.DB,
				)
				require.Nil(ts.T(), err)
				operators[opName] = *op

				artfName := fmt.Sprintf("%s_artf", opName)
				artf, err := ts.artifact.Create(
					ts.ctx,
					artfName,
					randString(15),
					shared.UntypedArtifact, // for now it's fine to have all artifacts untyped.
					ts.DB,
				)
				require.Nil(ts.T(), err)
				artifacts[artfName] = *artf

				_, err = ts.dagEdge.Create(
					ts.ctx,
					dag.ID,
					shared.OperatorToArtifactDAGEdge,
					op.ID,
					artf.ID,
					int16(0),
					ts.DB,
				)
				require.Nil(ts.T(), err)

				createdOperatorNames[opName] = true
			}
		}

		// create artf -> op edge based on dependency
		_, err := ts.dagEdge.Create(
			ts.ctx,
			dag.ID,
			shared.ArtifactToOperatorDAGEdge,
			artifacts[fmt.Sprintf("%s_artf", dep.From)].ID,
			operators[dep.To].ID,
			int16(dep.Idx),
			ts.DB,
		)
		require.Nil(ts.T(), err)
	}

	return dag, operators, artifacts
}
