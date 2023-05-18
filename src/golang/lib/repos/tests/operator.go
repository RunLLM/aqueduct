package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/function"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestOperator_Exists() {
	operators := ts.seedOperator(1)
	operator := operators[0]

	exists, err := ts.operator.Exists(ts.ctx, operator.ID, ts.DB)
	require.Nil(ts.T(), err)
	require.True(ts.T(), exists)

	// Check for non-existent operator
	exists, err = ts.operator.Exists(ts.ctx, uuid.Nil, ts.DB)
	require.Nil(ts.T(), err)
	require.False(ts.T(), exists)
}

func (ts *TestSuite) TestOperator_Get() {
	operators := ts.seedOperator(1)
	expectedOperator := operators[0]

	actualOperator, err := ts.operator.Get(ts.ctx, expectedOperator.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedOperator, *actualOperator)
}

func (ts *TestSuite) TestOperator_GetNode() {
	_, _, _, expectedOpNodes, _ := ts.seedComplexWorkflow()
	for _, expectedOp := range expectedOpNodes {
		actualOp, err := ts.operator.GetNode(ts.ctx, expectedOp.ID, ts.DB)
		require.Nil(ts.T(), err)
		requireDeepEqual(ts.T(), expectedOp, *actualOp)
	}
}

func (ts *TestSuite) TestOperator_GetBatch() {
	expectedOperators := ts.seedOperator(3)

	IDs := make([]uuid.UUID, 0, len(expectedOperators))
	for _, expectedOperator := range expectedOperators {
		IDs = append(IDs, expectedOperator.ID)
	}

	actualOperators, err := ts.operator.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualOperators(ts.T(), expectedOperators, actualOperators)
}

func (ts *TestSuite) TestOperator_GetByDAG() {
	users := ts.seedUser(1)
	user := users[0]
	dags := ts.seedDAGWithUser(1, user)
	dag := dags[0]

	expectedOperators := ts.seedOperatorWithDAG(3, dag.ID, user.ID, operator.FunctionType)

	actualOperators, err := ts.operator.GetByDAG(ts.ctx, dag.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualOperators(ts.T(), expectedOperators, actualOperators)
}

func (ts *TestSuite) TestOperator_GetNodesByDAG() {
	dag, _, _, expectedOpNodes, _ := ts.seedComplexWorkflow()
	actualOpNodes, err := ts.operator.GetNodesByDAG(ts.ctx, dag.ID, ts.DB)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), len(expectedOpNodes), len(actualOpNodes))
	for _, actualOp := range actualOpNodes {
		expectedOp, ok := expectedOpNodes[actualOp.Name]
		require.True(ts.T(), ok)
		requireDeepEqual(ts.T(), expectedOp, actualOp)
	}
}

func (ts *TestSuite) TestOperator_GetDistinctLoadOPsByWorkflow() {
	users := ts.seedUser(1)
	user := users[0]
	dags := ts.seedDAGWithUser(1, user)
	dag := dags[0]

	expectedOperators := ts.seedOperatorWithDAG(3, dag.ID, user.ID, operator.LoadType)

	expectedLoadOperators := make([]views.LoadOperator, 0, len(expectedOperators))
	for _, expectedLoadOperator := range expectedOperators {
		load := expectedLoadOperator.Spec.Load()
		loadParams := load.Parameters
		resource, err := ts.resource.Get(ts.ctx, load.ResourceId, ts.DB)
		require.Nil(ts.T(), err)

		expectedLoadOperators = append(expectedLoadOperators, views.LoadOperator{
			OperatorID:   expectedLoadOperator.ID,
			OperatorName: expectedLoadOperator.Name,
			ModifiedAt:   dag.CreatedAt,
			ResourceName: resource.Name,
			Spec: connector.Load{
				Service:    testResourceService,
				ResourceId: resource.ID,
				Parameters: loadParams,
			},
		})
	}

	actualOperators, err := ts.operator.GetDistinctLoadOPsByWorkflow(ts.ctx, dag.WorkflowID, ts.DB)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 3, len(actualOperators))
	requireDeepEqualLoadOperators(ts.T(), expectedLoadOperators, actualOperators)
}

