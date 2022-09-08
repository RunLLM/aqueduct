package airflow

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
)

// TriggerWorkflow invokves a new Airflow DAGRun for `dag`.
func TriggerWorkflow(
	ctx context.Context,
	dbDag *workflow_dag.DBWorkflowDag,
	vault vault.Vault,
) error {
	authConf, err := auth.ReadConfigFromSecret(
		ctx,
		dbDag.EngineConfig.AirflowConfig.IntegrationId,
		vault,
	)
	if err != nil {
		return err
	}

	cli, err := newClient(ctx, authConf)
	if err != nil {
		return err
	}

	return cli.triggerDAGRun(dbDag.EngineConfig.AirflowConfig.DagId)
}
