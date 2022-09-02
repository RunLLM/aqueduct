package artifact

import (
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/google/uuid"
)

type Response struct {
	Id          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        artifact.Type `json:"type"`
	// Once we clean up DBArtifact we should include inputs / outputs fields here.
}

type RawResultResponse struct {
	Id          uuid.UUID              `json:"id"`
	ContentPath string                 `json:"content_path"`
	ExecState   *shared.ExecutionState `json:"exec_state"`
}

type ResultResponse struct {
	Response
	Result *RawResultResponse `json:"result"`
}

func NewResultResponseFromDbObjects(
	DbArtifact *artifact.DBArtifact,
	DbArtifactResult *artifact_result.ArtifactResult,
) *ResultResponse {
	metadata := Response{
		Id:          DbArtifact.Id,
		Name:        DbArtifact.Name,
		Description: DbArtifact.Description,
		Type:        DbArtifact.Type,
	}

	if DbArtifactResult == nil {
		return &ResultResponse{Response: metadata}
	}

	var execState *shared.ExecutionState = nil
	if !DbArtifactResult.ExecState.IsNull {
		execState = &DbArtifactResult.ExecState.ExecutionState
	}

	return &ResultResponse{
		Response: metadata,
		Result: &RawResultResponse{
			Id:          DbArtifactResult.Id,
			ContentPath: DbArtifactResult.ContentPath,
			ExecState:   execState,
		},
	}
}
