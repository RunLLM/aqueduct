package execution_environment

import (
	"context"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

const (
	CondaPathKey = "conda_path"
	ExecStateKey = "exec_state"
)

func ValidateCondaDevelop() error {
	// This is to ensure we can use `conda develop` to update the python path later on.
	args := []string{
		"develop",
		"--help",
	}
	_, _, err := lib_utils.RunCmd(CondaCmdPrefix, args, "", false)
	return err
}

// Returns the Conda path and std outputs, in addition to any errors.
func InitializeConda() (string, string, error) {
	out, _, err := lib_utils.RunCmd(CondaCmdPrefix, []string{"info", "--base"}, "", false)
	if err != nil {
		return "", out, errors.Wrap(err, "Failed to run Conda command.")
	}

	condaPath := strings.TrimSpace(out)

	err = createBaseEnvs()
	if err != nil {
		return condaPath, out, errors.Wrap(err, "Failed to create base Conda envs.")
	}
	return condaPath, out, nil
}

func GetCondaIntegration(
	ctx context.Context,
	userID uuid.UUID,
	integrationRepo repos.Integration,
	DB database.Database,
) (*models.Resource, error) {
	integrations, err := integrationRepo.GetByServiceAndUser(
		ctx,
		shared.Conda,
		userID,
		DB,
	)
	if err != nil {
		return nil, err
	}

	if len(integrations) == 0 {
		return nil, nil
	}

	return &integrations[0], nil
}
