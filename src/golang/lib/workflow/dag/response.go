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
	DagId         uuid.UUID             `json:"dag_id"`
	DagCreatedAt  time.Time             `json:"dag_created_at"`
	StorageConfig *shared.StorageConfig `json:"storage_config"`
	EngineConfig  *shared.EngineConfig  `json:"engine_config"`

	// Workflow metadata
	WorkflowId        uuid.UUID                 `json:"workflow_id"`
	WorkflowCreatedAt time.Time                 `json:"workflow_created_at"`
	UserId            uuid.UUID                 `json:"user_id"`
	Name              string                    `json:"name"`
	Description       string                    `json:"description"`
	Schedule          *workflow.Schedule        `json:"schedule"`
	RetentionPolicy   *workflow.RetentionPolicy `json:"retention_policy"`
}

type Response struct {
	MetadataResponse
	Operators map[uuid.UUID]operator.Response `json:"operators"`
	Artifacts map[uuid.UUID]artifact.Response `json:"artifacts"`
}

type RawResultResponse struct {
	Id uuid.UUID `json:"id"`

	// TODO (ENG-1613, ENG-1614):
	// These will be replaced with ExecutionState with ExecState
	Status    shared.ExecutionStatus `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
}

type ResultResponse struct {
	MetadataResponse
	Result    *RawResultResponse                    `json:"result"`
	Operators map[uuid.UUID]operator.ResultResponse `json:"operators"`
	Artifacts map[uuid.UUID]artifact.ResultResponse `json:"artifacts"`
}

func NewResultResponseFromDbObjects(
	DbWorkflowDag *workflow_dag.DBWorkflowDag,
	DbWorkflowDagResult *workflow_dag_result.WorkflowDagResult,
	DbOperatorResults []operator_result.OperatorResult,
	DbArtifactResults []artifact_result.ArtifactResult,
) *ResultResponse {
	metadataResponse := MetadataResponse{
		DagId:         DbWorkflowDag.Id,
		DagCreatedAt:  DbWorkflowDag.CreatedAt,
		StorageConfig: &DbWorkflowDag.StorageConfig,
		EngineConfig:  &DbWorkflowDag.EngineConfig,

		WorkflowId: DbWorkflowDag.WorkflowId,
	}

	rawResultResponse := RawResultResponse{
		Id:        DbWorkflowDagResult.Id,
		Status:    DbWorkflowDagResult.Status,
		CreatedAt: DbWorkflowDagResult.CreatedAt,
	}

	if DbWorkflowDag.Metadata != nil {
		wfMetadata := DbWorkflowDag.Metadata
		metadataResponse.WorkflowCreatedAt = wfMetadata.CreatedAt
		metadataResponse.UserId = wfMetadata.UserId
		metadataResponse.Name = wfMetadata.Name
		metadataResponse.Description = wfMetadata.Description
		metadataResponse.Schedule = &wfMetadata.Schedule
		metadataResponse.RetentionPolicy = &wfMetadata.RetentionPolicy
	}

	operatorsResponse := make(map[uuid.UUID]operator.ResultResponse)
	artifactsResponse := make(map[uuid.UUID]artifact.ResultResponse)
	for _, opResult := range DbOperatorResults {
		if op, ok := DbWorkflowDag.Operators[opResult.OperatorId]; ok {
			opResultResponse := operator.NewResultResponseFromDbObjects(&op, &opResult)
			operatorsResponse[op.Id] = *opResultResponse
		}
	}

	// Handle operators without results
	for id, op := range DbWorkflowDag.Operators {
		if _, ok := operatorsResponse[id]; !ok {
			operatorsResponse[id] = *(operator.NewResultResponseFromDbObjects(&op, nil))
		}
	}

	for _, artfResult := range DbArtifactResults {
		if artf, ok := DbWorkflowDag.Artifacts[artfResult.ArtifactId]; ok {
			artfResultResponse := artifact.NewResultResponseFromDbObjects(&artf, &artfResult)
			artifactsResponse[artf.Id] = *artfResultResponse
		}
	}

	// Handle artifacts without results
	for id, artf := range DbWorkflowDag.Artifacts {
		if _, ok := artifactsResponse[id]; !ok {
			artifactsResponse[id] = *(artifact.NewResultResponseFromDbObjects(&artf, nil))
		}
	}

	return &ResultResponse{
		MetadataResponse: metadataResponse,
		Result:           &rawResultResponse,
		Operators:        operatorsResponse,
		Artifacts:        artifactsResponse,
	}
}
