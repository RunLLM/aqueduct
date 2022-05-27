package artifact_result

import "strings"

const (
	tableName = "artifact_result"

	// ArtifactResult table column names
	IdColumn                  = "id"
	WorkflowDagResultIdColumn = "workflow_dag_result_id"
	ArtifactIdColumn          = "artifact_id"
	ContentPathColumn         = "content_path"
	StatusColumn              = "status"
	MetadataColumn            = "metadata"
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
		},
		",",
	)
}
