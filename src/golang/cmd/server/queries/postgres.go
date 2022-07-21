package queries

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type postgresReaderImpl struct {
	standardReaderImpl
}

func newPostgresReader() Reader {
	return &postgresReaderImpl{standardReaderImpl{}}
}

func (r *postgresReaderImpl) GetLoadOperatorSpecByOrganization(
	ctx context.Context,
	organizationId string,
	db database.Database,
) ([]LoadOperatorSpecResponse, error) {
	query := fmt.Sprintf(
		`SELECT DISTINCT workflow_dag_edge.from_id AS artifact_id, artifact.name AS artifact_name, operator.id AS load_operator_id, workflow.name AS workflow_name, workflow.id AS workflow_id, operator.spec 
		 FROM app_user, workflow, workflow_dag, workflow_dag_edge, operator, artifact
		 WHERE app_user.id = workflow.user_id AND workflow.id = workflow_dag.workflow_id
		 AND workflow_dag.id = workflow_dag_edge.workflow_dag_id AND workflow_dag_edge.to_id = operator.id
		 AND artifact.id = workflow_dag_edge.from_id
		 AND json_extract_path_text(operator.spec, 'type') = '%s'
		 AND app_user.organization_id = $1;`,
		operator.LoadType,
	)

	var response []LoadOperatorSpecResponse
	err := db.Query(ctx, &response, query, organizationId)
	return response, err
}

func (r *postgresReaderImpl) GetCheckResultsByArtifactIds(
	ctx context.Context,
	artifactIds []uuid.UUID,
	db database.Database,
) ([]ArtifactCheckResponse, error) {
	if len(artifactIds) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		`SELECT DISTINCT workflow_dag_edge.from_id AS artifact_id, operator.name AS name, operator_result.status, 
		 operator_result.metadata, operator_result.workflow_dag_result_id 
		 FROM workflow_dag_edge, operator, operator_result 
		 WHERE workflow_dag_edge.to_id = operator.id
		 AND operator.id = operator_result.operator_id
		 AND workflow_dag_edge.from_id IN (%s)
		 AND json_extract_path_text(operator.spec, 'type') = '%s';`,
		stmt_preparers.GenerateArgsList(len(artifactIds), 1),
		operator.CheckType,
	)

	args := stmt_preparers.CastIdsListToInterfaceList(artifactIds)

	var response []ArtifactCheckResponse
	err := db.Query(ctx, &response, query, args...)
	return response, err
}
