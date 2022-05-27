package job

import (
	"context"
	"testing"

	"github.com/go-co-op/gocron"
	"github.com/stretchr/testify/require"
)

var (
	dummyProcessConfig = &ProcessConfig{}
	dummyWorkflowSpec  = &WorkflowSpec{}
)

func TestDeployCronJob(t *testing.T) {
	jobManager, err := NewProcessJobManager(dummyProcessConfig)
	require.Nil(t, err)

	ctx := context.Background()

	workflowName := "workflow"
	cronString := "0 * * * *"

	// Deploy an unpaused workflow.
	err = jobManager.DeployCronJob(ctx, workflowName, cronString, dummyWorkflowSpec)
	require.Nil(t, err)
	require.Equal(t, 1, len(jobManager.cronMapping))
	require.NotEqual(t, (*gocron.Job)(nil), jobManager.cronMapping[workflowName].cronJob)
	require.Equal(t, 1, len(jobManager.cronScheduler.Jobs()))

	// Deploy a paused workflow.
	pausedWorkflowName := "paused_workflow"
	emptyCronString := ""
	err = jobManager.DeployCronJob(ctx, pausedWorkflowName, emptyCronString, dummyWorkflowSpec)
	require.Nil(t, err)
	require.Equal(t, 2, len(jobManager.cronMapping))
	require.Equal(t, (*gocron.Job)(nil), jobManager.cronMapping[pausedWorkflowName].cronJob)
	require.Equal(t, 1, len(jobManager.cronScheduler.Jobs()))
}

func TestEditCronJob(t *testing.T) {
	jobManager, err := NewProcessJobManager(dummyProcessConfig)
	require.Nil(t, err)

	ctx := context.Background()

	workflowName := "workflow"
	pausedWorkflowName := "paused_workflow"
	cronString := "0 * * * *"
	newCronString := "1 * * * *"
	emptyCronString := ""

	jobManager.DeployCronJob(ctx, workflowName, cronString, dummyWorkflowSpec)
	jobManager.DeployCronJob(ctx, pausedWorkflowName, emptyCronString, dummyWorkflowSpec)

	// Edit an unpaused workflow to another schedule.
	err = jobManager.EditCronJob(ctx, workflowName, newCronString)
	require.Nil(t, err)
	require.Equal(t, 2, len(jobManager.cronMapping))
	require.NotEqual(t, (*gocron.Job)(nil), jobManager.cronMapping[workflowName].cronJob)
	require.Equal(t, 1, len(jobManager.cronScheduler.Jobs()))

	// Edit an unpaused workflow to paused.
	err = jobManager.EditCronJob(ctx, workflowName, emptyCronString)
	require.Nil(t, err)
	require.Equal(t, 2, len(jobManager.cronMapping))
	require.Equal(t, (*gocron.Job)(nil), jobManager.cronMapping[workflowName].cronJob)
	require.Equal(t, 0, len(jobManager.cronScheduler.Jobs()))

	// Edit a paused workflow to unpaused.
	err = jobManager.EditCronJob(ctx, pausedWorkflowName, cronString)
	require.Nil(t, err)
	require.Equal(t, 2, len(jobManager.cronMapping))
	require.NotEqual(t, (*gocron.Job)(nil), jobManager.cronMapping[pausedWorkflowName].cronJob)
	require.Equal(t, 1, len(jobManager.cronScheduler.Jobs()))
}

func TestDeleteCronJob(t *testing.T) {
	jobManager, err := NewProcessJobManager(dummyProcessConfig)
	require.Nil(t, err)

	ctx := context.Background()

	workflowName := "workflow"
	cronString := "0 * * * *"

	jobManager.DeployCronJob(ctx, workflowName, cronString, dummyWorkflowSpec)
	err = jobManager.DeleteCronJob(ctx, workflowName)
	require.Nil(t, err)
	require.Equal(t, 0, len(jobManager.cronMapping))
	require.Equal(t, 0, len(jobManager.cronScheduler.Jobs()))
}
