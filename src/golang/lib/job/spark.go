package job

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/spark"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const defaultSparkFunctionExtractPath = "/tmp/function/"

type SparkJobManager struct {
	livyClient *spark.LivyClient
	sessionID  int
	conf       *SparkJobManagerConfig
	runMap     map[string]int
}

func NewSparkJobManager(conf *SparkJobManagerConfig) (*SparkJobManager, error) {
	livyClient := spark.NewLivyClient(conf.LivyServerURL)

	session, err := livyClient.CreateSession(&spark.CreateSessionRequest{
		Kind:                     "pyspark",
		HeartbeatTimeoutInSecond: 600,
		Archives:                 []string{fmt.Sprintf("%s#environment", conf.EnvironmentPathURI)},
		Conf: map[string]string{
			"spark.yarn.appMasterEnv.PYSPARK_PYTHON": "./environment/bin/python",
			"spark.jars.packages":                    "net.snowflake:snowflake-jdbc:3.13.28,net.snowflake:spark-snowflake_2.12:2.11.1-spark_3.3",
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating session on spark.")
	}
	err = livyClient.WaitForSession(session.ID, time.Minute*3)
	if err != nil {
		return nil, errors.Wrap(err, "Timeout waiting for sesion to create.")
	}

	return &SparkJobManager{
		livyClient: livyClient,
		sessionID:  session.ID,
		conf:       conf,
		runMap:     map[string]int{},
	}, nil
}

func (j *SparkJobManager) Config() Config {
	return j.conf
}

func (j *SparkJobManager) Launch(
	ctx context.Context,
	name string,
	spec Spec,
) JobError {
	scriptString, err := j.mapJobTypeToScript(spec)
	if err != nil {
		return systemError(err)
	}
	log.Info(scriptString)
	statement, err := j.livyClient.RunStatement(j.sessionID, &spark.StatementRequest{
		Code: scriptString,
	})
	if err != nil {
		return systemError(err)
	}
	j.runMap[name] = statement.ID

	return nil
}

func (j *SparkJobManager) Poll(ctx context.Context, name string) (shared.ExecutionStatus, JobError) {
	statmentID, ok := j.runMap[name]
	if !ok {
		return shared.UnknownExecutionStatus, jobMissingError(errors.New("Job doesn't exist."))
	}
	statement, err := j.livyClient.GetStatement(j.sessionID, statmentID)
	if err != nil {
		return shared.UnknownExecutionStatus, systemError(errors.Wrap(err, "Unable to get Session from spark."))
	}
	switch statement.State {
	case spark.Waiting, spark.Running, spark.Cancelling:
		return shared.RunningExecutionStatus, nil
	case spark.StatementError:
		return shared.FailedExecutionStatus, nil
	case spark.Available:
		switch statement.Output.Status {
		case spark.Error:
			return shared.FailedExecutionStatus, nil
		case spark.OK:
			return shared.SucceededExecutionStatus, nil
		}
	}
	return shared.UnknownExecutionStatus, nil
}

func (j *SparkJobManager) DeployCronJob(
	ctx context.Context,
	name string,
	period string,
	spec Spec,
) JobError {
	return nil
}

func (j *SparkJobManager) CronJobExists(ctx context.Context, name string) bool {
	return false
}

func (j *SparkJobManager) EditCronJob(ctx context.Context, name string, cronString string) JobError {
	return nil
}

func (j *SparkJobManager) DeleteCronJob(ctx context.Context, name string) JobError {
	return nil
}

func (j *SparkJobManager) mapJobTypeToScript(spec Spec) (string, error) {
	// Add S3 Access Keys to all specs
	storageConfig, err := spec.GetStorageConfig()
	if err != nil {
		return "", errors.Wrap(err, "Spec unexpectedly has no storage config.")
	}
	storageConfig.S3Config.AWSAccessKeyID = j.conf.AwsAccessKeyID
	storageConfig.S3Config.AWSSecretAccessKey = j.conf.AwsSecretAccessKey
	var scriptString string
	log.Infof("JobType : %s", spec.Type())
	if spec.Type() == FunctionJobType {
		functionSpec, ok := spec.(*FunctionSpec)
		if !ok {
			return "", ErrInvalidJobSpec()
		}

		functionSpec.FunctionExtractPath = path.Join(defaultSparkFunctionExtractPath, uuid.New().String())
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return "", err
		}

		scriptString = fmt.Sprintf(spark.FunctionEntrypoint, specStr)
	} else if spec.Type() == ParamJobType {
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return "", err
		}

		scriptString = fmt.Sprintf(spark.ParamEntrypoint, specStr)
	} else if IsDataType(spec.Type()) {
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return "", err
		}

		scriptString = fmt.Sprintf(spark.DataEntrypoint, specStr)
	} else if spec.Type() == SystemMetricJobType {
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return "", err
		}

		scriptString = fmt.Sprintf(spark.SystemMetricEntrypoint, specStr)
	} else {
		return "", errors.New("Unsupported JobType was passed in.")
	}
	return scriptString, nil
}
