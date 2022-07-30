package artifact_result

import "strings"

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
