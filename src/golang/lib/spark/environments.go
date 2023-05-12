package spark

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aqueducthq/aqueduct/config"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

// CreateSparkEnvFile runs a Docker image to create and pack a conda environment
// with the user-provided requirements. This env is then pushed to S3 where Spark
// retrieves it to be consumed as the environment where Aqueduct operators run in.
func CreateSparkEnvFile(
	ctx context.Context,
	workflowEnv *exec_env.ExecutionEnvironment,
) (string, error) {
	storageConfig := config.Storage()
	if storageConfig.Type != shared.S3StorageType {
		return "", errors.New("Must use S3 storage config for Spark engine.")
	}

	storageObj := storage.NewStorage(&storageConfig)
	keyId, secretKey, err := lib_utils.ExtractAwsCredentials(storageConfig.S3Config)
	if err != nil {
		return "", errors.Wrap(err, "Unable to extract AWS credentials from file.")
	}

	// We use the env name to avoid duplicated s3 file
	s3EnvPath := fmt.Sprintf("%s.tar.gz", workflowEnv.Name())
	sparkEnvPath := fmt.Sprintf("%s/%s", storageConfig.S3Config.Bucket, s3EnvPath)
	if storageObj.Exists(ctx, s3EnvPath) {
		return sparkEnvPath, nil
	}

	sparkImage, err := PullSparkImage(workflowEnv.PythonVersion)
	if err != nil {
		return "", err
	}

	err = RunSparkImage(
		strings.Join(workflowEnv.Dependencies, " "),
		s3EnvPath,
		keyId,
		secretKey,
		storageConfig.S3Config.Region,
		storageConfig.S3Config.Bucket,
		sparkImage,
	)
	if err != nil {
		return "", err
	}
	if storageObj.Exists(ctx, s3EnvPath) {
		return sparkEnvPath, nil
	} else {
		return "", errors.New("Unable to find spark env in S3.")
	}
}

func RunSparkImage(
	dependencyList string,
	envFileName string,
	accessKeyID string,
	secretAccessKey string,
	region string,
	bucket string,
	versionedSparkImage string,
) error {
	cmd := exec.Command(
		"docker",
		"run",
		"-e",
		fmt.Sprintf("DEPENDENCIES=%s", dependencyList),
		"-e",
		fmt.Sprintf("ENV_FILE_NAME=%s", envFileName),
		"-e",
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", accessKeyID),
		"-e",
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", secretAccessKey),
		"-e",
		fmt.Sprintf("AWS_REGION=%s", region),
		"-e",
		fmt.Sprintf("S3_BUCKET=%s", bucket),
		"-e",
		fmt.Sprintf("VERSION_TAG=\"%s\"", config.VersionTag()),
		versionedSparkImage,
	)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		log.Info(stdout.String())
		log.Info(stderr.String())
		return errors.Wrap(err, "Error executing docker run command.")
	}
	return nil
}

func PullSparkImage(pythonVersion string) (string, error) {
	// Pull the Image from public ECR Library.
	sparkImage, err := mapPythonVerisonToSparkDockerImage(pythonVersion)
	if err != nil {
		return "", errors.Wrap(err, "Unable to map function type to image.")
	}
	versionedSparkImage := fmt.Sprintf("%s:%s", sparkImage, "hari_test")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command("docker", "pull", versionedSparkImage)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		log.Info(stdout.String())
		log.Info(stderr.String())
		return "", errors.Wrap(err, "Unable to pull docker image from dockerhub.")
	}
	return versionedSparkImage, nil
}

func mapPythonVerisonToSparkDockerImage(pythonVersion string) (string, error) {
	switch pythonVersion {
	case python37:
		return python37SparkImage, nil
	case python38:
		return python38SparkImage, nil
	case python39:
		return python39SparkImage, nil
	case python310:
		return python310SparkImage, nil
	default:
		return "", errors.New("Unsupported python version.")
	}
}
