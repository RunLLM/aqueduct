package job

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/dropbox/godropbox/errors"
)

var (
	ErrInvalidJobManagerConfig = errors.New("Job manager config is not valid.")
	ErrJobNotExist             = errors.New("Job does not exist.")
	ErrJobAlreadyExists        = errors.New("Job already exists.")
	ErrAsyncExecution          = errors.New("Unknown job status due to asynchronous execution.")
	ErrPollJobTimeout          = errors.New("Reached timeout waiting for the job to finish.")
)

// These error codes come from our JobManagers when they fail to properly guide
// their a job through its proper lifecycle. Errors surfaced this way are propagated
// outside the python executor context. Their meaning is consistent across all
// types of JobManagers.
type JobErrorCode int

const (
	// Indicates an unknown system issue that we cannot recover from.
	System JobErrorCode = 0

	// Indicates that the issue was the user's fault, and to surface the error message
	// to the user.
	User = 1
)

type JobError struct {
	errors.DropboxError
	Code JobErrorCode
}

func wrapInJobError(code JobErrorCode, err error) error {
	if dropboxErr, ok := err.(errors.DropboxError); ok {
		return &JobError{
			DropboxError: dropboxErr,
			Code:         code,
		}
	}

	return &JobError{
		DropboxError: errors.New(err.Error()),
		Code:         code,
	}
}

type JobManager interface {
	Config() Config
	Launch(ctx context.Context, name string, spec Spec) error
	Poll(ctx context.Context, name string) (shared.ExecutionStatus, error)
	DeployCronJob(ctx context.Context, name string, period string, spec Spec) error
	CronJobExists(ctx context.Context, name string) bool
	EditCronJob(ctx context.Context, name string, cronString string) error
	DeleteCronJob(ctx context.Context, name string) error
}

func NewJobManager(conf Config) (JobManager, error) {
	if conf.Type() == ProcessType {
		processConfig, ok := conf.(*ProcessConfig)
		if !ok {
			return nil, ErrInvalidJobManagerConfig
		}
		return NewProcessJobManager(processConfig)
	}
	if conf.Type() == K8sType {
		k8sConfig, ok := conf.(*K8sJobManagerConfig)
		if !ok {
			return nil, ErrInvalidJobManagerConfig
		}
		return NewK8sJobManager(k8sConfig)
	}
	if conf.Type() == LambdaType {
		lambdaConfig, ok := conf.(*LambdaJobManagerConfig)
		if !ok {
			return nil, ErrInvalidJobManagerConfig
		}
		return NewLambdaJobManager(lambdaConfig)
	}
	if conf.Type() == DatabricksType {
		databricksConfig, ok := conf.(*DatabricksJobManagerConfig)
		if !ok {
			return nil, ErrInvalidJobManagerConfig
		}
		return NewDatabricksJobManager(databricksConfig)
	}

	return nil, ErrInvalidJobManagerConfig
}
