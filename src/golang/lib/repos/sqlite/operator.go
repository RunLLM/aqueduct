package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type operatorRepo struct {
	operatorReader
	operatorWriter
}

type operatorReader struct{}

type operatorWriter struct{}

func NewOperatorRepo() repos.Operator {
	return &operatorRepo{
		operatorReader: operatorReader{},
		operatorWriter: operatorWriter{},
	}
}

func (*operatorReader) Exists(ctx context.Context, ID uuid.UUID, DB database.Database) (bool, error) {
	return IDExistsInTable(ctx, ID, models.OperatorTable, DB)
}

func (*operatorReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Operator, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM operator WHERE id = $1;`,
		models.OperatorCols(),
	)
	args := []interface{}{ID}

	return getOperator(ctx, DB, query, args...)
}

func (*operatorReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.Operator, error) {
	if len(IDs) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		`SELECT %s FROM operator WHERE id IN (%s);`,
		models.OperatorCols(),
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getOperators(ctx, DB, query, args...)
}

func (*operatorReader) GetByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) ([]models.Operator, error) {
	// Gets all operators that are a node with an incoming (id in `to_id`) or outgoing edge
	// (id in `from_id`) in the `workflow_dag_edge` for the specified DAG.
	query := fmt.Sprintf(
		`SELECT %s FROM operator WHERE id IN
		(
			SELECT from_id 
			FROM workflow_dag_edge 
			WHERE workflow_dag_id = $1 AND type = '%s' 
			UNION 
			SELECT to_id 
			FROM workflow_dag_edge 
			WHERE workflow_dag_id = $1 AND type = '%s'
		)`,
		models.OperatorCols(),
		shared.OperatorToArtifactDAGEdge,
		shared.ArtifactToOperatorDAGEdge,
	)
	args := []interface{}{dagID}

	return getOperators(ctx, DB, query, args...)
}

func (*operatorReader) GetDistinctLoadOPsByWorkflow(
	ctx context.Context,
	workflowID uuid.UUID,
	DB database.Database,
) ([]views.LoadOperator, error) {
	// Get all unique load operator (defined as a unique combination of operator name, integration,
	// and operator spec) that has an edge (in `from_id` or `to_id`) in a DAG
	// belonging to the specified workflow in order of when the operator was last modified.
	query := `
	SELECT
		operator.name AS operator_name, 
		workflow_dag.created_at AS modified_at,
		integration.name AS integration_name,
		CAST(json_extract(operator.spec, '$.load') AS BLOB) AS spec 	
	FROM 
		operator, integration, workflow_dag_edge, workflow_dag
	WHERE (
		json_extract(operator.spec, '$.type')='load' AND 
		integration.id = json_extract(operator.spec, '$.load.integration_id') AND
		( 
			workflow_dag_edge.from_id = operator.id OR 
			workflow_dag_edge.to_id = operator.id 
		) AND 
		workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
		workflow_dag.workflow_id = $1
	)
	GROUP BY
		operator.name,
		integration.name,
		json_extract(operator.spec, '$.load')	
	ORDER BY modified_at DESC;
	`
	args := []interface{}{workflowID}

	var operators []views.LoadOperator
	err := DB.Query(ctx, &operators, query, args...)
	return operators, err
}

func (*operatorReader) GetExtractAndLoadOPsByIntegration(
	ctx context.Context,
	integrationID uuid.UUID,
	DB database.Database,
) ([]models.Operator, error) {
	query := fmt.Sprintf(
		`SELECT %s 
		FROM operator
		WHERE 
			json_extract(spec, '$.load.integration_id') = $1
			OR json_extract(spec, '$.extract.integration_id') = $2`,
		models.OperatorCols(),
	)
	args := []interface{}{integrationID, integrationID}

	return getOperators(ctx, DB, query, args...)
}

// This currently only works with relational and S3 loads!
func (*operatorReader) GetLoadOPsByWorkflowAndIntegration(
	ctx context.Context,
	workflowID uuid.UUID,
	integrationID uuid.UUID,
	objectName string,
	DB database.Database,
) ([]models.Operator, error) {
	// Get all load operators where table=objectName & integration_id=integrationId
	// and has an edge (in `from_id` or `to_id`) in a DAG belonging to the specified
	// workflow.
	query := fmt.Sprintf(`
	SELECT %s
	FROM operator
	WHERE
		json_extract(spec, '$.type') = '%s' AND 
		(
			json_extract(spec, '$.load.parameters.table') = $1 OR
			json_extract(spec, '$.load.parameters.filepath') = $1
		) AND
		json_extract(spec, '$.load.integration_id') = $2 AND
		EXISTS 
		(
			SELECT 1 
			FROM 
				workflow_dag_edge, workflow_dag 
			WHERE 
			( 
				workflow_dag_edge.from_id = operator.id OR 
				workflow_dag_edge.to_id = operator.id 
			) AND 
			workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
			workflow_dag.workflow_id = $3
		);`,
		models.OperatorCols(),
		operator.LoadType,
	)
	args := []interface{}{objectName, integrationID, workflowID}

	return getOperators(ctx, DB, query, args...)
}

func (*operatorReader) GetLoadOPsByIntegration(
	ctx context.Context,
	integrationID uuid.UUID,
	objectName string,
	DB database.Database,
) ([]models.Operator, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM operator
		WHERE 
			json_extract(spec, '$.load.integration_id') = $1
			OR json_extract(spec, '$.extract.integration_id') = $2`,
		models.OperatorCols(),
	)
	args := []interface{}{integrationID, integrationID}

	return getOperators(ctx, DB, query, args...)
}

