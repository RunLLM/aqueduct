package preview_cache

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

// Entry is the object that a cache-user will be fetching.
type Entry struct {
	ArtifactContentPath  string
	ArtifactMetadataPath string
	OpMetadataPath       string
}

type CacheManager interface {
	// Attempts to fetch the cache entry, keyed by an artifact's signature.
	// Along with the result, returns a boolean indicating whether this was a cache hit.
	Get(ctx context.Context, artifactSignature uuid.UUID) (bool, Entry, error)

	// Batch version of Get(). Returns a boolean indicating whether all keys had a cache hit
	// The cached results are returned in a map keyed by the artifact's signature.
	GetMulti(ctx context.Context, artifactSignatures []uuid.UUID) (bool, map[uuid.UUID]Entry, error)

	// Writes the given entries into the cache. If entries already exist with the same artifact ID,
	// they will be deleted before the write takes place.
	Put(ctx context.Context, artifactSignature uuid.UUID, execPaths *utils.ExecPaths) error
}

type inMemoryPreviewCacheManagerImpl struct {
	cache *lru.Cache

	storageConfig *shared.StorageConfig
}

func deleteDataForEntry(ctx context.Context, storageConfig *shared.StorageConfig, val interface{}) {
	entry, ok := val.(Entry)
	if !ok {
		log.Error("Preview Artifact Cache is storing an unexpected data structure. Cannot delete storage paths.")
		return
	}

	utils.CleanupStorageFile(ctx, storageConfig, entry.ArtifactContentPath)
	utils.CleanupStorageFile(ctx, storageConfig, entry.ArtifactMetadataPath)
	utils.CleanupStorageFile(ctx, storageConfig, entry.OpMetadataPath)
}

func (c *inMemoryPreviewCacheManagerImpl) Get(ctx context.Context, artifactSignature uuid.UUID) (bool, Entry, error) {
	allFound, entryByID, err := c.GetMulti(ctx, []uuid.UUID{artifactSignature})
	if err != nil {
		return false, Entry{}, err
	}

	if !allFound {
		return false, Entry{}, nil
	}
	return true, entryByID[artifactSignature], nil
}

func (c *inMemoryPreviewCacheManagerImpl) GetMulti(_ context.Context, artifactSignatures []uuid.UUID) (bool, map[uuid.UUID]Entry, error) {
	cachedEntries := make(map[uuid.UUID]Entry, len(artifactSignatures))
	for _, signature := range artifactSignatures {
		entry, exists := c.cache.Get(signature)
		if exists {
			cachedEntries[signature] = entry.(Entry)
		}
	}

	for _, signature := range artifactSignatures {
		if _, ok := cachedEntries[signature]; !ok {
			return false, cachedEntries, nil
		}
	}
	return true, cachedEntries, nil
}

func (c *inMemoryPreviewCacheManagerImpl) Put(ctx context.Context, artifactSignature uuid.UUID, execPaths *utils.ExecPaths) error {
	return c.putMulti(ctx, []uuid.UUID{artifactSignature}, []*utils.ExecPaths{execPaths})
}

func (c *inMemoryPreviewCacheManagerImpl) putMulti(ctx context.Context, artifactSignatures []uuid.UUID, execPathsList []*utils.ExecPaths) error {
	for i, signatures := range artifactSignatures {
		// If the entry already exists, delete the data it points to, since the entry will be overridden.
		val, ok := c.cache.Peek(signatures)
		if ok {
			deleteDataForEntry(ctx, c.storageConfig, val)
		}

		c.cache.Add(signatures, Entry{
			ArtifactContentPath:  execPathsList[i].ArtifactContentPath,
			ArtifactMetadataPath: execPathsList[i].ArtifactMetadataPath,
			OpMetadataPath:       execPathsList[i].OpMetadataPath,
		})
	}
	return nil
}

func NewInMemoryPreviewCacheManager(
	storageConfig *shared.StorageConfig,
	numEntries int,
) (CacheManager, error) {
	// Cleanup storage paths on eviction.
	cache, err := lru.NewWithEvict(numEntries, func(key interface{}, val interface{}) {
		ctx := context.Background()
		deleteDataForEntry(ctx, storageConfig, val)
	})
	if err != nil {
		return nil, err
	}

	return &inMemoryPreviewCacheManagerImpl{
		cache:         cache,
		storageConfig: storageConfig,
	}, nil
}
