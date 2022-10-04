package utils

import (
	"path/filepath"

	"github.com/google/uuid"
)

// The subdirectory within the storage directory containing all the outputs of previewed operators.
// The contents of this directory will be dropped on every server start.
const previewDir = "preview"

// ExecPaths packages together all the storage paths that are written to by a python operator.
type ExecPaths struct {
	OpMetadataPath       string
	ArtifactContentPath  string
	ArtifactMetadataPath string
}

func InitializeExecOutputPaths(isPreview bool, opMetadataPath string) *ExecPaths {
	return &ExecPaths{
		OpMetadataPath:       opMetadataPath,
		ArtifactContentPath:  InitializePath(isPreview),
		ArtifactMetadataPath: InitializePath(isPreview),
	}
}

func InitializePath(isPreview bool) string {
	var pathPrefix string
	if isPreview {
		pathPrefix = previewDir
	}
	return filepath.Join(pathPrefix, uuid.New().String())
}
