package artifact_result

import (
	"fmt"
	"strings"
)

const (
	tableName = "artifact_result"

	// ArtifactResult table column names
	IdColumn                  = "id"
	WorkflowDagResultIdColumn = "workflow_dag_result_id"
	ArtifactIdColumn          = "artifact_id"
	ContentPathColumn         = "content_path"

	// `Status` is initialized to "PENDING" for each new artifact result.
	StatusColumn   = "status"
	MetadataColumn = "metadata"

	// `ExecState` is initialized to nil. Expected to be set on updates only.
	ExecStateColumn = "execution_state"
)

// Returns a joined string of all ArtifactResult columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			WorkflowDagResultIdColumn,
			ArtifactIdColumn,
			ContentPathColumn,
			StatusColumn,
			MetadataColumn,
			ExecStateColumn,
		},
		",",
	)
}

// Returns a joined string of all WorkflowDagResult columns prefixed by table name.
func allColumnsWithPrefix() string {
	return strings.Join(
		[]string{
			fmt.Sprintf("%s.%s", tableName, IdColumn),
			fmt.Sprintf("%s.%s", tableName, WorkflowDagResultIdColumn),
			fmt.Sprintf("%s.%s", tableName, ArtifactIdColumn),
			fmt.Sprintf("%s.%s", tableName, ContentPathColumn),
			fmt.Sprintf("%s.%s", tableName, StatusColumn),
			fmt.Sprintf("%s.%s", tableName, MetadataColumn),
			fmt.Sprintf("%s.%s", tableName, ExecStateColumn),
		},
		",",
	)
}
