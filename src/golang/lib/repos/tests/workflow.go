package tests

func (ts *TestSuite) TestWorkflow_Exists() {
	// TODO: Fix test once user refactor is complete
	workflows := ts.seedWorkflow(1)
	ts.Require().Nil(workflows)
}
