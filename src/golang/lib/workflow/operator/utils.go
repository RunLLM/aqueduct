package operator

import (
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
)

var (
	errWrongNumInputs  = errors.New("Wrong number of operator inputs")
	errWrongNumOutputs = errors.New("Wrong number of operator outputs")
)

// Returns list of content paths and metadata paths, in that order.
func unzipExecPathsToRawPaths(execPaths []*utils.ExecPaths) ([]string, []string) {
	contentPaths := make([]string, 0, len(execPaths))
	metadataPaths := make([]string, 0, len(execPaths))
	for _, ep := range execPaths {
		contentPaths = append(contentPaths, ep.ArtifactContentPath)
		metadataPaths = append(metadataPaths, ep.ArtifactMetadataPath)
	}
	return contentPaths, metadataPaths
}
