package dag

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/google/uuid"
)

type MetadataResponse struct {
	// Dag metadata
	DagId         uuid.UUID            `json:"dag_id"`
	DagCreatedAt  time.Time            `json:"dag_created_at"`
	StorageConfig shared.StorageConfig `json:"storage_config"`
	EngineConfig  shared.EngineConfig  `json:"engine_config"`

	// Workflow metadata
	WorkflowId        uuid.UUID              `json:"workflow_id"`
	WorkflowCreatedAt time.Time              `json:"workflow_created_at"`
	UserId            uuid.UUID              `json:"user_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Schedule          shared.Schedule        `json:"schedule"`
	RetentionPolicy   shared.RetentionPolicy `json:"retention_policy"`
}

type Response struct {
	MetadataResponse
	Operators map[uuid.UUID]operator.Response `json:"operators"`
	Artifacts map[uuid.UUID]artifact.Response `json:"artifacts"`
}

type RawResultResponse struct {
	// Contains only the `result`. It mostly mirrors 'workflow_dag_result' schema.
	Id uuid.UUID `json:"id"`

	ExecState *shared.ExecutionState `json:"exec_state"`
}

type ResultResponse struct {
	MetadataResponse
	Result    *RawResultResponse                    `json:"result"`
	Operators map[uuid.UUID]operator.ResultResponse `json:"operators"`
	Artifacts map[uuid.UUID]artifact.ResultResponse `json:"artifacts"`
}

func NewResponseFromDbObjects(dag *models.DAG) *Response {
	metadataResponse := MetadataResponse{
		DagId:         dag.ID,
		DagCreatedAt:  dag.CreatedAt,
		StorageConfig: dag.StorageConfig,
		EngineConfig:  dag.EngineConfig,

		WorkflowId: dag.WorkflowID,
	}

	operatorsResponse := make(map[uuid.UUID]operator.Response, len(dag.Operators))
	artifactsResponse := make(map[uuid.UUID]artifact.Response, len(dag.Artifacts))
	artifactToUpstreamOpId := make(map[uuid.UUID]uuid.UUID, len(dag.Artifacts))
	artifactToDownstreamOpIds := make(map[uuid.UUID][]uuid.UUID, len(dag.Artifacts))

	if dag.Metadata != nil {
		wfMetadata := dag.Metadata
		metadataResponse.WorkflowCreatedAt = wfMetadata.CreatedAt
		metadataResponse.UserId = wfMetadata.UserID
		metadataResponse.Name = wfMetadata.Name
		metadataResponse.Description = wfMetadata.Description
		metadataResponse.Schedule = wfMetadata.Schedule
		metadataResponse.RetentionPolicy = wfMetadata.RetentionPolicy
	}

	// Handle operators without results and update artifact maps
	for id, op := range dag.Operators {
		if _, ok := operatorsResponse[id]; !ok {
			operatorsResponse[id] = *(operator.NewResponseFromDbObject(&op))
		}

		for _, id := range op.Outputs {
			artifactToUpstreamOpId[id] = op.ID
		}

		for _, id := range op.Inputs {
			artifactToDownstreamOpIds[id] = append(artifactToDownstreamOpIds[id], op.ID)
		}
	}

	// Handle artifacts without results
	for id, artf := range dag.Artifacts {
		artifactsResponse[id] = *(artifact.NewResponseFromDbObject(
			&artf,
			artifactToUpstreamOpId[artf.ID],
			artifactToDownstreamOpIds[artf.ID],
		))
	}

	return &Response{
		MetadataResponse: metadataResponse,
		Operators:        operatorsResponse,
		Artifacts:        artifactsResponse,
	}
}

func NewResultResponseFromDbObjects(
	dag *models.DAG,
	dagResult *models.DAGResult,
	dbOperatorResults []models.OperatorResult,
	dbArtifactResults []models.ArtifactResult,
	contents map[string]string,
) *ResultResponse {
	resp := NewResponseFromDbObjects(dag)
	rawResultResponse := RawResultResponse{
		Id: dagResult.ID,
	}

	if !dagResult.ExecState.IsNull {
		// make a value copy of execState
		execStateVal := dagResult.ExecState.ExecutionState
		rawResultResponse.ExecState = &execStateVal
	}

	operatorsResponse := make(map[uuid.UUID]operator.ResultResponse, len(dag.Operators))
	artifactsResponse := make(map[uuid.UUID]artifact.ResultResponse, len(dag.Artifacts))

	for _, opResult := range dbOperatorResults {
		if op, ok := dag.Operators[opResult.OperatorID]; ok {
			opResultResponse := operator.NewResultResponseFromDbObjects(&op, &opResult)
			operatorsResponse[op.ID] = *opResultResponse
		}
	}

	// Handle operators without results and update artifact maps
	for id, op := range dag.Operators {
		if _, ok := operatorsResponse[id]; !ok {
			operatorsResponse[id] = *(operator.NewResultResponseFromDbObjects(&op, nil))
		}
	}

	for _, artfResult := range dbArtifactResults {
		if artf, ok := dag.Artifacts[artfResult.ArtifactID]; ok {
			content, ok := contents[artfResult.ContentPath]
			var contentPtr *string = nil
			if ok {
				contentPtr = &content
			}

			artfResultResponse := artifact.NewResultResponseFromDbObjects(
				&artf,
				&artfResult,
				contentPtr,
				resp.Artifacts[artf.ID].From,
				resp.Artifacts[artf.ID].To,
			)
			artifactsResponse[artf.ID] = *artfResultResponse
		}
	}

	// Handle artifacts without results
	for id, artf := range dag.Artifacts {
		if _, ok := artifactsResponse[id]; !ok {
			artifactsResponse[id] = *(artifact.NewResultResponseFromDbObjects(
				&artf,
				nil,
				nil,
				resp.Artifacts[artf.ID].From,
				resp.Artifacts[artf.ID].To,
			))
		}
	}

	return &ResultResponse{
		MetadataResponse: resp.MetadataResponse,
		Result:           &rawResultResponse,
		Operators:        operatorsResponse,
		Artifacts:        artifactsResponse,
	}
}
