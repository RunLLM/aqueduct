package cronjob

import (
	"context"
)

type CronjobManager interface {
	DeployCronJob(ctx context.Context, name string, period string, cronFunction func()) error
	CronJobExists(ctx context.Context, name string) bool
	EditCronJob(ctx context.Context, name string, cronString string, cronFunction func()) error
	DeleteCronJob(ctx context.Context, name string) error
}
