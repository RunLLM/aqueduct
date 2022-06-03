package job

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPythonExecutorPackage = "aqueduct_executor"
	connectorPythonPath          = "operators.connectors.tabular.main"
	paramPythonPath              = "operators.param_executor.main"
	workflowExecutorBinary       = "executor"
	functionExecutorBashScript   = "start-function-executor.sh"

	processRunningStatus = "R"

	BinaryDir          = "bin/"
	OperatorStorageDir = "storage/operators/"
)

var (
	defaultBinaryDir          = path.Join(os.Getenv("HOME"), ".aqueduct", BinaryDir)
	defaultOperatorStorageDir = path.Join(os.Getenv("HOME"), ".aqueduct", OperatorStorageDir)
)

type Command struct {
	cmd    *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

type cronMetadata struct {
	// If the cronJob is nil, it means the corresponding workflow has been paused.
	cronJob *gocron.Job
	// We need to store the job spec because when the workflow is resumed from the pause state, we
	// need the spec to redeploy the cron job.
	jobSpec Spec
}

type ProcessJobManager struct {
	conf          *ProcessConfig
	cmds          map[string]*Command
	cronScheduler *gocron.Scheduler
	// A mapping from cron job name to cron job object pointer.
	cronMapping map[string]*cronMetadata
}

func NewProcessJobManager(conf *ProcessConfig) (*ProcessJobManager, error) {
	if conf.PythonExecutorPackage == "" {
		conf.PythonExecutorPackage = defaultPythonExecutorPackage
	}

	if conf.BinaryDir == "" {
		conf.BinaryDir = defaultBinaryDir
	}

	if conf.OperatorStorageDir == "" {
		conf.OperatorStorageDir = defaultOperatorStorageDir
	}

	cronScheduler := gocron.NewScheduler(time.UTC)
	cronScheduler.StartAsync()

	return &ProcessJobManager{
		conf:          conf,
		cmds:          map[string]*Command{},
		cronScheduler: cronScheduler,
		cronMapping:   map[string]*cronMetadata{},
	}, nil
}

func (j *ProcessJobManager) mapJobTypeToCmd(spec Spec) (*exec.Cmd, error) {
	if spec.Type() == WorkflowJobType {
		workflowSpec, ok := spec.(*WorkflowSpec)
		if !ok {
			return nil, errors.New("Unable to cast job spec to WorkflowSpec.")
		}

		specStr, err := EncodeSpec(workflowSpec, GobSerializationType)
		if err != nil {
			return nil, err
		}

		return exec.Command(
			fmt.Sprintf("%s/%s", j.conf.BinaryDir, workflowExecutorBinary),
			"--spec",
			specStr,
		), nil
	} else if spec.Type() == FunctionJobType {
		functionSpec, ok := spec.(*FunctionSpec)
		if !ok {
			return nil, ErrInvalidJobSpec
		}

		functionSpec.FunctionExtractPath = path.Join(j.conf.OperatorStorageDir, uuid.New().String())
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return nil, err
		}

		return exec.Command(
			"bash",
			filepath.Join(j.conf.BinaryDir, functionExecutorBashScript),
			specStr,
		), nil
	} else if spec.Type() == ParamJobType {
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return nil, err
		}

		return exec.Command(
			"python3",
			"-m",
			fmt.Sprintf("%s.%s", j.conf.PythonExecutorPackage, paramPythonPath),
			"--spec",
			specStr,
		), nil
	} else {
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return nil, err
		}

		return exec.Command(
			"python3",
			"-m",
			fmt.Sprintf("%s.%s", j.conf.PythonExecutorPackage, connectorPythonPath),
			"--spec",
			specStr,
		), nil
	}
}

func (j *ProcessJobManager) generateCronFunction(name string, jobSpec Spec) func() {
	return func() {
		jobName := fmt.Sprintf("%s-%d", name, time.Now().Unix())
		log.Infof("Running cron job %s", jobName)
		err := j.Launch(context.Background(), jobName, jobSpec)
		if err != nil {
			log.Errorf("Error running cron job %s: %v", jobName, err)
		}

		j.Poll(context.Background(), jobName)
		else {
			log.Infof("Successfully ran cron job %s", jobName)
		}
	}
}

func (j *ProcessJobManager) Config() Config {
	return j.conf
}

func (j *ProcessJobManager) Launch(
	ctx context.Context,
	name string,
	spec Spec,
) error {
	if _, ok := j.cmds[name]; ok {
		return ErrJobAlreadyExists
	}

	cmd, err := j.mapJobTypeToCmd(spec)
	cmd.Env = os.Environ()
	if err != nil {
		return err
	}

	j.cmds[name] = &Command{
		cmd:    cmd,
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
	cmd.Stdout = j.cmds[name].stdout
	cmd.Stderr = j.cmds[name].stderr

	return cmd.Start()
}

func (j *ProcessJobManager) Poll(ctx context.Context, name string) (shared.ExecutionStatus, error) {
	command, ok := j.cmds[name]
	if !ok {
		return shared.UnknownExecutionStatus, ErrJobNotExist
	}

	proc, err := process.NewProcess(int32(command.cmd.Process.Pid))
	if err != nil {
		return shared.UnknownExecutionStatus, err
	}

	status, err := proc.Status()
	if err != nil {
		return shared.UnknownExecutionStatus, err
	}

	if status == processRunningStatus {
		return shared.PendingExecutionStatus, nil
	}

	err = command.cmd.Wait()
	// After wait, we are done with this job and already consumed all of its output, so we garbage
	// collect the entry in j.cmds.
	defer delete(j.cmds, name)
	if err != nil {
		log.Errorf("Unexpected error occured while executing the job: \nStdout: %s\nStderr: %s",
			command.stdout.String(),
			command.stderr.String(),
		)

		return shared.FailedExecutionStatus, nil
	}

	return shared.SucceededExecutionStatus, nil
}

func (j *ProcessJobManager) DeployCronJob(
	ctx context.Context,
	name string,
	period string,
	spec Spec,
) error {
	if _, ok := j.cronMapping[name]; ok {
		return errors.Newf("Cron job with name %s already exists", name)
	}

	j.cronMapping[name] = &cronMetadata{
		cronJob: nil,
		jobSpec: spec,
	}

	if period != "" {
		cronJob, err := j.cronScheduler.Cron(period).Do(j.generateCronFunction(name, spec))
		if err != nil {
			return err
		}

		j.cronMapping[name].cronJob = cronJob
	}

	return nil
}

func (j *ProcessJobManager) CronJobExists(ctx context.Context, name string) bool {
	_, ok := j.cronMapping[name]
	return ok
}

func (j *ProcessJobManager) EditCronJob(ctx context.Context, name string, cronString string) error {
	cronMetadata, ok := j.cronMapping[name]
	if !ok {
		return errors.New("Cron job not found")
	} else {
		if cronMetadata.cronJob == nil {
			// This means the current cron job is paused.
			if cronString == "" {
				return errors.Newf("Attempting to pause an already paused cron job %s", name)
			}

			cronJob, err := j.cronScheduler.Cron(cronString).Do(j.generateCronFunction(name, cronMetadata.jobSpec))
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

func (j *ProcessJobManager) DeleteCronJob(ctx context.Context, name string) error {
	cronMetadata, ok := j.cronMapping[name]
	if ok {
		j.cronScheduler.RemoveByReference(cronMetadata.cronJob)
		delete(j.cronMapping, name)
	}

	return nil
}
