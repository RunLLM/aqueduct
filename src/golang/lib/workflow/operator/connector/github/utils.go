package github

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/function"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/google/uuid"
)

func ErrGithubMetadataMissing() error {
	return errors.New("Github metadata is missing for a github function.")
}

func IsFunctionFromGithub(spec *function.Function) (bool, error) {
	if spec.Type != function.GithubFunctionType {
		return false, nil
	}

	if spec.GithubMetadata == nil {
		return true, ErrGithubMetadataMissing()
	}

	return true, nil
}

func IsExtractFromGithub(spec *connector.Extract) bool {
	relSpec, ok := connector.CastToRelationalDBExtractParams(spec.Parameters)
	if !ok {
		return false
	}

	return relSpec.GithubMetadata != nil
}

// Perform 'background' update to bring the spec to the latest version.
// For now, this method update any github spec to the latest commit, together with any
// github content (storage path for function files etc.)
//
// If storageConfig is provided, it uploads content to storage config if possible.
func PullOperator(
	ctx context.Context,
	client Client,
	spec *operator.Spec,
	storageConfig *shared.StorageConfig,
) (bool, error) {
	if spec.IsExtract() {
		if !IsExtractFromGithub(spec.Extract()) {
			return false, nil
		}
		updated, err := client.PullExtract(ctx, spec.Extract())
		return updated, err
	}

	if !spec.HasFunction() {
		return false, nil
	}

	fn := spec.Function()
	isGhFunction, err := IsFunctionFromGithub(fn)
	if err != nil {
		return false, err
	}

	if !isGhFunction {
		return false, nil
	}

	updated, zipball, err := client.PullAndUpdateFunction(
		ctx,
		fn,
		false, /* always extract */
	)
	if !updated || err != nil {
		return false, err
	}

	if storageConfig == nil {
		return true, errors.New("Invalid Storage Config")
	}

	storagePath := uuid.New().String()
	err = storage.NewStorage(storageConfig).Put(ctx, storagePath, zipball)
	if err != nil {
		return true, err
	}

	fn.StoragePath = storagePath

	return true, nil
}
