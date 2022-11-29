package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestDAGEdge_GetArtifactToOperatorByDAG() {
	edges := ts.seedDAGEdge(5)
	dagID := edges[0].DagID

	var expectedEdges []models.DAGEdge
	for _, edge := range edges {
		if edge.Type == shared.ArtifactToOperatorDAGEdge {
			expectedEdges = append(expectedEdges, edge)
		}
	}

	actualEdges, err := ts.dagEdge.GetArtifactToOperatorByDAG(ts.ctx, dagID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualDAGEdges(ts.T(), expectedEdges, actualEdges)
}

func (ts *TestSuite) TestDAGEdge_GetByDAGBatch() {
	dags := ts.seedDAG(2)
	dagA, dagB := dags[0], dags[1]

	edgesA := ts.seedDAGEdgeWithDAG(5, dagA.ID)
	edgesB := ts.seedDAGEdgeWithDAG(5, dagB.ID)

	expectedEdges := edgesA
	expectedEdges = append(expectedEdges, edgesB...)

	actualEdges, err := ts.dagEdge.GetByDAGBatch(ts.ctx, []uuid.UUID{dagA.ID, dagB.ID}, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualDAGEdges(ts.T(), expectedEdges, actualEdges)
}

func (ts *TestSuite) TestDAGEdge_GetOperatorToArtifactByDAG() {
	edges := ts.seedDAGEdge(5)
	dagID := edges[0].DagID

	var expectedEdges []models.DAGEdge
	for _, edge := range edges {
		if edge.Type == shared.OperatorToArtifactDAGEdge {
			expectedEdges = append(expectedEdges, edge)
		}
	}

	actualEdges, err := ts.dagEdge.GetOperatorToArtifactByDAG(ts.ctx, dagID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualDAGEdges(ts.T(), expectedEdges, actualEdges)
}

func (ts *TestSuite) TestDAGEdge_Create() {
	dags := ts.seedDAG(1)
	dag := dags[0]

	expectedEdge := &models.DAGEdge{
		DagID:  dag.ID,
		Type:   shared.ArtifactToOperatorDAGEdge,
		FromID: uuid.New(),
		ToID:   uuid.New(),
		Idx:    0,
	}

	actualEdge, err := ts.dagEdge.Create(
		ts.ctx,
		expectedEdge.DagID,
		expectedEdge.Type,
		expectedEdge.FromID,
		expectedEdge.ToID,
		expectedEdge.Idx,
		ts.DB,
	)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedEdge, actualEdge)
}

func (ts *TestSuite) TestDAGEdge_DeleteByDAGBatch() {
	dags := ts.seedDAG(2)
	dagA, dagB := dags[0], dags[1]

	ts.seedDAGEdgeWithDAG(5, dagA.ID)
	ts.seedDAGEdgeWithDAG(5, dagB.ID)

	err := ts.dagEdge.DeleteByDAGBatch(ts.ctx, []uuid.UUID{dagA.ID, dagB.ID}, ts.DB)
	require.Nil(ts.T(), err)
}
