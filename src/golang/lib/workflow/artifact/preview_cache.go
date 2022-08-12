package artifact

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

// PreviewCacheEntry is the object that a user of this cache will see when fetching.
type PreviewCacheEntry struct {
	ArtifactContentPath  string
	ArtifactMetadataPath string
	OpMetadataPath       string
}

type PreviewCacheManager interface {
	// Attempts to fetch the cache entries for the given artifacts, returning them
	// in a map for easy lookup. An entry that does not exist will not be present in the
	// returned map. Also returns a boolean indicating whether all artifacts were found in the
	// cache.
	GetMulti(ctx context.Context, logicalIDs []uuid.UUID) (bool, map[uuid.UUID]PreviewCacheEntry, error)

	// Writes the given entries into the cache. If entries already exist with the same artifact ID,
	// they will be deleted before the write takes place.
	Put(ctx context.Context, logicalID uuid.UUID, execPaths *utils.ExecPaths) error
}

type inMemoryPreviewCacheManagerImpl struct {
	cache *lru.Cache

	storageConfig *shared.StorageConfig
}

func deleteDataForEntry(ctx context.Context, storageConfig *shared.StorageConfig, val interface{}) {
	entry, ok := val.(PreviewCacheEntry)
	if !ok {
		log.Error("Preview Artifact Cache is storing an unexpected data structure. Cannot delete storage paths.")
		return
	}

	utils.CleanupStorageFile(ctx, storageConfig, entry.ArtifactContentPath)
	utils.CleanupStorageFile(ctx, storageConfig, entry.ArtifactMetadataPath)
	utils.CleanupStorageFile(ctx, storageConfig, entry.OpMetadataPath)
}

func (c *inMemoryPreviewCacheManagerImpl) GetMulti(_ context.Context, logicalIDs []uuid.UUID) (bool, map[uuid.UUID]PreviewCacheEntry, error) {
	allCached := true
	cachedEntries := make(map[uuid.UUID]PreviewCacheEntry, len(logicalIDs))
	for _, logicalID := range logicalIDs {
		entry, exists := c.cache.Get(logicalID)
		if exists {
			cachedEntries[logicalID] = entry.(PreviewCacheEntry)
		} else {
			allCached = false
		}
	}
	return allCached, cachedEntries, nil
}

func (c *inMemoryPreviewCacheManagerImpl) Put(ctx context.Context, logicalID uuid.UUID, execPaths *utils.ExecPaths) error {
	return c.putMulti(ctx, []uuid.UUID{logicalID}, []*utils.ExecPaths{execPaths})
}

func (c *inMemoryPreviewCacheManagerImpl) putMulti(ctx context.Context, logicalIDs []uuid.UUID, execPathsList []*utils.ExecPaths) error {
	for i, logicalID := range logicalIDs {
		// If the entry already exists, delete the data it points to, since the entry will be overridden.
		val, ok := c.cache.Peek(logicalID)
		if ok {
			deleteDataForEntry(ctx, c.storageConfig, val)
		}

		c.cache.Add(logicalID, PreviewCacheEntry{
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
) (PreviewCacheManager, error) {
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
