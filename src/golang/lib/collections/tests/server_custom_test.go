package tests

import (
	"context"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact/table"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"

	"github.com/aqueducthq/aqueduct/cmd/server/queries"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetLoadOperatorSpecByOrganization(t *testing.T) {
	defer resetDatabase(t)

	dags := seedWorkflowDag(t, 1)
	testDag := dags[0]

	testWorkflow, err := readers.workflowReader.GetWorkflow(context.Background(), testDag.WorkflowId, db)
	require.Nil(t, err)

	testUser, err := readers.userReader.GetUser(context.Background(), testWorkflow.UserId, db)
	require.Nil(t, err)

	testArtifact, err := writers.artifactWriter.CreateArtifact(
		context.Background(),
		randString(5),
		randString(10),
		artifact.NewSpecFromTable(table.Table{}),
		db,
	)
	require.Nil(t, err)

	testOps := seedOperatorWithSpecs(t, 1, []operator.Spec{
		*operator.NewSpecFromLoad(connector.Load{
			Service:       integration.Postgres,
			IntegrationId: uuid.New(),
			Parameters:    &connector.PostgresLoadParams{connector.RelationalDBLoadParams{Table: "test"}},
		}),
	})

	loadOp := testOps[0]
	require.True(t, loadOp.Spec.IsLoad())

	seedWorkflowDagEdgeWithDagId(
		t,
		map[uuid.UUID]uuid.UUID{
			testArtifact.Id: loadOp.Id,
		},
		testDag.Id,
	)

	expectedResponse := queries.LoadOperatorSpecResponse{
		ArtifactId:     testArtifact.Id,
		ArtifactName:   testArtifact.Name,
		LoadOperatorId: loadOp.Id,
		WorkflowName:   testWorkflow.Name,
		WorkflowId:     testWorkflow.Id,
		Spec:           loadOp.Spec,
	}

	loadOpSpecResp, err := readers.serverReader.GetLoadOperatorSpecByOrganization(context.Background(), testUser.OrganizationId, db)
	require.Nil(t, err)
	require.Len(t, loadOpSpecResp, 1)

	requireDeepEqual(t, expectedResponse, loadOpSpecResp[0])
}

func TestGetLatestWorkflowDagIdsByOrganizationIdAndEngine(t *testing.T) {
	defer resetDatabase(t)

	testWorkflows := seedWorkflow(t, 2)
	require.Len(t, testWorkflows, 2)

	testWorkflow1, testWorkflow2 := testWorkflows[0], testWorkflows[1]

	// Create 2 workflow dags with a default Aqueduct engine
	seedWorkflowDagWithWorkflows(t, 1, []uuid.UUID{testWorkflow1.Id})
	seedWorkflowDagWithWorkflows(t, 1, []uuid.UUID{testWorkflow2.Id})

	// Create another workflow dag for testWorkflow1 using an Airflow engine
	airflowTestDag1, err := writers.workflowDagWriter.CreateWorkflowDag(
		context.Background(),
		testWorkflow1.Id,
		&shared.StorageConfig{
			Type: shared.FileStorageType,
		},
		&shared.EngineConfig{
			Type:          shared.AirflowEngineType,
			AirflowConfig: &shared.AirflowConfig{},
		},
		db,
	)
	require.Nil(t, err)

	airflowDagIds, err := readers.serverReader.GetLatestWorkflowDagIdsByOrganizationIdAndEngine(
		context.Background(),
		testOrganizationId,
		shared.AirflowEngineType,
		db,
	)
	require.Nil(t, err)

	require.Len(t, airflowDagIds, 1)
	require.Equal(t, airflowTestDag1.Id, airflowDagIds[0].Id)
}
