package utils

import "github.com/google/uuid"

// ExecPaths packages together all the storage paths that are written to by a python operator.
type ExecPaths struct {
	OpMetadataPath       string
	ArtifactContentPath  string
	ArtifactMetadataPath string
}

func InitializeExecOutputPaths() *ExecPaths {
	return &ExecPaths{
		OpMetadataPath:       uuid.New().String(),
		ArtifactContentPath:  uuid.New().String(),
		ArtifactMetadataPath: uuid.New().String(),
	}
}
