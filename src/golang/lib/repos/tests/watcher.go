package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestWatcher_Create() {
	workflows := ts.seedWorkflow(1)
	workflow := workflows[0]

	expectedWatcher := &models.Watcher{
		WorkflowID: workflow.ID,
		UserID:     workflow.UserID,
	}

	actualWatcher, err := ts.watcher.Create(
		ts.ctx,
		expectedWatcher.WorkflowID,
		expectedWatcher.UserID,
		ts.DB,
	)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedWatcher, actualWatcher)
}

func (ts *TestSuite) TestWatcher_Delete() {
	watcher := ts.seedWatcher()

	err := ts.watcher.Delete(ts.ctx, watcher.WorkflowID, watcher.UserID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestWatcher_DeleteByWorkflow() {
	watcher := ts.seedWatcher()

	err := ts.watcher.DeleteByWorkflow(ts.ctx, watcher.WorkflowID, ts.DB)
	require.Nil(ts.T(), err)
}
