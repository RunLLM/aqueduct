package operator_test

import (
	"encoding/json"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/function"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSerializingAndDeserializingOperator(t *testing.T) {
	id := uuid.New()

	op := operator.DBOperator{
		Id:   id,
		Name: "test",
		Spec: *operator.NewSpecFromFunction(
			function.Function{
				Language:    "eng",
				Type:        function.FileFunctionType,
				Granularity: function.TableGranularity,
			},
		),
	}

	rawOp, err := json.Marshal(op)
	require.Nil(t, err)

	var reconstructedOp operator.DBOperator
	err = json.Unmarshal(rawOp, &reconstructedOp)
	require.Nil(t, err)
	require.True(t, reconstructedOp.Spec.IsFunction())
	require.NotNil(t, reconstructedOp.Spec.Function())
}
