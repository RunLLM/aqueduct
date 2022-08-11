package queries

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

func (r *standardReaderImpl) GetLatestWorkflowDagIdsByOrganizationId(
	ctx context.Context,
	organizationId string,
	db database.Database,
) ([]WorkflowDagId, error) {
	query := `
		 SELECT workflow_dag.id FROM workflow_dag WHERE created_at IN (
		 SELECT MAX(workflow_dag.created_at) FROM app_user, workflow, workflow_dag 
		 WHERE app_user.id = workflow.user_id AND workflow.id = workflow_dag.workflow_id AND 
		 app_user.organization_id = $1 
		 GROUP BY workflow.id);`

	var workflowDags []WorkflowDagId
	err := db.Query(ctx, &workflowDags, query, organizationId)
	return workflowDags, err
}

func (r *standardReaderImpl) GetArtifactIdsFromWorkflowDagIdsAndDownstreamOperatorIds(
	ctx context.Context,
	operatorIds []uuid.UUID,
	workflowDagIds []uuid.UUID,
	db database.Database,
) ([]ArtifactId, error) {
	if len(workflowDagIds) == 0 || len(operatorIds) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		`SELECT DISTINCT from_id AS artifact_id FROM workflow_dag_edge WHERE workflow_dag_id IN (%s) 
		 AND to_id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(workflowDagIds), 1),
		stmt_preparers.GenerateArgsList(len(operatorIds), len(workflowDagIds)+1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(workflowDagIds)
	args = append(args, stmt_preparers.CastIdsListToInterfaceList(operatorIds)...)

	var result []ArtifactId
	err := db.Query(ctx, &result, query, args...)
	return result, err
}

func (r *standardReaderImpl) GetArtifactResultsByArtifactIds(
	ctx context.Context,
	artifactIds []uuid.UUID,
	db database.Database,
) ([]ArtifactResponse, error) {
	if len(artifactIds) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		`SELECT artifact_result.artifact_id, artifact_result.workflow_dag_result_id, artifact_result.status, workflow_dag_result.created_at AS timestamp 
		 FROM artifact_result, workflow_dag_result 
		 WHERE artifact_result.workflow_dag_result_id = workflow_dag_result.id AND 
		 artifact_result.artifact_id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(artifactIds), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(artifactIds)

	var response []ArtifactResponse
	err := db.Query(ctx, &response, query, args...)
	return response, err
}

func (r *standardReaderImpl) GetOperatorResultsByArtifactIdsAndWorkflowDagResultIds(
	ctx context.Context,
	artifactIds, workflowDagResultIds []uuid.UUID,
	db database.Database,
) ([]ArtifactOperatorResponse, error) {
	if len(artifactIds) == 0 || len(workflowDagResultIds) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		`SELECT DISTINCT
			workflow_dag_edge.to_id AS artifact_id,
			operator_result.execution_state as metadata,
			operator_result.workflow_dag_result_id  
		 FROM workflow_dag_edge, operator_result 
		 WHERE workflow_dag_edge.from_id = operator_result.operator_id AND workflow_dag_edge.to_id IN (%s) AND 
		 operator_result.workflow_dag_result_id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(artifactIds), 1),
		stmt_preparers.GenerateArgsList(len(workflowDagResultIds), len(artifactIds)+1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(artifactIds)
	args = append(args, stmt_preparers.CastIdsListToInterfaceList(workflowDagResultIds)...)

	var response []ArtifactOperatorResponse
	err := db.Query(ctx, &response, query, args...)
	return response, err
}

func (r *standardReaderImpl) GetWorkflowIdsFromOperatorIds(
	ctx context.Context,
	operatorIds []uuid.UUID,
	db database.Database,
) ([]WorkflowIdsFromOperatorIdsResponse, error) {
	// This query looks up all operators with at least one upstream
	query := fmt.Sprintf(
		`
			SELECT
				workflow.id as workflow_id,
				workflow_dag.id as workflow_dag_id,
				workflow_dag_edge.from_id as operator_id
			FROM
				workflow,
				workflow_dag,
				workflow_dag_edge 
			WHERE workflow_dag_edge.workflow_dag_id = workflow_dag.id
			AND workflow.id = workflow_dag.workflow_id
			AND workflow_dag_edge.type = '%s'
			AND workflow_dag_edge.from_id IN (%s)
		UNION
			SELECT
				workflow.id as workflow_id,
				workflow_dag.id as workflow_dag_id,
				workflow_dag_edge.to_id as operator_id
			FROM
				workflow,
				workflow_dag,
				workflow_dag_edge 
			WHERE workflow_dag_edge.workflow_dag_id = workflow_dag.id
			AND workflow.id = workflow_dag.workflow_id
			AND workflow_dag_edge.type = '%s'
			AND workflow_dag_edge.to_id IN (%s)
		`,
		workflow_dag_edge.OperatorToArtifactType,
		stmt_preparers.GenerateArgsList(len(operatorIds), 1),
		workflow_dag_edge.ArtifactToOperatorType,
		stmt_preparers.GenerateArgsList(len(operatorIds), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(operatorIds)

	var results []WorkflowIdsFromOperatorIdsResponse

	err := db.Query(ctx, &results, query, args...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *standardReaderImpl) GetLatestWorkflowDagIdsFromWorkflowIds(
	ctx context.Context,
	workflowIds []uuid.UUID,
	db database.Database,
) (map[uuid.UUID]uuid.UUID, error) {
	getLatestWorkflowDagIdsQuery := fmt.Sprintf(
		`
		SELECT workflow_dag_id, workflow_id FROM (
			SELECT
				id as workflow_dag_id,
				workflow_id,
				MAX(created_at) as created_at
			FROM workflow_dag
			WHERE workflow_id IN (%s)
			GROUP BY workflow_id
		)`,
		stmt_preparers.GenerateArgsList(len(workflowIds), 1),
	)

	var ids []struct {
		WorkflowDagId uuid.UUID `db:"workflow_dag_id" json:"workflow_dag_id"`
		WorkflowId    uuid.UUID `db:"workflow_id" json:"workflow_id"`
	}

	args := stmt_preparers.CastIdsListToInterfaceList(workflowIds)
	err := db.Query(ctx, &ids, getLatestWorkflowDagIdsQuery, args...)
	if err != nil {
		return nil, err
	}

	results := make(map[uuid.UUID]uuid.UUID, len(ids))
	for _, item := range ids {
		results[item.WorkflowId] = item.WorkflowDagId
	}

	return results, nil
}
