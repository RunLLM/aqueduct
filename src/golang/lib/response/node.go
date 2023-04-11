package response

import (
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// This file should map exactly to
// `src/ui/common/src/handlers/responses/node.ts`
type Artifact struct {
	ID          uuid.UUID           `json:"id"`
	DagID       uuid.UUID           `json:"dag_id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Type        shared.ArtifactType `json:"type"`
	// Once we clean up DBArtifact we should include inputs / outputs fields here.

	// upstream operator ID. It should be unique in practice, but we do not
	// enforce the assumption here.
	Inputs []uuid.UUID `json:"inputs"`

	// downstream operator IDs, could be multiple or empty.
	Outputs []uuid.UUID `json:"outputs"`
}

func NewArtifactFromDBObjet(dbArtifactNode *views.ArtifactNode) *Artifact {
	return &Artifact{
		ID:          dbArtifactNode.ID,
		DagID:       dbArtifactNode.DagID,
		Name:        dbArtifactNode.Name,
		Description: dbArtifactNode.Description,
		Type:        dbArtifactNode.Type,
		Inputs:      dbArtifactNode.Inputs,
		Outputs:     dbArtifactNode.Outputs,
	}
}

type ArtifactResult struct {
	// Contains only the `result`. It mostly mirrors 'artifact_result' schema.
	ID                uuid.UUID                        `json:"id"`
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
	ID          uuid.UUID      `json:"id"`
	DagID       uuid.UUID      `json:"dag_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Spec        *operator.Spec `json:"spec"`

	Inputs  []uuid.UUID `json:"inputs"`
	Outputs []uuid.UUID `json:"outputs"`
}

func NewOperatorFromDBObject(dbOperatorNode *views.OperatorNode) *Operator {
	return &Operator{
		ID:          dbOperatorNode.ID,
		DagID:       dbOperatorNode.DagID,
		Name:        dbOperatorNode.Name,
		Description: dbOperatorNode.Description,
		Spec:        &dbOperatorNode.Spec,
		Inputs:      dbOperatorNode.Inputs,
		Outputs:     dbOperatorNode.Outputs,
	}
}

type OperatorResult struct {
	// Contains only the `result`. It mostly mirrors 'operator_result' schema.
	ID        uuid.UUID              `json:"id"`
	ExecState *shared.ExecutionState `json:"exec_state"`
}

type Nodes struct {
	Operators []Operator `json:"operators"`
	Artifacts []Artifact `json:"artifacts`
}

func NewNodesFromDBObjects(
	operatorNodes []views.OperatorNode,
	artifactNodes []views.ArtifactNode,
) *Nodes {
	operators := make([]Operator, 0, len(operatorNodes))
	artifacts := make([]Artifact, 0, len(artifactNodes))
	for _, opNode := range operatorNodes {
		operators = append(operators, *NewOperatorFromDBObject(&opNode))
	}

	for _, artfNode := range artifactNodes {
		artifacts = append(artifacts, *NewArtifactFromDBObjet(&artfNode))
	}
	return &Nodes{
		Operators: operators,
		Artifacts: artifacts,
	}
}
