package job

import "github.com/aqueducthq/aqueduct/lib/errors"

// JobErrorCode come from our JobManagers when they fail to properly guide
// their a job through its proper lifecycle. Errors surfaced this way are propagated
// outside of the python executor context. Their meaning is consistent across all
// types of JobManagers.
type JobErrorCode int

const (
	// System indicates an unexpected system issue that we cannot recover from.
	System JobErrorCode = iota

	// User error code indicates that the issue was the user's fault, and to surface that message
	// to the user.
	User

	// JobMissing indicates that the job manager could not find the specified job.
	JobMissing

	// Noop indicates that the job manager does not have the context to make a definitive claim
	// as to whether the job has succeeded or not. This is equivalent to saying "Do not ask me, I don't know.".
	//
	// For example, polling a Lambda JobManager will return this error because there is no concept of
	// fetching a specific job's information from Lambda. The caller must figure out the job status
	// through other means.
	Noop
)

type JobError interface {
	errors.AqError

	Code() JobErrorCode
}

type jobErrorImpl struct {
	errors.AqError
	code JobErrorCode
}

func (je *jobErrorImpl) Code() JobErrorCode {
	return je.code
}

func wrapInJobError(code JobErrorCode, err error) JobError {
	if aqErr, ok := err.(errors.AqError); ok {
		return &jobErrorImpl{
			AqError: aqErr,
			code:    code,
		}
	}

	return &jobErrorImpl{
		AqError: errors.New(err.Error()),
		code:    code,
	}
}

func systemError(err error) JobError {
	return wrapInJobError(System, err)
}

func userError(err error) JobError {
	return wrapInJobError(User, err)
}

func jobMissingError(err error) JobError {
	return wrapInJobError(JobMissing, err)
}

func noopError(err error) JobError {
	return wrapInJobError(Noop, err)
}
