package airflow

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
)

// TriggerWorkflow invokves a new Airflow DAGRun for `dag`.
func TriggerWorkflow(
	ctx context.Context,
	dag *models.DAG,
	vault vault.Vault,
) error {
	authConf, err := auth.ReadConfigFromSecret(
		ctx,
		dag.EngineConfig.AirflowConfig.ResourceID,
		vault,
	)
	if err != nil {
		return err
	}

	cli, err := newClient(ctx, authConf)
	if err != nil {
		return err
	}

	return cli.triggerDAGRun(dag.EngineConfig.AirflowConfig.DagID)
}
