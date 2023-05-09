package lib_utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This function appends a prefix to the resource name
// so that it conforms to the k8s's accepted format (name must start with an alphabet).
func AppendPrefix(name string) string {
	return fmt.Sprintf("aqueduct-%s", name)
}

func ParseStatus(st *status.Status) (string, int) {
	var errorMsg string
	var ok bool

	if len(st.Details()) == 0 {
		errorMsg = st.Message()
	} else {
		errorMsg, ok = st.Details()[0].(string) // Details should only have one object, and it should be a string.
		if !ok {
			log.Errorf("Unable to correctly parse gRPC status: %v\n", st)
		}
	}

	var errorCode int
	if st.Code() == codes.InvalidArgument {
		errorCode = http.StatusBadRequest
	} else if st.Code() == codes.Internal {
		errorCode = http.StatusInternalServerError
	} else if st.Code() == codes.NotFound {
		errorCode = http.StatusNotFound
	} else {
		errorCode = http.StatusInternalServerError
	}

	return errorMsg, errorCode
}

// RunCmd executes command with args under working directory dir.
// If stream is set to true, it streams the stdout and stderr and returns an error object indicating
// whether the cmd succeeded. Otherwise, it stores the stdout, stderr into buffers and returns them
// to the caller together with an error object indicating whether the cmd succeeded.
func RunCmd(command string, args []string, dir string, stream bool) (string, string, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()

	if dir != "" {
		cmd.Dir = dir
	}

	log.Infof("Running command %s", cmd.String())
	if stream {
		// create pipes for the command's standard output and standard error
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return "", "", errors.Wrap(err, "Error creating stdout pipe")
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return "", "", errors.Wrap(err, "Error creating stderr pipe")
		}

		// start the command
		if err := cmd.Start(); err != nil {
			return "", "", errors.Wrap(err, "Error starting command")
		}

		// create scanners to read from the pipes
		stdoutScanner := bufio.NewScanner(stdout)
		stderrScanner := bufio.NewScanner(stderr)

		var stderrMsg string
		// Create a channel to communicate between the main process and the goroutine to exchange
		// the stderr message.
		ch := make(chan string)

		// start separate goroutines to stream the output from each scanner
		go func() {
			// When the cmd exits, the scanner will break out of the loop.
			for stdoutScanner.Scan() {
				log.Infof("stdout: %s", stdoutScanner.Text())
			}
		}()
		go func() {
			var sb strings.Builder
			// When the cmd exits, the scanner will break out of the loop.
			for stderrScanner.Scan() {
				log.Errorf("stderr: %s", stderrScanner.Text())
				sb.WriteString(stderrScanner.Text())
				sb.WriteString("\n")
			}
			ch <- sb.String()
		}()

		// Wait for the stderr goroutine to finish and receive the stderr from the channel.
		// Even if there is no stderr message and we send an empty string to the channel,
		// we will still be unblocked.
		stderrMsg = <-ch

		// Wait for the command to complete.
		if err := cmd.Wait(); err != nil {
			return "", stderrMsg, errors.New(stderrMsg)
		}

		return "", "", nil
	} else {
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb

		err := cmd.Run()
		if err != nil {
			errMsg := fmt.Sprintf("Error running command: %s. Stdout: %s, Stderr: %s.", cmd.String(), outb.String(), errb.String())
			return outb.String(), errb.String(), errors.New(errMsg)
		}

		return outb.String(), errb.String(), nil
	}
}

