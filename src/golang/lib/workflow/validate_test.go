package workflow

import (
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCheckForCycle(t *testing.T) {
	// Generate 4 workflows that have the following cascading dependencies
	// A --> B --> D
	// |
	// | --> C
	workflowA := models.Workflow{ID: uuid.New()}
	workflowB := models.Workflow{ID: uuid.New()}
	workflowC := models.Workflow{ID: uuid.New()}
	workflowD := models.Workflow{ID: uuid.New()}

	workflowB.Schedule.SourceID = workflowA.ID
	workflowC.Schedule.SourceID = workflowA.ID
	workflowD.Schedule.SourceID = workflowB.ID

	targetWorkflows := []models.Workflow{workflowA, workflowB, workflowC, workflowD}

	type test struct {
		workflowID uuid.UUID
		sourceID   uuid.UUID
		formsCycle bool
	}

	tests := []test{
		{workflowID: workflowC.ID, sourceID: workflowB.ID, formsCycle: false},
		{workflowID: workflowC.ID, sourceID: workflowD.ID, formsCycle: false},
		{workflowID: workflowB.ID, sourceID: workflowC.ID, formsCycle: false},
		{workflowID: workflowB.ID, sourceID: workflowD.ID, formsCycle: true},
		{workflowID: workflowD.ID, sourceID: workflowA.ID, formsCycle: false},
		{workflowID: workflowD.ID, sourceID: workflowC.ID, formsCycle: false},
		{workflowID: workflowA.ID, sourceID: workflowC.ID, formsCycle: true},
		{workflowID: workflowA.ID, sourceID: workflowB.ID, formsCycle: true},
		{workflowID: workflowA.ID, sourceID: workflowD.ID, formsCycle: true},
	}

	for _, tc := range tests {
		require.Equal(
			t,
			tc.formsCycle,
			checkForCycle(tc.workflowID, tc.sourceID, targetWorkflows),
		)
	}
}
