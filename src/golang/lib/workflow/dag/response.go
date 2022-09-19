package dag

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
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

	ExecState *shared.ExecutionState `json:"exec_state"`
}

type ResultResponse struct {
	MetadataResponse
	Result    *RawResultResponse                    `json:"result"`
	Operators map[uuid.UUID]operator.ResultResponse `json:"operators"`
	Artifacts map[uuid.UUID]artifact.ResultResponse `json:"artifacts"`
}

func NewResultResponseFromDbObjects(
	dbWorkflowDag *workflow_dag.DBWorkflowDag,
	dbWorkflowDagResult *workflow_dag_result.WorkflowDagResult,
	dbOperatorResults []operator_result.OperatorResult,
	dbArtifactResults []artifact_result.ArtifactResult,
	contents map[string]string,
) *ResultResponse {
	metadataResponse := MetadataResponse{
		DagId:         dbWorkflowDag.Id,
		DagCreatedAt:  dbWorkflowDag.CreatedAt,
		StorageConfig: dbWorkflowDag.StorageConfig,
		EngineConfig:  dbWorkflowDag.EngineConfig,

		WorkflowId: dbWorkflowDag.WorkflowId,
	}

	rawResultResponse := RawResultResponse{
		Id: dbWorkflowDagResult.Id,
	}

	if !dbWorkflowDagResult.ExecState.IsNull {
		// make a value copy of execState
		execStateVal := dbWorkflowDagResult.ExecState.ExecutionState
		rawResultResponse.ExecState = &execStateVal
	}

	if dbWorkflowDag.Metadata != nil {
		wfMetadata := dbWorkflowDag.Metadata
		metadataResponse.WorkflowCreatedAt = wfMetadata.CreatedAt
		metadataResponse.UserId = wfMetadata.UserId
		metadataResponse.Name = wfMetadata.Name
		metadataResponse.Description = wfMetadata.Description
		metadataResponse.Schedule = wfMetadata.Schedule
		metadataResponse.RetentionPolicy = wfMetadata.RetentionPolicy
	}

	operatorsResponse := make(map[uuid.UUID]operator.ResultResponse, len(dbWorkflowDag.Operators))
	artifactsResponse := make(map[uuid.UUID]artifact.ResultResponse, len(dbWorkflowDag.Artifacts))
	artifactToUpstreamOpId := make(map[uuid.UUID]uuid.UUID, len(dbWorkflowDag.Artifacts))
	artifactToDownstreamOpIds := make(map[uuid.UUID][]uuid.UUID, len(dbWorkflowDag.Artifacts))

	for _, opResult := range dbOperatorResults {
		if op, ok := dbWorkflowDag.Operators[opResult.OperatorId]; ok {
			opResultResponse := operator.NewResultResponseFromDbObjects(&op, &opResult)
			operatorsResponse[op.Id] = *opResultResponse
		}
	}

	// Handle operators without results and update artifact maps
	for id, op := range dbWorkflowDag.Operators {
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
		if artf, ok := dbWorkflowDag.Artifacts[artfResult.ArtifactId]; ok {
			content, ok := contents[artfResult.ContentPath]
			var contentPtr *string = nil
			if ok {
				contentPtr = &content
			}

			artfResultResponse := artifact.NewResultResponseFromDbObjects(
				&artf,
				&artfResult,
				contentPtr,
				artifactToUpstreamOpId[artf.Id],
				artifactToDownstreamOpIds[artf.Id],
			)
			artifactsResponse[artf.Id] = *artfResultResponse
		}
	}

	// Handle artifacts without results
	for id, artf := range dbWorkflowDag.Artifacts {
		if _, ok := artifactsResponse[id]; !ok {
			artifactsResponse[id] = *(artifact.NewResultResponseFromDbObjects(
				&artf,
				nil,
				nil,
				artifactToUpstreamOpId[artf.Id],
				artifactToDownstreamOpIds[artf.Id],
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
