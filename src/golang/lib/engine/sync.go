package engine

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/airflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
)

// SyncSelfOrchestratedWorkflows syncs any workflow DAG results for any workflows running on a
// self-orchestrated engine.
// If orgID is empty, it syncs all workflows in the server. Otherwise, it syncs all workflows
// in the provided org.
func SyncSelfOrchestratedWorkflows(
	ctx context.Context,
	orgID string,
	artifactRepo repos.Artifact,
	artifactResultRepo repos.ArtifactResult,
	dagRepo repos.DAG,
	dagEdgeRepo repos.DAGEdge,
	dagResultRepo repos.DAGResult,
	operatorRepo repos.Operator,
	operatorResultRepo repos.OperatorResult,
	workflowRepo repos.Workflow,
	vaultObject vault.Vault,
	DB database.Database,
) error {
	// Sync workflows running on self-orchestrated engines
	airflowDagIDs, err := dagRepo.GetLatestIDsByOrgAndEngine(
		ctx,
		orgID,
		shared.AirflowEngineType,
		DB,
	)
	if err != nil {
		return err
	}

	return airflow.SyncDAGs(
		ctx,
		airflowDagIDs,
		workflowRepo,
		dagRepo,
		operatorRepo,
		artifactRepo,
		dagEdgeRepo,
		dagResultRepo,
		operatorResultRepo,
		artifactResultRepo,
		vaultObject,
		DB,
	)
}