func (ts *TestSuite) TestOperator_GetLoadOPsByWorkflowAndResource() {
	users := ts.seedUser(1)
	user := users[0]
	dags := ts.seedDAGWithUser(1, user)
	dag := dags[0]

	operators := ts.seedOperatorWithDAG(3, dag.ID, user.ID, operator.LoadType)

	load := operators[0].Spec.Load()
	loadParams := load.Parameters
	relationalLoad, ok := connector.CastToRelationalDBLoadParams(loadParams)
	require.True(ts.T(), ok)
	resource, err := ts.resource.Get(ts.ctx, load.ResourceId, ts.DB)
	require.Nil(ts.T(), err)

	actualOperators, err := ts.operator.GetLoadOPsByWorkflowAndResource(ts.ctx, dag.WorkflowID, resource.ID, relationalLoad.Table, ts.DB)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 1, len(actualOperators))
	requireDeepEqualOperators(ts.T(), []models.Operator{operators[0]}, actualOperators)
}

func (ts *TestSuite) TestOperator_GetLoadOPsByResource() {
	users := ts.seedUser(1)
	user := users[0]
	dags := ts.seedDAGWithUser(1, user)
	dag := dags[0]

	operators := ts.seedOperatorWithDAG(3, dag.ID, user.ID, operator.LoadType)

	load := operators[0].Spec.Load()
	loadParams := load.Parameters
	relationalLoad, ok := connector.CastToRelationalDBLoadParams(loadParams)
	require.True(ts.T(), ok)
	resource, err := ts.resource.Get(ts.ctx, load.ResourceId, ts.DB)
	require.Nil(ts.T(), err)

	actualOperators, err := ts.operator.GetLoadOPsByResource(ts.ctx, resource.ID, relationalLoad.Table, ts.DB)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 1, len(actualOperators))
	requireDeepEqualOperators(ts.T(), []models.Operator{operators[0]}, actualOperators)
}

func (ts *TestSuite) TestOperator_GetByEngineResourceID() {
	users := ts.seedUser(1)
	user := users[0]
	dags := ts.seedDAGWithUser(1, user)
	dag := dags[0]

	operators := ts.seedOperatorWithDAG(2, dag.ID, user.ID, operator.FunctionType)
	k8sOperator := operators[0]
	lambdaOperator := operators[1]

	lambdaResourceID := uuid.New()
	k8sResourceID := uuid.New()
	_, err := ts.dag.Update(
		ts.ctx,
		dag.ID,
		map[string]interface{}{
			models.DagEngineConfig: &shared.EngineConfig{
				Type: shared.LambdaEngineType,
				LambdaConfig: &shared.LambdaConfig{
					ResourceID: lambdaResourceID,
				},
			},
		},
		ts.DB,
	)
	require.Nil(ts.T(), err)

	k8sOpSpec := k8sOperator.Spec.SetEngineConfig(
		&shared.EngineConfig{
			Type: shared.K8sEngineType,
			K8sConfig: &shared.K8sConfig{
				ResourceID: k8sResourceID,
			},
		},
	)

	_, err = ts.operator.Update(
		ts.ctx,
		k8sOperator.ID,
		map[string]interface{}{
			models.OperatorSpec: k8sOpSpec,
		},
		ts.DB,
	)
	require.Nil(ts.T(), err)

	operators, err = ts.operator.GetByEngineResourceID(
		ts.ctx, lambdaResourceID, ts.DB,
	)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 1, len(operators))
	require.Equal(ts.T(), lambdaOperator.ID, operators[0].ID)

	operators, err = ts.operator.GetByEngineResourceID(
		ts.ctx, k8sResourceID, ts.DB,
	)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 1, len(operators))
	require.Equal(ts.T(), k8sOperator.ID, operators[0].ID)
}

