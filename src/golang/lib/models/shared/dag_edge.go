package shared

type DAGEdgeType string

const (
	OperatorToArtifactDAGEdge DAGEdgeType = "operator_to_artifact"
	ArtifactToOperatorDAGEdge DAGEdgeType = "artifact_to_operator"
)