// ParseK8sConfig takes in an auth.Config and parses into a K8s config.
// It also returns an error, if any.
func ParseK8sConfig(conf auth.Config) (*shared.K8sIntegrationConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c shared.K8sIntegrationConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func ParseLambdaConfig(conf auth.Config) (*shared.LambdaIntegrationConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c shared.LambdaIntegrationConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func ParseDatabricksConfig(conf auth.Config) (*shared.DatabricksIntegrationConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c shared.DatabricksIntegrationConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func ParseEmailConfig(conf auth.Config) (*shared.EmailConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c struct {
		User              string                   `json:"user"`
		Password          string                   `json:"password"`
		Host              string                   `json:"host"`
		Port              string                   `json:"port"`
		TargetsSerialized string                   `json:"targets_serialized"`
		Level             shared.NotificationLevel `json:"level"`
		Enabled           string                   `json:"enabled"`
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	var targets []string
	if err := json.Unmarshal([]byte(c.TargetsSerialized), &targets); err != nil {
		return nil, err
	}

	return &shared.EmailConfig{
		User:     c.User,
		Password: c.Password,
		Host:     c.Host,
		Port:     c.Port,
		Targets:  targets,
		Level:    c.Level,
		Enabled:  c.Enabled == "true",
	}, nil
}

func ParseSlackConfig(conf auth.Config) (*shared.SlackConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c struct {
		Token              string                   `json:"token"`
		ChannelsSerialized string                   `json:"channels_serialized"`
		Level              shared.NotificationLevel `json:"level"`
		Enabled            string                   `json:"enabled"`
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	var channels []string
	if err := json.Unmarshal([]byte(c.ChannelsSerialized), &channels); err != nil {
		return nil, err
	}

	return &shared.SlackConfig{
		Token:    c.Token,
		Channels: channels,
		Level:    c.Level,
		Enabled:  c.Enabled == "true",
	}, nil
}

func ParseSparkConfig(conf auth.Config) (*shared.SparkIntegrationConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c shared.SparkIntegrationConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func ParseAWSConfig(conf auth.Config) (*shared.AWSConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c struct {
		AccessKeyId       string `json:"access_key_id"`
		SecretAccessKey   string `json:"secret_access_key"`
		Region            string `json:"region"`
		ConfigFilePath    string `json:"config_file_path"`
		ConfigFileProfile string `json:"config_file_profile"`
		K8sSerialized     string `json:"k8s_serialized"`
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	var dynamicK8sConfig shared.DynamicK8sConfig
	if len(c.K8sSerialized) > 0 {
		if err := json.Unmarshal([]byte(c.K8sSerialized), &dynamicK8sConfig); err != nil {
			return nil, err
		}
	}

	return &shared.AWSConfig{
		AccessKeyId:       c.AccessKeyId,
		SecretAccessKey:   c.SecretAccessKey,
		Region:            c.Region,
		ConfigFilePath:    c.ConfigFilePath,
		ConfigFileProfile: c.ConfigFileProfile,
		K8s:               &dynamicK8sConfig,
	}, nil
}

func ParseECRConfig(conf auth.Config) (*shared.ECRConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c shared.ECRConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func ExtractAwsCredentials(config *shared.S3Config) (string, string, error) {
	var awsAccessKeyId string
	var awsSecretAccessKey string
	profileString := fmt.Sprintf("[%s]", config.CredentialsProfile)

	file, err := os.Open(config.CredentialsPath)
	if err != nil {
		return "", "", errors.Wrap(err, "Unable to open AWS credentials file.")
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		if profileString == fileScanner.Text() {
			// Parse `aws_access_key_id`.
			if fileScanner.Scan() {
				accessKeyIdRegex := regexp.MustCompile(`aws_access_key_id\s*=\s*([^\n]+)`)
				match := accessKeyIdRegex.FindStringSubmatch(fileScanner.Text())
				if len(match) < 2 {
					log.Errorf("Unable to scan access key id from credentials file. The file may be malformed.")
					return "", "", errors.New("Unable to extract AWS credentials.")
				}
				awsAccessKeyId = strings.TrimSpace(match[1])
			} else {
				return "", "", errors.New("Unable to extract AWS credentials.")
			}

			// Parse `aws_secret_access_key`.
			if fileScanner.Scan() {
				secretAccessKeyRegex := regexp.MustCompile(`aws_secret_access_key\s*=\s*([^\n]+)`)
				match := secretAccessKeyRegex.FindStringSubmatch(fileScanner.Text())
				if len(match) < 2 {
					log.Errorf("Unable to scan access key id from credentials file. The file may be malformed.")
					return "", "", errors.New("Unable to extract AWS credentials.")
				}
				awsSecretAccessKey = strings.TrimSpace(match[1])
			} else {
				return "", "", errors.New("Unable to extract AWS credentials.")
			}

			return awsAccessKeyId, awsSecretAccessKey, nil
		}
	}
	return "", "", errors.New("Unable to extract AWS credentials.")
}
