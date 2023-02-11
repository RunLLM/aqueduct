package lambda

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aqueducthq/aqueduct/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

type EcrAuth struct {
	Username      string
	Password      string
	ProxyEndpoint string
}

func AuthenticateDockerToECR() error {
	// Authenticate ECR client.
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	ecrSvc := ecr.New(sess)

	token, err := ecrSvc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return errors.Wrap(err, "Unable to get authorization token.")
	}
	auth, err := extractToken(*token.AuthorizationData[0].AuthorizationToken, *token.AuthorizationData[0].ProxyEndpoint)
	if err != nil {
		return errors.Wrap(err, "Unable to extract username and password.")
	}

	cmd := exec.Command(
		"docker",
		"login",
		"--username",
		auth.Username,
		"--password",
		auth.Password,
		auth.ProxyEndpoint,
	)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		log.Info(stdout.String())
		log.Info(stderr.String())
		return errors.Wrap(err, "Unable to authenticate docker client to ECR.")
	}
	return nil
}

func PullImageFromECR(functionType LambdaFunctionType) error {
	// Pull the Image from public ECR Library.
	lambdaImageUri, _, err := mapFunctionType(functionType)
	if err != nil {
		return errors.Wrap(err, "Unable to map function type to image.")
	}
	versionedLambdaImageUri := fmt.Sprintf("%s:%s", lambdaImageUri, lib.ServerVersionNumber)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command("docker", "pull", versionedLambdaImageUri)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		log.Info(stdout.String())
		log.Info(stderr.String())
		return errors.Wrap(err, "Unable to pull docker image from dockerhub.")
	}
	return nil
}

func CreateLambdaFunction(functionType LambdaFunctionType, roleArn string) error {
	// For each lambda function we create, we take the following steps:
	// 1. Pull the image from the public ECR repository.
	// 2. Create the private ECR repo if it doesn't exist
	// 3. Get the ECR auth token and log in the docker client.
	// 4. Push the image to the private ECR repo
	// 5. Create the lambda function using the private ECR repo as the code.
	lambdaImageUri, userRepoName, err := mapFunctionType(functionType)
	if err != nil {
		return errors.Wrap(err, "Unable to map function type to image.")
	}
	versionedLambdaImageUri := fmt.Sprintf("%s:%s", lambdaImageUri, lib.ServerVersionNumber)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	ecrSvc := ecr.New(sess)

	_, err = ecrSvc.DeleteRepository(
		&ecr.DeleteRepositoryInput{
			Force:          aws.Bool(true),
			RepositoryName: aws.String(userRepoName),
		},
	)
	if err != nil {
		// No need to fail here, repository doesn't exist.
		log.Info(err)
	}

	result, err := ecrSvc.CreateRepository(&ecr.CreateRepositoryInput{
		RepositoryName: aws.String(userRepoName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() != ecr.ErrCodeRepositoryAlreadyExistsException {
				return errors.Wrap(err, "Unable to create ECR repository.")
			}
		} else {
			return errors.Wrap(err, "Unable to create ECR repository.")
		}
	}

	repositoryUri := fmt.Sprintf("%s:%s", *result.Repository.RepositoryUri, lib.ServerVersionNumber)

	cmd := exec.Command("docker", "tag", versionedLambdaImageUri, repositoryUri)
	err = cmd.Run()
	if err != nil {
		log.Info(stdout.String())
		log.Info(stderr.String())
		return errors.Wrap(err, "Unable to tag docker image from ECR.")
	}

	cmd = exec.Command("docker", "push", repositoryUri)
	err = cmd.Run()
	if err != nil {
		log.Info(stdout.String())
		log.Info(stderr.String())
		return errors.Wrap(err, "Unable to push docker image to ECR.")
	}

	lambdaService := lambda.New(sess)
	_, err = lambdaService.GetFunction(&lambda.GetFunctionInput{FunctionName: &userRepoName})
	if err != nil {
		// Function doesn't exist and needs to be created.
		createArgs := &lambda.CreateFunctionInput{
			Code: &lambda.FunctionCode{
				ImageUri: &repositoryUri,
			},
			FunctionName: &userRepoName,
			Role:         &roleArn,
			PackageType:  aws.String("Image"),
			Publish:      aws.Bool(true),
			MemorySize:   aws.Int64(1000),
			Timeout:      aws.Int64(300),
		}

		_, err := lambdaService.CreateFunction(createArgs)
		if err != nil {
			return errors.Wrap(err, "Unable to create lambda function.")
		}
	} else {
		// Function does exist and needs to be updated.
		updateArgs := &lambda.UpdateFunctionCodeInput{
			FunctionName: &userRepoName,
			ImageUri:     &repositoryUri,
			Publish:      aws.Bool(true),
		}
		_, err := lambdaService.UpdateFunctionCode(updateArgs)
		if err != nil {
			return errors.Wrap(err, "Unable to update lambda function.")
		}
	}

	err = DeleteDockerImage(versionedLambdaImageUri)
	if err != nil {
		return errors.Wrap("finishing", err, "Unable to delete downloaded docker image.")
	}
	return nil
}

func extractToken(token string, proxyEndpoint string) (*EcrAuth, error) {
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to decode token.")
	}

	parts := strings.SplitN(string(decodedToken), ":", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid token: expected two parts, got %d", len(parts))
	}

	return &EcrAuth{
		Username:      parts[0],
		Password:      parts[1],
		ProxyEndpoint: proxyEndpoint,
	}, nil
}

func mapFunctionType(functionType LambdaFunctionType) (string, string, error) {
	switch functionType {
	case FunctionExecutor37Type:
		return FunctionLambdaImage37, FunctionLambdaFunction37, nil
	case FunctionExecutor38Type:
		return FunctionLambdaImage38, FunctionLambdaFunction38, nil
	case FunctionExecutor39Type:
		return FunctionLambdaImage39, FunctionLambdaFunction39, nil
	case ParamExecutorType:
		return ParameterLambdaImage, ParameterLambdaFunction, nil
	case SystemMetricType:
		return SystemMetricLambdaImage, SystemMetricLambdaFunction, nil
	case AthenaConnectorType:
		return AthenaConnectorLambdaImage, AthenaLambdaFunction, nil
	case BigQueryConnectorType:
		return BigQueryConnectorLambdaImage, BigQueryLambdaFunction, nil
	case PostgresConnectorType:
		return PostgresConnectorLambdaImage, PostgresLambdaFunction, nil
	case S3ConnectorType:
		return S3ConnectorLambdaImage, S3LambdaFunction, nil
	case SnowflakeConnectorType:
		return SnowflakeConnectorLambdaImage, SnowflakeLambdaFunction, nil
	default:
		return "", "", errors.New("Invalide function type")

	}
}
