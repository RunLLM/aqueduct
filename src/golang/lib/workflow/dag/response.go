package dag

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/models"
	mdl_shared "github.com/aqueducthq/aqueduct/lib/models/shared"
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
	WorkflowId        uuid.UUID                `json:"workflow_id"`
	WorkflowCreatedAt time.Time                `json:"workflow_created_at"`
	UserId            uuid.UUID                `json:"user_id"`
	Name              string                   `json:"name"`
	Description       string                   `json:"description"`
	Schedule          workflow.Schedule        `json:"schedule"`
	RetentionPolicy   workflow.RetentionPolicy `json:"retention_policy"`
}

type Response struct {
	MetadataResponse
	Operators map[uuid.UUID]operator.Response `json:"operators"`
	Artifacts map[uuid.UUID]artifact.Response `json:"artifacts"`
}

type RawResultResponse struct {
	// Contains only the `result`. It mostly mirrors 'workflow_dag_result' schema.
	Id uuid.UUID `json:"id"`

	ExecState *mdl_shared.ExecutionState `json:"exec_state"`
}

type ResultResponse struct {
	MetadataResponse
	Result    *RawResultResponse                    `json:"result"`
	Operators map[uuid.UUID]operator.ResultResponse `json:"operators"`
	Artifacts map[uuid.UUID]artifact.ResultResponse `json:"artifacts"`
}

func NewResultResponseFromDbObjects(
	dag *models.DAG,
	dagResult *models.DAGResult,
	dbOperatorResults []models.OperatorResult,
	dbArtifactResults []models.ArtifactResult,
	contents map[string]string,
) *ResultResponse {
	metadataResponse := MetadataResponse{
		DagId:         dag.ID,
		DagCreatedAt:  dag.CreatedAt,
		StorageConfig: dag.StorageConfig,
		EngineConfig:  dag.EngineConfig,

		WorkflowId: dag.WorkflowID,
	}

	rawResultResponse := RawResultResponse{
		Id: dagResult.ID,
	}

	if !dagResult.ExecState.IsNull {
		// make a value copy of execState
		execStateVal := dagResult.ExecState.ExecutionState
		rawResultResponse.ExecState = &execStateVal
	}

	if dag.Metadata != nil {
		wfMetadata := dag.Metadata
		metadataResponse.WorkflowCreatedAt = wfMetadata.CreatedAt
		metadataResponse.UserId = wfMetadata.UserID
		metadataResponse.Name = wfMetadata.Name
		metadataResponse.Description = wfMetadata.Description
		metadataResponse.Schedule = wfMetadata.Schedule
		metadataResponse.RetentionPolicy = wfMetadata.RetentionPolicy
	}

	operatorsResponse := make(map[uuid.UUID]operator.ResultResponse, len(dag.Operators))
	artifactsResponse := make(map[uuid.UUID]artifact.ResultResponse, len(dag.Artifacts))
	artifactToUpstreamOpId := make(map[uuid.UUID]uuid.UUID, len(dag.Artifacts))
	artifactToDownstreamOpIds := make(map[uuid.UUID][]uuid.UUID, len(dag.Artifacts))

	for _, opResult := range dbOperatorResults {
		if op, ok := dag.Operators[opResult.OperatorID]; ok {
			opResultResponse := operator.NewResultResponseFromDbObjects(&op, &opResult)
			operatorsResponse[op.Id] = *opResultResponse
		}
	}

	// Handle operators without results and update artifact maps
	for id, op := range dag.Operators {
		if _, ok := operatorsResponse[id]; !ok {
			operatorsResponse[id] = *(operator.NewResultResponseFromDbObjects(&op, nil))
		}

		for _, id := range op.Outputs {
			artifactToUpstreamOpId[id] = op.Id
		}

		for _, id := range op.Inputs {
			artifactToDownstreamOpIds[id] = append(artifactToDownstreamOpIds[id], op.Id)
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
				artifactToUpstreamOpId[artf.ID],
				artifactToDownstreamOpIds[artf.ID],
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
				artifactToUpstreamOpId[artf.ID],
				artifactToDownstreamOpIds[artf.ID],
			))
		}
	}

	return &ResultResponse{
		MetadataResponse: metadataResponse,
		Result:           &rawResultResponse,
		Operators:        operatorsResponse,
		Artifacts:        artifactsResponse,
	}
}
