package operator

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/views"
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

func NewResponseFromDbObject(dbOperator *models.Operator) *Response {
	// make a value copy of `Spec` field
	spec := dbOperator.Spec
	return &Response{
		Id:          dbOperator.ID,
		Name:        dbOperator.Name,
		Description: dbOperator.Description,
		Spec:        &spec,
		Inputs:      dbOperator.Inputs,
		Outputs:     dbOperator.Outputs,
	}
}

func NewResultResponseFromDbObjects(
	dbOperator *models.Operator,
	dbOperatorResult *models.OperatorResult,
) *ResultResponse {
	resp := NewResponseFromDbObject(dbOperator)

	if dbOperatorResult == nil {
		return &ResultResponse{Response: *resp}
	}

	var execState *shared.ExecutionState = nil
	if !dbOperatorResult.ExecState.IsNull {
		// make a value copy of execState
		execStateVal := dbOperatorResult.ExecState.ExecutionState
		execState = &execStateVal
	}

	return &ResultResponse{
		Response: *resp,
		Result: &RawResultResponse{
			Id:        dbOperatorResult.ID,
			ExecState: execState,
		},
	}
}

func NewResultResponseFromDBView(
	dbViewOpWithResult *views.OperatorWithResult,
) *ResultResponse {
	return NewResultResponseFromDbObjects(
		&models.Operator{
			ID:                     dbViewOpWithResult.ID,
			Name:                   dbViewOpWithResult.Name,
			Description:            dbViewOpWithResult.Description,
			Spec:                   dbViewOpWithResult.Spec,
			ExecutionEnvironmentID: dbViewOpWithResult.ExecutionEnvironmentID,
		},
		&models.OperatorResult{
			ID:          dbViewOpWithResult.ResultID,
			DAGResultID: dbViewOpWithResult.DAGResultID,
			OperatorID:  dbViewOpWithResult.ID,
			Status:      dbViewOpWithResult.Status,
			ExecState:   dbViewOpWithResult.ExecState,
		},
	)
}
