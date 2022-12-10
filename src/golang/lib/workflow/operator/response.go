package operator

import (
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
)

type Response struct {
	Id          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Spec        *operator.Spec `json:"spec"`

	Inputs  []uuid.UUID `json:"inputs"`
	Outputs []uuid.UUID `json:"outputs"`
}

type RawResultResponse struct {
	// Contains only the `result`. It mostly mirrors 'operator_result' schema.
	Id        uuid.UUID              `json:"id"`
	ExecState *shared.ExecutionState `json:"exec_state"`
}

type ResultResponse struct {
	Response
	Result *RawResultResponse `json:"result"`
}

func NewResultResponseFromDbObjects(
	dbOperator *models.Operator,
	dbOperatorResult *models.OperatorResult,
) *ResultResponse {
	// make a value copy of `Spec` field
	spec := dbOperator.Spec
	metadata := Response{
		Id:          dbOperator.ID,
		Name:        dbOperator.Name,
		Description: dbOperator.Description,
		Spec:        &spec,
		Inputs:      dbOperator.Inputs,
		Outputs:     dbOperator.Outputs,
	}

	if dbOperatorResult == nil {
		return &ResultResponse{Response: metadata}
	}

	var execState *shared.ExecutionState = nil
	if !dbOperatorResult.ExecState.IsNull {
		// make a value copy of execState
		execStateVal := dbOperatorResult.ExecState.ExecutionState
		execState = &execStateVal
	}

	return &ResultResponse{
		Response: metadata,
		Result: &RawResultResponse{
			Id:        dbOperatorResult.ID,
			ExecState: execState,
		},
	}
}
