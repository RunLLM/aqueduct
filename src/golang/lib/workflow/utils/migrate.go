package utils

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/vault"
)

// MigrateVault migrates all vault content from `oldVault` to `newVault`.
// This includes:
//   - integration credentials
//
// If the migration is successful, the above content is deleted from `oldVault`.
func MigrateVault(
	ctx context.Context,
	oldVault vault.Vault,
	newVault vault.Vault,
	orgID string,
	integrationReader integration.Reader,
	db database.Database,
) error {
	integrations, err := integrationReader.GetIntegrationsByOrganization(ctx, orgID, db)
	if err != nil {
		return err
	}

	// For each connected integration, migrate its credentials
	for _, integrationDB := range integrations {
		// The vault key for the credentials is the integration record's ID
		key := integrationDB.Id.String()

		val, err := oldVault.Get(ctx, key)
		if err != nil {
			return err
		}

		if err := newVault.Put(ctx, key, val); err != nil {
			return err
		}
	}

	return nil
}
