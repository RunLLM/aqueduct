package response

import (
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// This file should map exactly to
// `src/ui/common/src/handlers/responses/node.ts`
type MergedNode struct {
	ID          uuid.UUID           `json:"id"`
	DagID       uuid.UUID           `json:"dag_id"`
	ArtifactID  uuid.UUID           `json:"artifact_id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Spec        *operator.Spec      `json:"spec"`
	Type        shared.ArtifactType `json:"type"`

	// Upstream artifact ID, could be multiple or empty.
	Inputs []uuid.UUID `json:"inputs"`

	// Downstream operator IDs, could be multiple or empty.
	Outputs []uuid.UUID `json:"outputs"`
}

func NewMergedNodeFromDBObject(dbMergedNode *views.MergedNode) *MergedNode {
	return &MergedNode{
		ID:          dbMergedNode.ID,
		DagID:       dbMergedNode.DagID,
		ArtifactID:  dbMergedNode.ArtifactID,
		Name:        dbMergedNode.Name,
		Description: dbMergedNode.Description,
		Spec:        &dbMergedNode.Spec,
		Type:        dbMergedNode.Type,
		// Inputs to the metric operator
		Inputs: dbMergedNode.Inputs,
		// Outputs of the metric artifact
		Outputs: dbMergedNode.Outputs,
	}
}

type MergedNodeResult struct {
	// Operator ID
	ID                uuid.UUID              `json:"id"`
	OperatorExecState *shared.ExecutionState `json:"operator_exec_state"`

	ArtifactID        uuid.UUID                        `json:"artifact_id"`
	SerializationType shared.ArtifactSerializationType `json:"serialization_type"`

	// If `ContentSerialized` is set, the content is small and we directly send
	// it as a part of response. It's consistent with the object stored in `ContentPath`.
	// The value is the string representation of the file stored in that path.
	//
	// Otherwise, the content is large and
	// one should send an additional request to fetch the content.
	ContentPath       string  `json:"content_path"`
	ContentSerialized *string `json:"content_serialized"`

	ArtifactExecState *shared.ExecutionState `json:"artifact_exec_state"`
}

func NewMergedNodeResultFromDBObject(
	dbMergedNodeResult *views.MergedNodeResult,
	content *string,
) *MergedNodeResult {
	result := &MergedNodeResult{
		ID:                dbMergedNodeResult.ID,
		ArtifactID:        dbMergedNodeResult.ArtifactID,
		SerializationType: dbMergedNodeResult.Metadata.SerializationType,
		ContentPath:       dbMergedNodeResult.ContentPath,
		ContentSerialized: content,
	}

	if !dbMergedNodeResult.OperatorExecState.IsNull {
		// make a copy of execState's value
		execStateVal := dbMergedNodeResult.OperatorExecState.ExecutionState
		result.OperatorExecState = &execStateVal
	}

	if !dbMergedNodeResult.ArtifactExecState.IsNull {
		// make a copy of execState's value
		execStateVal := dbMergedNodeResult.ArtifactExecState.ExecutionState
		result.ArtifactExecState = &execStateVal
	}

	return result
}

type Artifact struct {
	ID          uuid.UUID           `json:"id"`
	DagID       uuid.UUID           `json:"dag_id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Type        shared.ArtifactType `json:"type"`
	// Once we clean up DBArtifact we should include inputs / outputs fields here.

	// Upstream operator ID.
	Input uuid.UUID `json:"input"`

	// Downstream operator IDs, could be multiple or empty.
	Outputs []uuid.UUID `json:"outputs"`
}

func NewArtifactFromDBObject(dbArtifactNode *views.ArtifactNode) *Artifact {
	return &Artifact{
		ID:          dbArtifactNode.ID,
		DagID:       dbArtifactNode.DagID,
		Name:        dbArtifactNode.Name,
		Description: dbArtifactNode.Description,
		Type:        dbArtifactNode.Type,
		Input:       dbArtifactNode.Input,
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

func NewArtifactResultFromDBObject(
	dbArtifactResult *models.ArtifactResult,
	content *string,
) *ArtifactResult {
	result := &ArtifactResult{
		ID:                dbArtifactResult.ID,
		SerializationType: dbArtifactResult.Metadata.SerializationType,
		ContentPath:       dbArtifactResult.ContentPath,
		ContentSerialized: content,
	}

	if !dbArtifactResult.ExecState.IsNull {
		// make a copy of execState's value
		execStateVal := dbArtifactResult.ExecState.ExecutionState
		result.ExecState = &execStateVal
	}

	return result
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

func NewOperatorResultFromDBObject(
	dbOperatorResult *models.OperatorResult,
) *OperatorResult {
	result := &OperatorResult{ID: dbOperatorResult.ID}
	if !dbOperatorResult.ExecState.IsNull {
		// make a copy of execState's value
		execStateVal := dbOperatorResult.ExecState.ExecutionState
		result.ExecState = &execStateVal
	}

	return result
}

type Nodes struct {
	Operators []Operator `json:"operators"`
	Artifacts []Artifact `json:"artifacts"`
	// Metrics []MergedNode `json:"metrics"`
	// Checks []MergedNode `json:"checks"`
}

func NewNodesFromDBObjects(
	operatorNodes []views.OperatorNode,
	artifactNodes []views.ArtifactNode,
	// metricNodes []views.MergedNode,
	// checkNodes []views.MergedNode,
) *Nodes {
	return &Nodes{
		Operators: slices.Map(
			operatorNodes,
			func(node views.OperatorNode) Operator {
				return *NewOperatorFromDBObject(&node)
			},
		),
		Artifacts: slices.Map(
			artifactNodes,
			func(node views.ArtifactNode) Artifact {
				return *NewArtifactFromDBObject(&node)
			},
		),
		// Metrics: slices.Map(
		// 	metricNodes,
		// 	func(node views.MergedNode) MergedNode {
		// 		return *NewMergedNodeFromDBObject(&node)
		// 	},
		// ),
		// Checks: slices.Map(
		// 	checkNodes,
		// 	func(node views.MergedNode) MergedNode {
		// 		return *NewMergedNodeFromDBObject(&node)
		// 	},
		// ),
	}
}

type NodeResults struct {
	Operators []OperatorResult `json:"operators"`
	Artifacts []ArtifactResult `json:"artifacts"`
	// Metrics []MergedNodeResult `json:"metrics"`
	// Checks []MergedNodeResult `json:"checks"`
}

func NewNodeResultsFromDBObjects(
	dbOperatorResults []models.OperatorResult,
	dbArtifactResults []models.ArtifactResult,
	contents map[string]string,
) *NodeResults {
	return &NodeResults{
		Operators: slices.Map(
			dbOperatorResults,
			func(result models.OperatorResult) OperatorResult {
				return *NewOperatorResultFromDBObject(&result)
			},
		),
		Artifacts: slices.Map(
			dbArtifactResults,
			func(result models.ArtifactResult) ArtifactResult {
				content, ok := contents[result.ContentPath]
				var contentPtr *string
				if ok {
					contentPtr = &content
				}

				return *NewArtifactResultFromDBObject(&result, contentPtr)
			},
		),
	}
}

// Node content represents the content of the requested node.
// It's currently used in two cases:
// * operator: NodeContent is the .zip file of the operator. `Name`
// is the file name and `Data` is the file bytes.
// * artifact result: NodeContent is the bytes data stored in content_path
// in storage. The exact format depends on the artifact result's `SerializationType`
// and is up to the caller to process. The `Name` field is just the artifact name and
// is not particularly useful.
type NodeContent struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}
