package tests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/stretchr/testify/require"
)

func requireDeepEqual(t *testing.T, expected, actual interface{}) {
	require.True(
		t,
		reflect.DeepEqual(
			expected,
			actual,
		),
		fmt.Sprintf("Expected: %v\n Actual: %v", expected, actual),
	)
}

// requireDeepEqualWorkflows asserts that the expected and actual lists of Workflows
// contain the same elements.
func requireDeepEqualWorkflows(t *testing.T, expected, actual []models.Workflow) {
	require.Equal(t, len(expected), len(actual))

	for _, expectedWorkflow := range expected {
		found := false
		var foundWorkflow models.Workflow

		for _, actualWorkflow := range actual {
			if expectedWorkflow.ID == actualWorkflow.ID {
				found = true
				foundWorkflow = actualWorkflow
				break
			}
		}

		require.True(t, found, "Unable to find workflow: %v", expectedWorkflow)
		requireDeepEqual(t, expectedWorkflow, foundWorkflow)
	}
}

// requireDeepEqualIntegration asserts that the expected and actual lists of Integrations
// contain the same elements.
func requireDeepEqualIntegrations(t *testing.T, expected, actual []models.Integration) {
	require.Equal(t, len(expected), len(actual))

	for _, expectedIntegration := range expected {
		found := false
		var foundIntegration models.Integration

		for _, actualIntegration := range actual {
			if expectedIntegration.ID == actualIntegration.ID {
				found = true
				foundIntegration = actualIntegration
			}
		}
		require.True(t, found, "Unable to find integration: %v", expectedIntegration)
		requireDeepEqual(t, expectedIntegration, foundIntegration)
	}
}

// requireDeepEqualDAGs asserts that the expected and actual lists of DAGs
// containt the same elements.
func requireDeepEqualDAGs(t *testing.T, expected, actual []models.DAG) {
	require.Equal(t, len(expected), len(actual))

	for _, expectedDAG := range expected {
		found := false
		var foundDAG models.DAG

		for _, actualDAG := range actual {
			if expectedDAG.ID == actualDAG.ID {
				found = true
				foundDAG = actualDAG
				break
			}
		}
		require.True(t, found, "Unable to find DAG: %v", expectedDAG)
		requireDeepEqual(t, expectedDAG, foundDAG)
	}
}

// requireDeepEqualArtifact asserts that the expected and actual lists of Artifacts
// containt the same elements.
func requireDeepEqualArtifacts(t *testing.T, expected, actual []models.Artifact) {
	require.Equal(t, len(expected), len(actual))

	for _, expectedArtifact := range expected {
		found := false
		var foundArtifact models.Artifact

		for _, actualArtifact := range actual {
			if expectedArtifact.ID == actualArtifact.ID {
				found = true
				foundArtifact = actualArtifact
				break
			}
		}
		require.True(t, found, "Unable to find Artifact: %v", expectedArtifact)
		requireDeepEqual(t, expectedArtifact, foundArtifact)
	}
}

// requireDeepEqualDAGResults asserts that the expected and actual lists
// of DAGResults containt the same elements.
func requireDeepEqualDAGResults(t *testing.T, expected, actual []models.DAGResult) {
	require.Equal(t, len(expected), len(actual))

	for _, expectedDAGResult := range expected {
		found := false
		var foundDAGResult models.DAGResult

		for _, actualDAGResult := range actual {
			if expectedDAGResult.ID == actualDAGResult.ID {
				found = true
				foundDAGResult = actualDAGResult
				break
			}
		}

		require.True(t, found, "Unable to find DAGResult: %v", expectedDAGResult)
		requireDeepEqual(t, expectedDAGResult, foundDAGResult)
	}
}

// requireDeepEqualDAGEdges asserts that the expected and actual lists
// of DAGEdges containt the same elements.
func requireDeepEqualDAGEdges(t *testing.T, expected, actual []models.DAGEdge) {
	require.Equal(t, len(expected), len(actual))

	for _, expectedDAGEdge := range expected {
		found := false
		var foundDAGEdge models.DAGEdge

		for _, actualDAGEdge := range actual {
			if expectedDAGEdge.DagID == actualDAGEdge.DagID &&
				expectedDAGEdge.FromID == actualDAGEdge.FromID &&
				expectedDAGEdge.ToID == actualDAGEdge.ToID &&
				expectedDAGEdge.Idx == actualDAGEdge.Idx {
				found = true
				foundDAGEdge = actualDAGEdge
				break
			}
		}

		require.True(t, found, "Unable to find DAGEdge: %v", expectedDAGEdge)
		requireDeepEqual(t, expectedDAGEdge, foundDAGEdge)
	}
}
