package lambda

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type EcrAuth struct {
	Username      string
	Password      string
	ProxyEndpoint string
}

const (
	MaxConcurrentDownload = 3
	MaxConcurrentUpload   = 5
	RoleArnKey            = "role_arn"
)

func ConnectToLambda(
	ctx context.Context,
	lambdaRoleArn string,
) error {
	functionsToShip := [10]LambdaFunctionType{
		FunctionExecutor37Type,
		FunctionExecutor38Type,
		FunctionExecutor39Type,
		ParamExecutorType,
		SystemMetricType,
		AthenaConnectorType,
		BigQueryConnectorType,
		PostgresConnectorType,
		S3ConnectorType,
		SnowflakeConnectorType,
	}

	err := AuthenticateDockerToECR()
	if err != nil {
		return errors.Wrap(err, "Unable to authenticate Lambda Function.")
	}

	err = CreateLambdaFunction(ctx, functionsToShip[:], lambdaRoleArn)
	if err != nil {
		return errors.Wrap(err, "Unable to create Lambda Function.")
	}
	return nil
}

func CreateLambdaFunction(ctx context.Context, functionsToShip []LambdaFunctionType, roleArn string) error {
	// For each lambda function we create, we take the following steps:
	// 1. Pull the image from the public ECR repository on a concurrency of `MaxConcurrentDownload`.
	// 2. Create the private ECR repo if it doesn't exist.
	// 3. Get the ECR auth token and log in the docker client.
	// 4. Push the image to the private ECR repo on a concurrency of `MaxConcurrentUpload`.
	// 5. Create the lambda function using the private ECR repo as the code.

	errGroup, errGroupCtx := errgroup.WithContext(ctx)

	// Create a `pullImageChannel` and `pushImageChannel` and add all lambda functions to the channel.
	pullImageChannel := make(chan LambdaFunctionType, len(functionsToShip))
	defer close(pullImageChannel)
	pushImageChannel := make(chan LambdaFunctionType, len(functionsToShip))
	defer close(pushImageChannel)
	AddFunctionTypeToChannel(functionsToShip[:], pullImageChannel)

	for i := 0; i < MaxConcurrentDownload; i++ {
		errGroup.Go(func() error {
			for {
				select {
				case functionType := <-pullImageChannel:
					lambdaFunctionType := functionType
					err := PullImageFromPublicECR(lambdaFunctionType)
					if err != nil {
						return err
					}
					pushImageChannel <- functionType
				case <-errGroupCtx.Done():
					return errGroupCtx.Err()
				default:
					// The case should only be hit when `pullImageChannel` is empty.
					return nil
				}
			}
		})
	}

	// The `incompleteWorkChannel` is empty if and only if all the all lambda functions are successfully created.
	incompleteWorkChannel := make(chan LambdaFunctionType, len(functionsToShip))
	defer close(incompleteWorkChannel)
	AddFunctionTypeToChannel(functionsToShip[:], incompleteWorkChannel)

	// Receive the downloaded docker images from push channels and create lambda functions on a concurrency of "MaxConcurrentUpload".
	for i := 0; i < MaxConcurrentUpload; i++ {
		errGroup.Go(func() error {
			for {
				select {
				case functionType := <-pushImageChannel:
					lambdaFunctionType := functionType
					err := PushImageToPrivateECR(lambdaFunctionType, roleArn)
					if err != nil {
						return err
					}
					<-incompleteWorkChannel
				case <-errGroupCtx.Done():
					return errGroupCtx.Err()
				default:
					time.Sleep(1 * time.Second)
					if len(incompleteWorkChannel) == 0 {
						return nil
					}
				}
			}
		})
	}

	if err := errGroup.Wait(); err != nil {
		err = DeleteAllDockerImages(functionsToShip[:])
		if err != nil {
			return errors.Wrap(err, "Unable to delete lambda functions.")
		}
		return errors.Wrap(err, "Unable to Create lambda functions.")
	}

	return nil
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

func PullImageFromPublicECR(functionType LambdaFunctionType) error {
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

func PushImageToPrivateECR(functionType LambdaFunctionType, roleArn string) error {
	// Push the image to the private ECR repo and create the lambda function using the private ECR repo as the code.
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
			return errors.Wrap(err, "Unable to create lambda function with the roleArn.")
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
		return errors.Wrap(err, "Unable to delete lambda function.")
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

func AddFunctionTypeToChannel(functionsToShip []LambdaFunctionType, channel chan LambdaFunctionType) {
	// Add lambda function types to buffered channel for pulling and creating lambda function.
	for _, lambdaFunctionType := range functionsToShip {
		lambdaFunctionTypeToPass := lambdaFunctionType
		channel <- lambdaFunctionTypeToPass
	}
}

func DeleteAllDockerImages(functionsToShip []LambdaFunctionType) error {
	for _, lambdaFunctionType := range functionsToShip {
		lambdaImageUri, _, err := mapFunctionType(lambdaFunctionType)
		if err != nil {
			return errors.Wrap(err, "Unable to map function type to image.")
		}
		versionedLambdaImageUri := fmt.Sprintf("%s:%s", lambdaImageUri, lib.ServerVersionNumber)
		err = DeleteDockerImage(versionedLambdaImageUri)
		if err != nil {
			return errors.Wrapf(err, "Unable to delete docker image of %s", versionedLambdaImageUri)
		}
	}
	return nil
}

func DeleteDockerImage(versionedLambdaImageUri string) error {
	// Remove Docker Image after finish creating Lambda functions.

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command(
		"docker",
		"images",
		fmt.Sprintf("--filter=reference=%s", versionedLambdaImageUri),
		"-q")

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		stdoutString := strings.TrimSpace(stdout.String())
		stderrString := strings.TrimSpace(stderr.String())
		return errors.Wrapf(err, "Unable to find the docker image. Stdout: %s. Stderr: %s.", stdoutString, stderrString)
	}

	imageId := strings.TrimSpace(stdout.String())

	// The docker image does not exist or has been deleted so we can't delete it again.
	if imageId == "" {
		return nil
	}

	cmd = exec.Command("docker", "rmi", "-f", imageId)

	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()

	if err != nil {
		stdoutString := strings.TrimSpace(stdout.String())
		stderrString := strings.TrimSpace(stderr.String())
		return errors.Wrapf(err, "Unable to delete docker image. Stdout: %s. Stderr: %s.", stdoutString, stderrString)
	}

	return nil
}
