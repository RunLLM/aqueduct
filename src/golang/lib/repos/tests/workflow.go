package tests

import "github.com/stretchr/testify/require"

func (ts *TestSuite) TestWorkflow_Exists() {
	// TODO: Fix test once user refactor is complete
	workflows := ts.seedWorkflow(1)
	require.Nil(ts.T(), workflows)
}