func (*operatorReader) GetLoadOPSpecsByOrg(ctx context.Context, orgID string, DB database.Database) ([]views.LoadOperatorSpec, error) {
	// Get the artifact id, artifact name, operator id, workflow name, workflow id,
	// and operator spec of all load operators (`to_id`s) and the artifact(s) going to
	// that operator (`from_id`s; these artifacts are the objects that will be saved
	// by the operator to the integration) in the workflows owned by the specified
	// organization.
	query := fmt.Sprintf(
		`SELECT DISTINCT 
			workflow_dag_edge.from_id AS artifact_id, 
			artifact.name AS artifact_name, 
		 	operator.id AS load_operator_id, 
			workflow.name AS workflow_name, 
			workflow.id AS workflow_id, operator.spec 
		 FROM 
		 	app_user, workflow, workflow_dag, 
			workflow_dag_edge, operator, artifact
		 WHERE 
		 	app_user.id = workflow.user_id 
			AND workflow.id = workflow_dag.workflow_id 
			AND workflow_dag.id = workflow_dag_edge.workflow_dag_id 
			AND workflow_dag_edge.to_id = operator.id 
			AND artifact.id = workflow_dag_edge.from_id 
			AND json_extract(operator.spec, '$.type') = '%s' 
			AND app_user.organization_id = $1;`,
		operator.LoadType,
	)
	args := []interface{}{orgID}

	var specs []views.LoadOperatorSpec
	err := DB.Query(ctx, &specs, query, args...)
	return specs, err
}

