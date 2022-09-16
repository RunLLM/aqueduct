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

	// upstream operator ID, must be unique.
	From uuid.UUID `json:"from"`

	// downstream operator IDs, could be multiple or empty.
	To []uuid.UUID `json:"to"`
}

type RawResultResponse struct {
	// Contains only the `result`. It mostly mirrors 'artifact_result' schema.
	Id uuid.UUID `json:"id"`

	// If `ContentSerialized` is set, the content is small and we directly send
	// it as a part of response. It's consistent with the object stored in `ContentPath`.
	//
	// Otherwise, the content is large and
	// one should send an additional request to fetch the content.
	ContentPath       string  `json:"content_path"`
	ContentSerialized *string `json:"content_serialized"`

	ExecState *shared.ExecutionState `json:"exec_state"`
}

type ResultResponse struct {
	Response
	Result *RawResultResponse `json:"result"`
}

func NewRawResultResponseFromDbObject(
	dbArtifactResult *artifact_result.ArtifactResult,
	content *string,
) *RawResultResponse {
	resultResp := &RawResultResponse{
		Id:                dbArtifactResult.Id,
		ContentPath:       dbArtifactResult.ContentPath,
		ContentSerialized: content,
	}

	if !dbArtifactResult.ExecState.IsNull {
		// make a copy of execState's value
		execStateVal := dbArtifactResult.ExecState.ExecutionState
		resultResp.ExecState = &execStateVal
	}

	return resultResp
}

func NewResultResponseFromDbObjects(
	dbArtifact *artifact.DBArtifact,
	dbArtifactResult *artifact_result.ArtifactResult,
	content *string,
	from uuid.UUID,
	to []uuid.UUID,
) *ResultResponse {
	metadata := Response{
		Id:          dbArtifact.Id,
		Name:        dbArtifact.Name,
		Description: dbArtifact.Description,
		Type:        dbArtifact.Type,
		From:        from,
		To:          to,
	}

	if dbArtifactResult == nil {
		return &ResultResponse{Response: metadata}
	}

	resultResp := &RawResultResponse{
		Id:                dbArtifactResult.Id,
		ContentPath:       dbArtifactResult.ContentPath,
		ContentSerialized: content,
	}

	if !dbArtifactResult.ExecState.IsNull {
		// make a copy of execState's value
		execStateVal := dbArtifactResult.ExecState.ExecutionState
		resultResp.ExecState = &execStateVal
	}

	return &ResultResponse{
		Response: metadata,
		Result:   resultResp,
	}
}
