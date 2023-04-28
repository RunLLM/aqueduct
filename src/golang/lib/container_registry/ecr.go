package container_registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ECRCredentials struct {
	Token         string
	ExpireAt      int64
	ProxyEndpoint string
}

func CreateAWSSessionFromAccessKey(accessKeyID, secretAccessKey, region string) (*session.Session, error) {
	creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
	return session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})
}

func CreateAWSSessionFromConfigFile(configFilePath, configFileProfile string) (*session.Session, error) {
	// If the authentication mode is credential file, we need to retrieve the AWS region via
	// `aws configure get region` and explicitly pass it to Terraform.
	region, stderr, err := lib_utils.RunCmd(
		"env",
		[]string{
			fmt.Sprintf("AWS_SHARED_CREDENTIALS_FILE=%s", configFilePath),
			fmt.Sprintf("AWS_PROFILE=%s", configFileProfile),
			"aws",
			"configure",
			"get",
			"region",
		},
		"",
		false,
	)
	// We need to check if stderr is empty because when the region is not specified in the
	// profile, the cmd will error and it will produce an empty stdout and stderr. In this case,
	// we should just set the region to an empty string, which means using the default region.
	if err != nil && stderr != "" {
		return nil, err
	}

	return session.NewSessionWithOptions(session.Options{
		// Use a custom path for the AWS credentials file.
		// Replace /path/to/credentials with your own custom path.
		SharedConfigFiles: []string{configFilePath},
		// Ensure that the AWS SDK looks for the credentials file in the custom path only.
		Profile: configFileProfile,
		// Specify the desired region.
		Config: aws.Config{
			Region: aws.String(strings.TrimRight(region, "\n")),
		},
	})
}

func getECRServiceHandle(conf *shared.ECRConfig) (*ecr.ECR, error) {
	var awsSession *session.Session
	var err error

	if conf.AccessKeyId != "" && conf.SecretAccessKey != "" && conf.Region != "" {
		awsSession, err = CreateAWSSessionFromAccessKey(conf.AccessKeyId, conf.SecretAccessKey, conf.Region)
		if err != nil {
			return nil, errors.Wrap(err, "Error creating AWS session from access key.")
		}
	} else {
		awsSession, err = CreateAWSSessionFromConfigFile(conf.ConfigFilePath, conf.ConfigFileProfile)
		if err != nil {
			return nil, errors.Wrap(err, "Error creating AWS session from config file.")
		}
	}

	return ecr.New(awsSession), nil
}

// ValidateECRImage validates if the given image exists in ECR based on the credentials provided.
func ValidateECRImage(conf *shared.ECRConfig, imageName string) error {
	ecrSvc, err := getECRServiceHandle(conf)
	if err != nil {
		return errors.Wrap(err, "Error getting ECR service handle.")
	}

	repoName := strings.Split(imageName, ":")[0]
	imageTag := strings.Split(imageName, ":")[1]

	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repoName),
		ImageIds: []*ecr.ImageIdentifier{
			{
				ImageTag: aws.String(imageTag),
			},
		},
	}

	_, err = ecrSvc.DescribeImages(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "ImageNotFoundException" {
			return errors.New("Image not found.")
		}

		return err
	}

	return nil
}

// GetECRToken returns ECRCredentials, which contains the ECR token, its expiration time, and proxy endpoint.
func GetECRCredentials(conf *shared.ECRConfig) (ECRCredentials, error) {
	emptyResponse := ECRCredentials{}

	ecrSvc, err := getECRServiceHandle(conf)
	if err != nil {
		return emptyResponse, errors.Wrap(err, "Error getting ECR service handle.")
	}

	result, err := ecrSvc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return emptyResponse, errors.Wrap(err, "Error getting authorization token.")
	}

	auth := *result.AuthorizationData[0]

	decodedToken, err := base64.StdEncoding.DecodeString(*auth.AuthorizationToken)
	if err != nil {
		return emptyResponse, errors.Wrap(err, "Error decoding token.")
	}

	return ECRCredentials{
		Token:         strings.Split(string(decodedToken), ":")[1],
		ExpireAt:      auth.ExpiresAt.Unix(),
		ProxyEndpoint: *auth.ProxyEndpoint,
	}, nil
}

// AuthenticateAndUpdateECRConfig authenticates the given auth config for ECR.
// It *also updates* `authConf` with the ECR token, its expiration time, and proxy endpoint.
func AuthenticateAndUpdateECRConfig(authConf auth.Config) error {
	conf, err := lib_utils.ParseECRConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	if conf.AccessKeyId != "" && conf.SecretAccessKey != "" && conf.Region != "" {
		if conf.ConfigFilePath != "" || conf.ConfigFileProfile != "" {
			return errors.New("When authenticating via access keys, credential file path and profile must be empty.")
		}
	} else if conf.ConfigFilePath != "" && conf.ConfigFileProfile != "" {
		if conf.AccessKeyId != "" || conf.SecretAccessKey != "" || conf.Region != "" {
			return errors.New("When authenticating via credential file, access key fields must be empty.")
		}
	} else {
		return errors.New("Either 1) AWS access key ID, secret access key, region, or 2) credential file path, profile must be provided.")
	}

	ecrCredentials, err := GetECRCredentials(conf)
	if err != nil {
		return errors.Wrap(err, "Error getting ECR credentials.")
	}

	// Update authConf with the token, its expiration time, and proxy endpoint.
	castedConf := authConf.(*auth.StaticConfig)
	castedConf.Set("token", ecrCredentials.Token)
	castedConf.Set("expire_at", strconv.FormatInt(ecrCredentials.ExpireAt, 10))
	castedConf.Set("proxy_endpoint", ecrCredentials.ProxyEndpoint)

	return nil
}

// RefreshECRCredentialsIfNeeded checks if the ECR token has expired.
// If so, it refreshes the token and writes the updated config to vault.
// It returns the updated ECR config.
func RefreshECRCredentialsIfNeeded(config auth.Config, registryID uuid.UUID, vaultObject vault.Vault) (*shared.ECRConfig, error) {
	ecrConfig, err := lib_utils.ParseECRConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse ECR config.")
	}

	// If the ECR token is expired, we need to update it.
	expirationTime, err := strconv.ParseInt(ecrConfig.ExpireAt, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing expiration time.")
	}

	if expirationTime < time.Now().Unix() {
		log.Info("ECR token has expired. Refreshing...")
		if err := AuthenticateAndUpdateECRConfig(config); err != nil {
			return nil, errors.Wrap(err, "Error updating ECR config.")
		}

		// Store config (including confidential information) in vault
		if err := auth.WriteConfigToSecret(
			context.Background(),
			registryID,
			config,
			vaultObject,
		); err != nil {
			return nil, errors.Wrap(err, "Unable to save refreshed credential to vault.")
		}

		ecrConfig, err = lib_utils.ParseECRConfig(config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse ECR config.")
		}
	}

	return ecrConfig, nil
}
