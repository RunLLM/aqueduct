package workflow_dag_edge

type Type string

const (
	OperatorToArtifactType Type = "operator_to_artifact"
	ArtifactToOperatorType Type = "artifact_to_operator"
)
