package tests

import (
	"context"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/function"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func seedOperator(t *testing.T, count int) []operator.Operator {
	specs := make([]operator.Spec, 0, count)

	for i := 0; i < count; i++ {
		spec := operator.NewSpecFromFunction(function.Function{
			Type:        function.FileFunctionType,
			Language:    "python",
			Granularity: function.TableGranularity,
		})
		specs = append(specs, *spec)
	}

	require.Len(t, specs, count)

	return seedOperatorWithSpecs(t, count, specs)
}

// seedOperatorWithSpecs populates the operator table with count operators
// using the specs provided.
func seedOperatorWithSpecs(t *testing.T, count int, specs []operator.Spec) []operator.Operator {
	require.Len(t, specs, count)

	operators := make([]operator.Operator, 0, count)

	for i := 0; i < count; i++ {
		testOperator, err := writers.operatorWriter.CreateOperator(
			context.Background(),
			randString(10),
			randString(15),
			&specs[i],
			db,
		)
		require.Nil(t, err)

		operators = append(operators, *testOperator)
	}

	require.Len(t, operators, count)

	return operators
}

func TestCreateOperator(t *testing.T) {
	defer resetDatabase(t)

	integrations := seedIntegration(t, 1)

	expectedOperator := &operator.Operator{
		Name:        "test-operator",
		Description: "testing op",
		Spec: *operator.NewSpecFromExtract(connector.Extract{
			Service:       integration.Postgres,
			IntegrationId: integrations[0].Id,
			Parameters: &connector.PostgresExtractParams{
				connector.RelationalDBExtractParams{
					Query: "SELECT * FROM mpg;",
				},
			},
		}),
	}

	actualOperator, err := writers.operatorWriter.CreateOperator(
		context.Background(),
		expectedOperator.Name,
		expectedOperator.Description,
		&expectedOperator.Spec,
		db,
	)
	require.Nil(t, err)
	require.NotEqual(t, uuid.Nil, actualOperator.Id)

	expectedOperator.Id = actualOperator.Id

	requireDeepEqual(t, expectedOperator, actualOperator)
}