func (*operatorReader) GetRelationBatch(
	ctx context.Context,
	IDs []uuid.UUID,
	DB database.Database,
) ([]views.OperatorRelation, error) {
	// Given a list of `operatorIds`, find all workflow DAGs that has the id in the
	// `from_id` or `to_id` field.
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
		WHERE 
			workflow_dag_edge.workflow_dag_id = workflow_dag.id
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
		WHERE 
			workflow_dag_edge.workflow_dag_id = workflow_dag.id
			AND workflow.id = workflow_dag.workflow_id
			AND workflow_dag_edge.type = '%s'
			AND workflow_dag_edge.to_id IN (%s)
		`,
		shared.OperatorToArtifactDAGEdge,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
		shared.ArtifactToOperatorDAGEdge,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	var relations []views.OperatorRelation
	err := DB.Query(ctx, &relations, query, args...)
	return relations, err
}

func (*operatorReader) GetUnusedCondaEnvNames(ctx context.Context, DB database.Database) ([]string, error) {
	// Note that we use `OperatorToArtifactType` as the filtering condition because an operator
	// is guaranteed to generate at least one artifact, so this filter is guaranteed to capture
	// all operators involved in a workflow DAG.
	query := fmt.Sprintf(`
	WITH latest_workflow_dag AS
	(
		SELECT 
			workflow_dag.id 
		FROM
			workflow_dag 
		WHERE 
			created_at IN (
				SELECT 
					MAX(workflow_dag.created_at) 
				FROM 
					workflow, workflow_dag 
				WHERE 
					workflow.id = workflow_dag.workflow_id 
				GROUP BY 
					workflow.id
			)
	),
	all_env_names AS
	(
		SELECT DISTINCT
			json_extract(operator.spec, '$.engine_config.aqueduct_conda_config.env') AS name,
			operator.id as op_id
		FROM 
			workflow_dag_edge, operator
		WHERE
			workflow_dag_edge.type = '%s' 
			AND 
			workflow_dag_edge.from_id = operator.id
			AND
			json_extract(operator.spec, '$.engine_config.aqueduct_conda_config.env') IS NOT NULL
	),
	active_env_names AS
	(
		SELECT DISTINCT name
		FROM 
			all_env_names, latest_workflow_dag, workflow_dag_edge
		WHERE
			latest_workflow_dag.id = workflow_dag_edge.workflow_dag_id 
			AND 
			workflow_dag_edge.type = '%s' 
			AND 
			workflow_dag_edge.from_id = all_env_names.op_id
	),
	SELECT 
		all_env_names.name AS name
	FROM 
		all_env_names LEFT JOIN active_env_names 
		ON all_env_names.name = active_env_names.name
	WHERE 
		active_env_names.name IS NULL;`,
		shared.OperatorToArtifactDAGEdge,
		shared.OperatorToArtifactDAGEdge,
	)

	type resultStruct struct {
		Name string `db:"name"`
	}

	var resultRows []resultStruct
	err := DB.Query(ctx, &resultRows, query)
	if err != nil {
		return nil, err
	}

	results := make([]string, 0, len(resultRows))
	for _, row := range resultRows {
		results = append(results, row.Name)
	}

	return results, nil
}

func (*operatorReader) GetByEngineType(ctx context.Context, engineType shared.EngineType, DB database.Database) ([]models.Operator, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM operator WHERE json_extract(operator.spec, '$.engine_config.type') = $1;",
		models.OperatorCols(),
	)

	return getOperators(ctx, DB, query, engineType)
}

func (*operatorReader) ValidateOrg(ctx context.Context, ID uuid.UUID, orgID string, DB database.Database) (bool, error) {
	return validateNodeOwnership(ctx, orgID, ID, DB)
}

func (*operatorWriter) Create(
	ctx context.Context,
	name string,
	description string,
	spec *operator.Spec,
	executionEnvironmentID *uuid.UUID,
	DB database.Database,
) (*models.Operator, error) {
	cols := []string{
		models.OperatorID,
		models.OperatorName,
		models.OperatorDescription,
		models.OperatorSpec,
		models.OperatorExecutionEnvironmentID,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.OperatorTable, cols, models.OperatorCols())

	ID, err := GenerateUniqueUUID(ctx, models.OperatorTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		name,
		description,
		spec,
		executionEnvironmentID,
	}

	return getOperator(ctx, DB, query, args...)
}

func (*operatorWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	return deleteOperators(ctx, DB, []uuid.UUID{ID})
}

func (*operatorWriter) DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error {
	return deleteOperators(ctx, DB, IDs)
}

func (*operatorWriter) Update(
	ctx context.Context,
	ID uuid.UUID,
	changes map[string]interface{},
	DB database.Database,
) (*models.Operator, error) {
	var operator models.Operator
	err := repos.UpdateRecordToDest(
		ctx,
		&operator,
		changes,
		models.OperatorTable,
		models.OperatorID,
		ID,
		models.OperatorCols(),
		DB,
	)
	return &operator, err
}

func getOperators(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.Operator, error) {
	var operators []models.Operator
	err := DB.Query(ctx, &operators, query, args...)
	return operators, err
}

func getOperator(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.Operator, error) {
	operators, err := getOperators(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(operators) == 0 {
		return nil, database.ErrNoRows
	}

	if len(operators) != 1 {
		return nil, errors.Newf("Expected 1 Operator but got %v", len(operators))
	}

	return &operators[0], nil
}

func deleteOperators(ctx context.Context, DB database.Database, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		`DELETE FROM operator WHERE id IN (%s)`,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return DB.Execute(ctx, query, args...)
}
