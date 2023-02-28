package execution_environment

import (
	"context"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/execution_state"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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
	_, _, err := lib_utils.RunCmd(CondaCmdPrefix, args...)
	return err
}

func InitializeConda(
	ctx context.Context,
	integrationID uuid.UUID,
	integrationRepo repos.Integration,
	DB database.Database,
) {
	now := time.Now()
	_, err := integrationRepo.Update(
		ctx,
		integrationID,
		map[string]interface{}{
			models.IntegrationConfig: (*shared.IntegrationConfig)(&map[string]string{
				ExecStateKey: execution_state.SerializedRunning(&now),
			}),
		},
		DB,
	)
	if err != nil {
		log.Errorf("Failed to update conda integration: %v", err)
		return
	}

	out, _, err := lib_utils.RunCmd(CondaCmdPrefix, "info", "--base")
	if err != nil {
		integrationConfig := (*shared.IntegrationConfig)(&map[string]string{
			CondaPathKey: "",
		})

		execution_state.UpdateOnFailure(
			ctx,
			out,
			err.Error(),
			integrationConfig,
			&now,
			integrationID,
			integrationRepo,
			DB,
		)

		return
	}

	condaPath := strings.TrimSpace(out)

	err = createBaseEnvs()
	if err != nil {
		integrationConfig := (*shared.IntegrationConfig)(&map[string]string{
			CondaPathKey: condaPath,
		})
		execution_state.UpdateOnFailure(
			ctx,
			out,
			err.Error(),
			integrationConfig,
			&now,
			integrationID,
			integrationRepo,
			DB,
		)

		return
	}

	_, err = integrationRepo.Update(
		ctx,
		integrationID,
		map[string]interface{}{
			models.IntegrationConfig: (*shared.IntegrationConfig)(&map[string]string{
				CondaPathKey: condaPath,
				ExecStateKey: execution_state.SerializedSuccess(&now),
			}),
		},
		DB,
	)

	if err != nil {
		log.Errorf("Failed to update conda integration: ")
	}
}

func GetCondaIntegration(
	ctx context.Context,
	userID uuid.UUID,
	integrationRepo repos.Integration,
	DB database.Database,
) (*models.Integration, error) {
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

