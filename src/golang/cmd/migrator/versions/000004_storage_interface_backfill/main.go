package _000004_storage_interface_backfill

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

func Up(ctx context.Context, db database.Database) error {
	s3Configs, err := getAllS3Configs(ctx, db)
	if err != nil {
		return err
	}

	for _, s3ConfigResponse := range s3Configs {
		storageConfig := storageConfig{
			Type:     s3StorageType,
			S3Config: &s3ConfigResponse.S3Config,
		}

		err = updateStorageConfig(ctx, s3ConfigResponse.Id, storageConfig, db)
		if err != nil {
			return err
		}
	}

	operatorSpecs, err := getAllOperatorSpecs(ctx, db)
	if err != nil {
		return err
	}

	for _, operatorSpecResponse := range operatorSpecs {
		updated := false
		if operatorSpecResponse.Spec.isFunction() {
			operatorSpecResponse.Spec.Function().StoragePath = operatorSpecResponse.Spec.Function().S3Path
			updated = true
		} else if operatorSpecResponse.Spec.isMetric() {
			operatorSpecResponse.Spec.Metric().Function.StoragePath = operatorSpecResponse.Spec.Metric().Function.S3Path
			updated = true
		} else if operatorSpecResponse.Spec.isValidation() {
			operatorSpecResponse.Spec.Validation().Function.StoragePath = operatorSpecResponse.Spec.Validation().Function.S3Path
			updated = true
		}

		if updated {
			err = updateOperatorSpec(ctx, operatorSpecResponse.Id, operatorSpecResponse.Spec, db)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Down(ctx context.Context, db database.Database) error {
	err := setStorageConfigToNull(ctx, db)
	if err != nil {
		return err
	}

	operatorSpecs, err := getAllOperatorSpecs(ctx, db)
	if err != nil {
		return err
	}

	for _, operatorSpecResponse := range operatorSpecs {
		updated := false
		if operatorSpecResponse.Spec.isFunction() {
			operatorSpecResponse.Spec.Function().StoragePath = ""
			updated = true
		} else if operatorSpecResponse.Spec.isMetric() {
			operatorSpecResponse.Spec.Metric().Function.StoragePath = ""
			updated = true
		} else if operatorSpecResponse.Spec.isValidation() {
			operatorSpecResponse.Spec.Validation().Function.StoragePath = ""
			updated = true
		}

		if updated {
			err = updateOperatorSpec(ctx, operatorSpecResponse.Id, operatorSpecResponse.Spec, db)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
