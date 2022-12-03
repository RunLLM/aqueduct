package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models/shared"
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
	dags := ts.seedDAG(1)
	dag := dags[0]

	expectedOperators := ts.seedOperatorWithDAG(3, dag.ID, shared.FunctionType)

	actualOperators, err := ts.operator.GetByDAG(ts.ctx, dag.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualOperators(ts.T(), expectedOperators, actualOperators)
}

func (ts *TestSuite) TestOperator_GetDistinctLoadOPsByWorkflow() {
	// TODO: Requires integration tests to be implemented
}

func (ts *TestSuite) TestOperator_GetLoadOPsByWorkflowAndIntegration() {
	//GetLoadOPsByWorkflowAndIntegration(
	// 	ctx context.Context,
	// 	workflowID uuid.UUID,
	// 	integrationID uuid.UUID,
	// 	objectName string,
	// 	DB database.Database,
	// ) ([]models.Operator, error)
}

func (ts *TestSuite) TestOperator_GetLoadOPsByIntegration() {
	//GetLoadOPsByIntegration(
	// 	ctx context.Context,
	// 	integrationID uuid.UUID,
	// 	objectName string,
	// 	DB database.Database,
	// ) ([]models.Operator, error)
}

func (ts *TestSuite) TestOperator_ValidateOrg() {
	//ValidateOrg(ctx context.Context, operatorId uuid.UUID, orgID string, DB database.Database) (bool, error)

}

func (ts *TestSuite) TestOperator_Create() {
	//Create(
	// 	ctx context.Context,
	// 	name string,
	// 	description string,
	// 	spec *shared.Spec,
	// 	DB database.Database,
	// ) (*models.Operator, error)

}

func (ts *TestSuite) TestOperator_Delete() {
	//Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

}

func (ts *TestSuite) TestOperator_DeleteBatch() {
	//DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error

}

func (ts *TestSuite) TestOperator_Update() {
	//Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.Operator, error)

}
