package cronjob

import (
	"context"
	"testing"

	"github.com/go-co-op/gocron"
	"github.com/stretchr/testify/require"
)

func generateDummyFunction() func() {
	return func() {}
}

func TestDeployCronJob(t *testing.T) {
	cronjobManager := NewProcessCronjobManager()

	ctx := context.Background()

	workflowName := "workflow"
	cronString := "0 * * * *"

	// Deploy an unpaused workflow.
	err := cronjobManager.DeployCronJob(ctx, workflowName, cronString, generateDummyFunction())
	require.Nil(t, err)
	require.Equal(t, 1, len(cronjobManager.cronMapping))
	require.NotEqual(t, (*gocron.Job)(nil), cronjobManager.cronMapping[workflowName].cronJob)
	require.Equal(t, 1, len(cronjobManager.cronScheduler.Jobs()))

	// Deploy a paused workflow.
	pausedWorkflowName := "paused_workflow"
	emptyCronString := ""
	err = cronjobManager.DeployCronJob(ctx, pausedWorkflowName, emptyCronString, generateDummyFunction())
	require.Nil(t, err)
	require.Equal(t, 2, len(cronjobManager.cronMapping))
	require.Equal(t, (*gocron.Job)(nil), cronjobManager.cronMapping[pausedWorkflowName].cronJob)
	require.Equal(t, 1, len(cronjobManager.cronScheduler.Jobs()))
}

func TestEditCronJob(t *testing.T) {
	cronjobManager := NewProcessCronjobManager()

	ctx := context.Background()

	workflowName := "workflow"
	pausedWorkflowName := "paused_workflow"
	cronString := "0 * * * *"
	newCronString := "1 * * * *"
	emptyCronString := ""

	cronjobManager.DeployCronJob(ctx, workflowName, cronString, generateDummyFunction())
	cronjobManager.DeployCronJob(ctx, pausedWorkflowName, emptyCronString, generateDummyFunction())

	// Edit an unpaused workflow to another schedule.
	err := cronjobManager.EditCronJob(ctx, workflowName, newCronString, generateDummyFunction())
	require.Nil(t, err)
	require.Equal(t, 2, len(cronjobManager.cronMapping))
	require.NotEqual(t, (*gocron.Job)(nil), cronjobManager.cronMapping[workflowName].cronJob)
	require.Equal(t, 1, len(cronjobManager.cronScheduler.Jobs()))

	// Edit an unpaused workflow to paused.
	err = cronjobManager.EditCronJob(ctx, workflowName, emptyCronString, generateDummyFunction())
	require.Nil(t, err)
	require.Equal(t, 2, len(cronjobManager.cronMapping))
	require.Equal(t, (*gocron.Job)(nil), cronjobManager.cronMapping[workflowName].cronJob)
	require.Equal(t, 0, len(cronjobManager.cronScheduler.Jobs()))

	// Edit a paused workflow to unpaused.
	err = cronjobManager.EditCronJob(ctx, pausedWorkflowName, cronString, generateDummyFunction())
	require.Nil(t, err)
	require.Equal(t, 2, len(cronjobManager.cronMapping))
	require.NotEqual(t, (*gocron.Job)(nil), cronjobManager.cronMapping[pausedWorkflowName].cronJob)
	require.Equal(t, 1, len(cronjobManager.cronScheduler.Jobs()))
}

func TestDeleteCronJob(t *testing.T) {
	cronjobManager := NewProcessCronjobManager()

	ctx := context.Background()

	workflowName := "workflow"
	cronString := "0 * * * *"

	cronjobManager.DeployCronJob(ctx, workflowName, cronString, generateDummyFunction())
	err := cronjobManager.DeleteCronJob(ctx, workflowName)
	require.Nil(t, err)
	require.Equal(t, 0, len(cronjobManager.cronMapping))
	require.Equal(t, 0, len(cronjobManager.cronScheduler.Jobs()))
}
