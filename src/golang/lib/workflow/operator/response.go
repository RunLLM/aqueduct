package operator

import (
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
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
	DbOperator *operator.DBOperator,
	DbOperatorResult *operator_result.OperatorResult,
) *ResultResponse {
	metadata := Response{
		Id:          DbOperator.Id,
		Name:        DbOperator.Name,
		Description: DbOperator.Description,
		Spec:        &DbOperator.Spec,
		Inputs:      DbOperator.Inputs,
		Outputs:     DbOperator.Outputs,
	}

	if DbOperatorResult == nil {
		return &ResultResponse{Response: metadata}
	}

	var execState *shared.ExecutionState = nil
	if !DbOperatorResult.ExecState.IsNull {
		execState = &DbOperatorResult.ExecState.ExecutionState
	}

	return &ResultResponse{
		Response: metadata,
		Result: &RawResultResponse{
			Id:        DbOperatorResult.Id,
			ExecState: execState,
		},
	}
}
