package airflow

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/dropbox/godropbox/errors"
)

// generateDagId generates an Airflow DAG ID for a workflow.
func generateDagId(workflowName string) (string, error) {
	return prepareId(workflowName)
}

// generateTaskId generates an Airflow task ID for an operator.
func generateTaskId(operatorName string) (string, error) {
	return prepareId(operatorName)
}

// prepareId replaces all non-alphanumeric characters in `s` with
// an underscore, as Airflow only allows alphanumerics
// and certain special characters for DAG and task IDs.
func prepareId(s string) (string, error) {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if ('a' <= c && c <= 'z') ||
			('A' <= c && c <= 'Z') ||
			('0' <= c && c <= '9') {
			if err := result.WriteByte(c); err != nil {
				return "", err
			}
		} else {
			// `c` is not an alphanumeric
			if _, err := result.WriteString("_"); err != nil {
				return "", err
			}
		}
	}

	return result.String(), nil
}

// wrapApiErrors wraps an error from the Airflow API using the error returned
// and the HTTP response.
func wrapApiError(err error, api string, resp *http.Response) error {
	return errors.Wrapf(err, "Airflow %v error with status: %v", api, resp.Status)
}

// mapDagStateToStatus maps an Airflow DagState to an ExecutionStatus
func mapDagStateToStatus(state airflow.DagState) shared.ExecutionStatus {
	switch state {
	case airflow.DAGSTATE_QUEUED:
		return shared.PendingExecutionStatus
	case airflow.DAGSTATE_RUNNING:
		return shared.RunningExecutionStatus
	case airflow.DAGSTATE_SUCCESS:
		return shared.SucceededExecutionStatus
	case airflow.DAGSTATE_FAILED:
		return shared.FailedExecutionStatus
	default:
		return shared.UnknownExecutionStatus
	}
}

// mapTaskStateToStatus maps an Airflow TaskState to an ExecutionStatus
func mapTaskStateToStatus(state airflow.TaskState) shared.ExecutionStatus {
	switch state {
	case airflow.TASKSTATE_RUNNING:
		return shared.RunningExecutionStatus
	case airflow.TASKSTATE_SUCCESS:
		return shared.SucceededExecutionStatus
	case airflow.TASKSTATE_FAILED:
		return shared.FailedExecutionStatus
	default:
		return shared.PendingExecutionStatus
	}
}

func getOperatorMetadataPath(metadataPathPrefix string, dagRunId string) string {
	return fmt.Sprintf("%s_%s", metadataPathPrefix, dagRunId)
}

func getArtifactMetadataPath(metadataPathPrefix string, dagRunId string) string {
	return fmt.Sprintf("%s_%s", metadataPathPrefix, dagRunId)
}

func getArtifactContentPath(contentPathPrefix string, dagRunId string) string {
	return fmt.Sprintf("%s_%s", contentPathPrefix, dagRunId)
}
