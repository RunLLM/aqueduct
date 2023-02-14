package preview_cache

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func requirePathsDoNotExist(t *testing.T, execPaths *utils.ExecPaths, errMsgTemplate string) {
	if _, err := os.Stat(execPaths.OpMetadataPath); err == nil {
		require.Fail(t, fmt.Sprintf(errMsgTemplate, execPaths.OpMetadataPath))
	}
	if _, err := os.Stat(execPaths.ArtifactMetadataPath); err == nil {
		require.Fail(t, fmt.Sprintf(errMsgTemplate, execPaths.ArtifactMetadataPath))
	}
	if _, err := os.Stat(execPaths.ArtifactContentPath); err == nil {
		require.Fail(t, fmt.Sprintf(errMsgTemplate, execPaths.ArtifactContentPath))
	}
}

func requirePathsDoExist(t *testing.T, execPaths *utils.ExecPaths, errMsgTemplate string) {
	if _, err := os.Stat(execPaths.OpMetadataPath); err != nil {
		require.Fail(t, fmt.Sprintf(errMsgTemplate, execPaths.OpMetadataPath))
	}
	if _, err := os.Stat(execPaths.ArtifactMetadataPath); err != nil {
		require.Fail(t, fmt.Sprintf(errMsgTemplate, execPaths.ArtifactMetadataPath))
	}
	if _, err := os.Stat(execPaths.ArtifactContentPath); err != nil {
		require.Fail(t, fmt.Sprintf(errMsgTemplate, execPaths.ArtifactContentPath))
	}
}

func writePathsToFilesystem(t *testing.T, execPaths *utils.ExecPaths) {
	f, err := os.Create(execPaths.OpMetadataPath)
	_ = f.Close()
	require.Nil(t, err)

	f, err = os.Create(execPaths.ArtifactContentPath)
	_ = f.Close()
	require.Nil(t, err)

	f, err = os.Create(execPaths.ArtifactMetadataPath)
	_ = f.Close()
	require.Nil(t, err)
}

func removePathsToFilesystem(t *testing.T, execPaths *utils.ExecPaths) {
	err := os.Remove(execPaths.OpMetadataPath)
	require.Nil(t, err)

	err = os.Remove(execPaths.ArtifactContentPath)
	require.Nil(t, err)

	err = os.Remove(execPaths.ArtifactMetadataPath)
	require.Nil(t, err)
}

// Returns a storage config struct pointing to the directory this test lives in.
func storageConfigForCurrentDirectory(t *testing.T) *shared.StorageConfig {
	wd, err := os.Getwd()
	require.Nil(t, err)
	return &shared.StorageConfig{
		Type: shared.FileStorageType,
		FileConfig: &shared.FileConfig{
			wd,
		},
	}
}

func TestPreviewCacheCollision(t *testing.T) {
	ctx := context.Background()

	key := uuid.New()
	execPaths := &utils.ExecPaths{
		"op_metadata_path",
		"artifact_content_path",
		"artifact_metadata_path",
	}
	requirePathsDoNotExist(t, execPaths, "%s already exists! You should remove this and retry.")

	// Create the data in the same directory as this test, to be overwritten.
	writePathsToFilesystem(t, execPaths)

	cache, err := NewInMemoryPreviewCacheManager(
		storageConfigForCurrentDirectory(t),
		5,
	)
	require.Nil(t, err)

	err = cache.Put(ctx, key, execPaths)
	require.Nil(t, err)

	found, entry, err := cache.Get(ctx, key)
	require.Nil(t, err)
	require.True(t, found)
	require.Equal(t, Entry{
		OpMetadataPath:       execPaths.OpMetadataPath,
		ArtifactMetadataPath: execPaths.ArtifactMetadataPath,
		ArtifactContentPath:  execPaths.ArtifactContentPath,
	}, entry)

	// Write the same key to the path. This will error, since the entry is not the same.
	// Nothing will be cleaned up.
	newExecPaths := &utils.ExecPaths{
		"op_metadata_path2",
		"artifact_content_path2",
		"artifact_metadata_path2",
	}
	err = cache.Put(ctx, key, newExecPaths)
	require.Error(t, err, "we expect the entry to be the same")

	requirePathsDoExist(t, execPaths, "%s should continue to exist.")
	removePathsToFilesystem(t, execPaths)
	requirePathsDoNotExist(t, execPaths, "%s should have been removed.")
}

func TestPreviewCacheEviction(t *testing.T) {
	ctx := context.Background()

	execPaths := &utils.ExecPaths{
		"op_metadata_path_to_evict",
		"artifact_content_path_to_evict",
		"artifact_metadata_path_to_evict",
	}
	requirePathsDoNotExist(t, execPaths, "%s already exists! You should remove this and retry.")

	// Create the data to be deleted when the entry is evicted.
	writePathsToFilesystem(t, execPaths)

	cache, err := NewInMemoryPreviewCacheManager(
		storageConfigForCurrentDirectory(t),
		1,
	)
	require.Nil(t, err)

	err = cache.Put(ctx, uuid.New(), execPaths)
	require.Nil(t, err)

	// Add a new entry with a different id. Because the cache size is 1, the previous entry will be forcably
	// evicted, and its filesystem data deleted.
	newExecPaths := &utils.ExecPaths{
		"op_metadata_path2",
		"artifact_content_path2",
		"artifact_metadata_path2",
	}
	err = cache.Put(ctx, uuid.New(), newExecPaths)
	require.Nil(t, err)

	requirePathsDoNotExist(t, execPaths, "%s should not exist after eviction.")
}
