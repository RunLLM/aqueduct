package utils

import (
	"github.com/google/uuid"
	"path/filepath"
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

func InitializeExecOutputPaths(isPreview bool) *ExecPaths {
	var pathPrefix string
	if isPreview {
		pathPrefix = previewDir
	}

	return &ExecPaths{
		OpMetadataPath:       filepath.Join(pathPrefix, uuid.New().String()),
		ArtifactContentPath:  filepath.Join(pathPrefix, uuid.New().String()),
		ArtifactMetadataPath: filepath.Join(pathPrefix, uuid.New().String()),
	}
}
