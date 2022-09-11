package job

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/dropbox/godropbox/errors"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidJobManagerConfig = errors.New("Job manager config is not valid.")
	ErrJobNotExist             = errors.New("Job does not exist.")
	ErrJobAlreadyExists        = errors.New("Job already exists.")
	ErrPollJobTimeout          = errors.New("Reached timeout waiting for the job to finish.")
)

type JobManager interface {
	Config() Config
	Launch(ctx context.Context, name string, spec Spec) error
	Poll(ctx context.Context, name string, metadataPath string, storageConfig *shared.StorageConfig) (shared.ExecutionStatus, error)
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
		logrus.Info("Creating a LambdaJobManager")
		return NewLambdaJobManager(lambdaConfig)
	}

	return nil, ErrInvalidJobManagerConfig
}
