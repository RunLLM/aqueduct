package shared

type FailureType int64

const (
	Success          FailureType = 0
	SystemFailure    FailureType = 1
	UserFatalFailure FailureType = 2

	// Orchestration can continue onwards, despite this failure.
	// Eg. Check operator with WARNING severity does not pass.
	UserNonFatalFailure FailureType = 3
)
