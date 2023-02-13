package lib_utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
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

// RunCmd executes command with arg.
// It returns the stdout, stderr, and an error object that contains an informative message that
// combines stdout and stderr.
func RunCmd(command string, arg ...string) (string, string, error) {
	cmd := exec.Command(command, arg...)
	cmd.Env = os.Environ()

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("Error running command: %s. Stdout: %s, Stderr: %s.", command, outb.String(), errb.String())
		log.Errorf(errMsg)
		return outb.String(), errb.String(), errors.New(errMsg)
	}

	return outb.String(), errb.String(), nil
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
