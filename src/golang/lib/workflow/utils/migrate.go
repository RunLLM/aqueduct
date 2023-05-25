package utils

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	log "github.com/sirupsen/logrus"
)

// MigrateVault migrates all vault content from `oldVault` to `newVault`.
// This includes:
//   - resource credentials
//
// It also returns the names of all the keys that have been migrated to `newVault`.
// It is the responsibility of the caller to delete the keys if necessary.
func MigrateVault(
	ctx context.Context,
	oldVault vault.Vault,
	newVault vault.Vault,
	orgID string,
	resourceRepo repos.Resource,
	DB database.Database,
) ([]string, error) {
	resources, err := resourceRepo.GetByOrg(ctx, orgID, DB)
	if err != nil {
		return nil, err
	}

	keys := []string{}

	log.Infof("There are %v resources to migrate", len(resources))

	// For each connected resource, migrate its credentials
	for _, resourceDB := range resources {
		log.Infof("Starting migration for resource %v %v", resourceDB.ID, resourceDB.Name)
		// The vault key for the credentials is the resource record's ID
		key := resourceDB.ID.String()

		val, err := oldVault.Get(ctx, key)
		if err != nil {
			log.Errorf("Unable to get resource credentials %v from old vault at path %s: %v", resourceDB.ID, key, err)
			return nil, err
		}

		if err := newVault.Put(ctx, key, val); err != nil {
			log.Errorf("Unable to write resource credentials %v to new vault at path %s: %v", resourceDB.ID, key, err)
			return nil, err
		}

		keys = append(keys, key)
	}

	return keys, nil
}