func (ts *TestSuite) TestOperator_ValidateOrg() {
	users := ts.seedUser(1)
	user := users[0]
	dags := ts.seedDAGWithUser(1, user)
	dag := dags[0]

	operators := ts.seedOperatorWithDAG(1, dag.ID, user.ID, operator.LoadType)
	operator := operators[0]

	valid, validErr := ts.operator.ValidateOrg(ts.ctx, operator.ID, testOrgID, ts.DB)
	require.Nil(ts.T(), validErr)
	require.True(ts.T(), valid)

	invalid, invalidErr := ts.operator.ValidateOrg(ts.ctx, operator.ID, randString(10), ts.DB)
	require.Nil(ts.T(), invalidErr)
	require.False(ts.T(), invalid)
}

func (ts *TestSuite) TestOperator_GetUnusedCondaEnvNames() {
	artifactID := uuid.New()
	users := ts.seedUser(1)
	userIDs := sampleUserIDs(1, users)
	workflows := ts.seedWorkflowWithUser(1, userIDs)
	workflowIDs := sampleWorkflowIDs(1, workflows)
	dags := ts.seedDAGWithWorkflow(2, []uuid.UUID{workflowIDs[0], workflowIDs[0]})
	historicalOp := ts.seedOperatorAndDAGOperatorToArtifact(artifactID, dags[0].ID, operator.FunctionType)
	historicalOp.Spec.SetEngineConfig(&shared.EngineConfig{
		Type: shared.AqueductCondaEngineType,
		AqueductCondaConfig: &shared.AqueductCondaConfig{
			Env: "historical",
		},
	})
	_, err := ts.operator.Update(
		ts.ctx,
		historicalOp.ID,
		map[string]interface{}{
			"spec": &historicalOp.Spec,
		},
		ts.DB,
	)
	require.Nil(ts.T(), err)

	latestOp := ts.seedOperatorAndDAGOperatorToArtifact(artifactID, dags[1].ID, operator.FunctionType)
	latestOp.Spec.SetEngineConfig(&shared.EngineConfig{
		Type: shared.AqueductCondaEngineType,
		AqueductCondaConfig: &shared.AqueductCondaConfig{
			Env: "latest",
		},
	})
	_, err = ts.operator.Update(
		ts.ctx,
		latestOp.ID,
		map[string]interface{}{
			"spec": &latestOp.Spec,
		},
		ts.DB,
	)
	require.Nil(ts.T(), err)

	names, err := ts.operator.GetUnusedCondaEnvNames(ts.ctx, ts.DB)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), len(names), 1)
	require.Equal(ts.T(), names[0], "historical")
}

