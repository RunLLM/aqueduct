package airflow

import (
	"strings"
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
