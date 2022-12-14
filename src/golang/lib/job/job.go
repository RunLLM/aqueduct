package job

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/dropbox/godropbox/errors"
)

var ErrAsyncExecution = errors.New("Unknown job status due to asynchronous execution.")

type JobManager interface {
	Config() Config
	Launch(ctx context.Context, name string, spec Spec) JobError
	Poll(ctx context.Context, name string) (shared.ExecutionStatus, JobError)
	DeployCronJob(ctx context.Context, name string, period string, spec Spec) JobError
	CronJobExists(ctx context.Context, name string) bool
	EditCronJob(ctx context.Context, name string, cronString string) JobError
	DeleteCronJob(ctx context.Context, name string) JobError
}

func NewJobManager(conf Config) (JobManager, error) {
	if conf.Type() == ProcessType {
		processConfig, ok := conf.(*ProcessConfig)
		if !ok {
			return nil, errors.New("JobManager config is not of type Process.")
		}
		return NewProcessJobManager(processConfig)
	}
	if conf.Type() == K8sType {
		k8sConfig, ok := conf.(*K8sJobManagerConfig)
		if !ok {
			return nil, errors.New("JobManager config is not of type K8s.")
		}
		return NewK8sJobManager(k8sConfig)
	}
	if conf.Type() == LambdaType {
		lambdaConfig, ok := conf.(*LambdaJobManagerConfig)
		if !ok {
			return nil, errors.New("JobManager config is not of type Lambda.")
		}
		return NewLambdaJobManager(lambdaConfig)
	}

	return nil, errors.Newf("JobManager config is of unsupported type %s", conf.Type())
}
