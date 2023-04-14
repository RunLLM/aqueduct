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
//   - integration credentials
//
// It also returns the names of all the keys that have been migrated to `newVault`.
// It is the responsibility of the caller to delete the keys if necessary.
func MigrateVault(
	ctx context.Context,
	oldVault vault.Vault,
	newVault vault.Vault,
	orgID string,
	integrationRepo repos.Integration,
	DB database.Database,
) ([]string, error) {
	integrations, err := integrationRepo.GetByOrg(ctx, orgID, DB)
	if err != nil {
		return nil, err
	}

	keys := []string{}

	log.Infof("There are %v integrations to migrate", len(integrations))

	// For each connected integration, migrate its credentials
	for _, integrationDB := range integrations {
		log.Infof("Starting migration for integration %v %v", integrationDB.ID, integrationDB.Name)
		// The vault key for the credentials is the integration record's ID
		key := integrationDB.ID.String()

		val, err := oldVault.Get(ctx, key)
		if err != nil {
			log.Errorf("Unable to get integration credentials %v from old vault at path %s: %v", integrationDB.ID, key, err)
			return nil, err
		}

		if err := newVault.Put(ctx, key, val); err != nil {
			log.Errorf("Unable to write integration credentials %v to new vault at path %s: %v", integrationDB.ID, key, err)
			return nil, err
		}

		keys = append(keys, key)
	}

	return keys, nil
}