func (ts *TestSuite) TestOperator_GetEngineTypesMapByDagIDs() {
	users := ts.seedUser(1)
	userIDs := sampleUserIDs(1, users)
	workflows := ts.seedWorkflowWithUser(1, userIDs)
	workflowIDs := sampleWorkflowIDs(1, workflows)
	dags := ts.seedDAGWithWorkflow(2, []uuid.UUID{workflowIDs[0], workflowIDs[0]})

	operators := ts.seedOperatorWithDAG(1, dags[0].ID, users[0].ID, operator.FunctionType)
	k8sOp := operators[0]
	k8sOp.Spec.SetEngineConfig(&shared.EngineConfig{
		Type: shared.K8sEngineType,
	})
	_, err := ts.operator.Update(
		ts.ctx,
		k8sOp.ID,
		map[string]interface{}{
			"spec": &k8sOp.Spec,
		},
		ts.DB,
	)
	require.Nil(ts.T(), err)

	operators = ts.seedOperatorWithDAG(1, dags[0].ID, users[0].ID, operator.FunctionType)
	databricksOp := operators[0]
	databricksOp.Spec.SetEngineConfig(&shared.EngineConfig{
		Type: shared.DatabricksEngineType,
	})
	_, err = ts.operator.Update(
		ts.ctx,
		databricksOp.ID,
		map[string]interface{}{
			"spec": &databricksOp.Spec,
		},
		ts.DB,
	)
	require.Nil(ts.T(), err)

	operators = ts.seedOperatorWithDAG(2, dags[1].ID, users[0].ID, operator.FunctionType)
	sparkOp := operators[0]
	sparkOp.Spec.SetEngineConfig(&shared.EngineConfig{
		Type: shared.SparkEngineType,
	})
	_, err = ts.operator.Update(
		ts.ctx,
		sparkOp.ID,
		map[string]interface{}{
			"spec": &sparkOp.Spec,
		},
		ts.DB,
	)
	require.Nil(ts.T(), err)

	dagIDToEngineTypes, err := ts.operator.GetEngineTypesMapByDagIDs(
		ts.ctx,
		[]uuid.UUID{dags[0].ID, dags[1].ID},
		ts.DB,
	)
	require.Nil(ts.T(), err)
	actualDag0Types := dagIDToEngineTypes[dags[0].ID]
	require.Equal(ts.T(), len(actualDag0Types), 2)
	require.Contains(ts.T(), actualDag0Types, shared.K8sEngineType)
	require.Contains(ts.T(), actualDag0Types, shared.DatabricksEngineType)

	actualDag1Types := dagIDToEngineTypes[dags[1].ID]
	require.Equal(ts.T(), len(actualDag1Types), 2)
	require.Contains(ts.T(), actualDag1Types, shared.SparkEngineType)
	require.Contains(ts.T(), actualDag1Types, shared.EngineType(""))
}

func (ts *TestSuite) TestOperator_Create() {
	name := randString(10)
	description := randString(10)
	spec := operator.NewSpecFromFunction(
		function.Function{},
	)
	expectedOperator := &models.Operator{
		Name:        name,
		Description: description,
		Spec:        *spec,
		ExecutionEnvironmentID: utils.NullUUID{
			IsNull: true,
		},
	}
	actualOperator, err := ts.operator.Create(ts.ctx, name, description, spec, nil, ts.DB)
	expectedOperator.ID = actualOperator.ID
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedOperator, actualOperator)
}

func (ts *TestSuite) TestOperator_Delete() {
	users := ts.seedUser(1)
	user := users[0]
	dags := ts.seedDAGWithUser(1, user)
	dag := dags[0]

	operators := ts.seedOperatorWithDAG(1, dag.ID, user.ID, operator.LoadType)
	operator := operators[0]

	err := ts.operator.Delete(ts.ctx, operator.ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestOperator_DeleteBatch() {
	users := ts.seedUser(1)
	user := users[0]
	dags := ts.seedDAGWithUser(1, user)
	dag := dags[0]

	operators := ts.seedOperatorWithDAG(3, dag.ID, user.ID, operator.LoadType)
	IDs := make([]uuid.UUID, 0, len(operators))
	for _, operator := range operators {
		IDs = append(IDs, operator.ID)
	}

	err := ts.operator.DeleteBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestOperator_Update() {
	users := ts.seedUser(1)
	user := users[0]
	dags := ts.seedDAGWithUser(1, user)
	dag := dags[0]

	operators := ts.seedOperatorWithDAG(1, dag.ID, user.ID, operator.LoadType)
	name := randString(10)
	spec := operator.NewSpecFromFunction(
		function.Function{},
	)
	changes := map[string]interface{}{
		models.OperatorName: name,
		models.OperatorSpec: spec,
	}

	newOperator, err := ts.operator.Update(ts.ctx, operators[0].ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), name, newOperator.Name)
	requireDeepEqual(ts.T(), spec, &newOperator.Spec)
}
