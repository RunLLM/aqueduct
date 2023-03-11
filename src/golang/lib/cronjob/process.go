package cronjob

import (
	"context"
	"sync"
	"time"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/go-co-op/gocron"
)

type cronMetadata struct {
	// If the cronJob is nil, it means the corresponding workflow has been paused.
	cronJob *gocron.Job
}

// Please use thread-safe read / insert / remove APIs to maintain maps.
// These APIs are wrapped with proper locks to support concurrency.
// Never try to access map using go's native APIs.
type ProcessCronjobManager struct {
	cronScheduler *gocron.Scheduler
	// A mapping from cron job name to cron job object pointer.
	cronMapping map[string]*cronMetadata
	cronMutex   *sync.RWMutex
}

func NewProcessCronjobManager() *ProcessCronjobManager {
	cronScheduler := gocron.NewScheduler(time.UTC)
	cronScheduler.StartAsync()

	return &ProcessCronjobManager{
		cronScheduler: cronScheduler,
		cronMapping:   map[string]*cronMetadata{},
		cronMutex:     &sync.RWMutex{},
	}
}

func (j *ProcessCronjobManager) getCronMap(key string) (*cronMetadata, bool) {
	j.cronMutex.RLock()
	cron, ok := j.cronMapping[key]
	j.cronMutex.RUnlock()
	return cron, ok
}

func (j *ProcessCronjobManager) setCronMap(key string, cron *cronMetadata) {
	j.cronMutex.Lock()
	j.cronMapping[key] = cron
	j.cronMutex.Unlock()
}

func (j *ProcessCronjobManager) deleteCronMap(key string) {
	j.cronMutex.Lock()
	delete(j.cronMapping, key)
	j.cronMutex.Unlock()
}

func (j *ProcessCronjobManager) DeployCronJob(
	ctx context.Context,
	name string,
	period string,
	cronFunction func(),
) error {
	if _, ok := j.getCronMap(name); ok {
		return errors.Newf("Cron job with name %s already exists", name)
	}

	cron := &cronMetadata{
		cronJob: nil,
	}

	j.setCronMap(name, cron)

	if period != "" {
		cronJob, err := j.cronScheduler.Cron(period).Do(cronFunction)
		if err != nil {
			return err
		}

		cron.cronJob = cronJob
	}

	return nil
}

func (j *ProcessCronjobManager) CronJobExists(ctx context.Context, name string) bool {
	_, ok := j.getCronMap(name)
	return ok
}

func (j *ProcessCronjobManager) EditCronJob(ctx context.Context, name string, cronString string, cronFunction func()) error {
	cronMetadata, ok := j.getCronMap(name)
	if !ok {
		return errors.New("Cron job not found")
	} else {
		if cronMetadata.cronJob == nil {
			// This means the current cron job is already paused.
			if cronString == "" {
				return nil
			}

			cronJob, err := j.cronScheduler.Cron(cronString).Do(cronFunction)
			if err != nil {
				return err
			}

			cronMetadata.cronJob = cronJob
		} else {
			if cronString == "" {
				// This means we want to pause the cron job.
				j.cronScheduler.RemoveByReference(cronMetadata.cronJob)
				cronMetadata.cronJob = nil
			} else {
				_, err := j.cronScheduler.Job(cronMetadata.cronJob).Cron(cronString).Update()
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func (j *ProcessCronjobManager) DeleteCronJob(ctx context.Context, name string) error {
	cronMetadata, ok := j.getCronMap(name)
	if ok {
		j.cronScheduler.RemoveByReference(cronMetadata.cronJob)
		j.deleteCronMap(name)
	}

	return nil
}
