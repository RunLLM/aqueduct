package artifact

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

type Response struct {
	Id          uuid.UUID           `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Type        shared.ArtifactType `json:"type"`
	// Once we clean up DBArtifact we should include inputs / outputs fields here.

	// upstream operator ID, must be unique.
	From uuid.UUID `json:"from"`

	// downstream operator IDs, could be multiple or empty.
	To []uuid.UUID `json:"to"`
}

type RawResultResponse struct {
	// Contains only the `result`. It mostly mirrors 'artifact_result' schema.
	Id                uuid.UUID                        `json:"id"`
	SerializationType shared.ArtifactSerializationType `json:"serialization_type"`

	// If `ContentSerialized` is set, the content is small and we directly send
	// it as a part of response. It's consistent with the object stored in `ContentPath`.
	// The value is the string representation of the file stored in that path.
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

func NewResponseFromDbObject(
	dbArtifact *models.Artifact,
	from uuid.UUID,
	to []uuid.UUID,
) *Response {
	return &Response{
		Id:          dbArtifact.ID,
		Name:        dbArtifact.Name,
		Description: dbArtifact.Description,
		Type:        dbArtifact.Type,
		From:        from,
		To:          to,
	}
}

func NewRawResultResponseFromDbObject(
	dbArtifactResult *models.ArtifactResult,
	content *string,
) *RawResultResponse {
	resultResp := &RawResultResponse{
		Id:                dbArtifactResult.ID,
		SerializationType: dbArtifactResult.Metadata.SerializationType,
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
	dbArtifact *models.Artifact,
	dbArtifactResult *models.ArtifactResult,
	content *string,
	from uuid.UUID,
	to []uuid.UUID,
) *ResultResponse {
	resp := NewResponseFromDbObject(dbArtifact, from, to)

	if dbArtifactResult == nil {
		return &ResultResponse{Response: *resp}
	}

	return &ResultResponse{
		Response: *resp,
		Result:   NewRawResultResponseFromDbObject(dbArtifactResult, content),
	}
}

func NewResultResponseFromDBView(
	dbViewArtfWithResult *views.ArtifactWithResult,
	content *string,
) *ResultResponse {
	return NewResultResponseFromDbObjects(
		&models.Artifact{
			ID:          dbViewArtfWithResult.ID,
			Name:        dbViewArtfWithResult.Name,
			Description: dbViewArtfWithResult.Description,
			Type:        dbViewArtfWithResult.Type,
		},
		&models.ArtifactResult{
			ID:          dbViewArtfWithResult.ResultID,
			DAGResultID: dbViewArtfWithResult.DAGResultID,
			ArtifactID:  dbViewArtfWithResult.ID,
			ContentPath: dbViewArtfWithResult.ContentPath,
			ExecState:   dbViewArtfWithResult.ExecState,
			Metadata:    dbViewArtfWithResult.Metadata,
		},
		content,
		uuid.UUID{}, // from, we ignore this field for now
		nil,         // to, we ignore this field for now
	)
}
