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

// requireDeepEqualArtifacts asserts that the expected and actual lists
// of Artifact containt the same elements.
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
