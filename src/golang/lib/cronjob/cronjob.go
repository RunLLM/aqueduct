package cronjob

import (
	"context"

	"github.com/dropbox/godropbox/errors"
)

var (
	ErrInvalidJobManagerConfig = errors.New("Job manager config is not valid.")
	ErrJobNotExist             = errors.New("Job does not exist.")
	ErrJobAlreadyExists        = errors.New("Job already exists.")
	ErrPollJobTimeout          = errors.New("Reached timeout waiting for the job to finish.")
)

type CronjobManager interface {
	DeployCronJob(ctx context.Context, name string, period string, cronFunction func()) error
	CronJobExists(ctx context.Context, name string) bool
	EditCronJob(ctx context.Context, name string, cronString string, cronFunction func()) error
	DeleteCronJob(ctx context.Context, name string) error
}
