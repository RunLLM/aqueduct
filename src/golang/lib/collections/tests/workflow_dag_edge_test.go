package tests

import (
	"context"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// seedWorkflowDagEdgeWithDagId populates the workflow_dag_edge table with the edges from
// edgesMap for the workflow dag specified.
func seedWorkflowDagEdgeWithDagId(t *testing.T, edgesMap map[uuid.UUID]uuid.UUID, workflowDagId uuid.UUID) []workflow_dag_edge.WorkflowDagEdge {
	edges := make([]workflow_dag_edge.WorkflowDagEdge, 0, len(edgesMap))

	for fromId, toId := range edgesMap {
		testEdge, err := writers.workflowDagEdgeWriter.CreateWorkflowDagEdge(
			context.Background(),
			workflowDagId,
			workflow_dag_edge.OperatorToArtifactType,
			fromId,
			toId,
			0,
			db,
		)
		require.Nil(t, err)

		edges = append(edges, *testEdge)
	}

	require.Len(t, edges, len(edgesMap))

	return edges
}
