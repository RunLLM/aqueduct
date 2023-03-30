package response

import (
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/google/uuid"
)

// This file should map exactly to
// `src/ui/common/src/handlers/responses/node.ts`
type Artifact struct {
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

type ArtifactResult struct {
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

type Operator struct {
	Id          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Spec        *operator.Spec `json:"spec"`

	Inputs  []uuid.UUID `json:"inputs"`
	Outputs []uuid.UUID `json:"outputs"`
}

type OperatorResult struct {
	// Contains only the `result`. It mostly mirrors 'operator_result' schema.
	Id        uuid.UUID              `json:"id"`
	ExecState *shared.ExecutionState `json:"exec_state"`
}
