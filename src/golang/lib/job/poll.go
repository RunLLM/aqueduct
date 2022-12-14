package job

import (
	"context"
	"time"


	"github.com/dropbox/godropbox/errors"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
)

// PollJob waits for the specified job to finish and returns its status.
// If a timeout is reached, or it is unable to check the job status, it returns an error.
func PollJob(
	ctx context.Context,
	name string,
	manager JobManager,
	pollInterval time.Duration,
	pollTimeout time.Duration,
) (shared.ExecutionStatus, error) {
	poller := time.NewTicker(pollInterval)
	timeout := time.NewTimer(pollTimeout)

	for {
		select {
		case <-poller.C:
			status, err := manager.Poll(ctx, name)
			if err != nil {
				return shared.UnknownExecutionStatus, err
			}

			if status == shared.SucceededExecutionStatus ||
				status == shared.FailedExecutionStatus {
				return status, nil
			}
		case <-timeout.C:
			return shared.UnknownExecutionStatus, errors.Newf("Reached timeout waiting for the job %s to finish.", name)
		}
	}
}
