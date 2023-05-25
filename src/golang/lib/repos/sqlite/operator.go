package sqlite

import (
	"context"
	"fmt"
	"strings"

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

const operatorNodeViewSubQuery = `
	WITH op_with_outputs AS ( -- Aggregate outputs
		SELECT
			operator.id AS id,
			workflow_dag.id AS dag_id,
			operator.name AS name,
			operator.description AS description,
			operator.spec AS spec,
			operator.execution_environment_id AS execution_environment_id,
			CAST( json_group_array( -- Group to_ids and idx into one array
				json_object(
					'value', workflow_dag_edge.to_id,
					'idx', workflow_dag_edge.idx
				)
			) AS BLOB) AS outputs
		FROM
			operator, workflow_dag, workflow_dag_edge
		WHERE
			workflow_dag.id = workflow_dag_edge.workflow_dag_id
			AND operator.id = workflow_dag_edge.from_id
		GROUP BY
			workflow_dag.id, operator.id
	),
	op_with_inputs AS ( -- Aggregate inputs
		SELECT
			operator.id AS id,
			workflow_dag.id AS dag_id,
			operator.name AS name,
			operator.description AS description,
			operator.spec AS spec,
			operator.execution_environment_id AS execution_environment_id,
			CAST( json_group_array( -- Group from_ids and idx into one array
				json_object(
					'value', workflow_dag_edge.from_id,
					'idx', workflow_dag_edge.idx
				)
			) AS BLOB) AS inputs
		FROM
			operator, workflow_dag, workflow_dag_edge
		WHERE
			workflow_dag.id = workflow_dag_edge.workflow_dag_id
			AND operator.id = workflow_dag_edge.to_id
		GROUP BY
			workflow_dag.id, operator.id
	)
	SELECT -- A full outer join to include operators without inputs / outputs.
		op_with_outputs.id AS id,
		op_with_outputs.dag_id AS dag_id,
		op_with_outputs.name AS name,
		op_with_outputs.description AS description,
		op_with_outputs.spec AS spec,
		op_with_outputs.execution_environment_id AS execution_environment_id,
		op_with_outputs.outputs AS outputs,
		op_with_inputs.inputs AS inputs
	FROM
		op_with_outputs LEFT JOIN op_with_inputs
	ON
		op_with_outputs.id = op_with_inputs.id
		AND op_with_outputs.dag_id = op_with_inputs.dag_id
	UNION ALL
	SELECT
		op_with_inputs.id AS id,
		op_with_inputs.dag_id AS dag_id,
		op_with_inputs.name AS name,
		op_with_inputs.description AS description,
		op_with_inputs.spec AS spec,
		op_with_inputs.execution_environment_id AS execution_environment_id,
		op_with_outputs.outputs AS outputs,
		op_with_inputs.inputs AS inputs
	FROM
		op_with_inputs LEFT JOIN op_with_outputs
	ON
		op_with_outputs.id = op_with_inputs.id
		AND op_with_outputs.dag_id = op_with_inputs.dag_id
	WHERE op_with_outputs.outputs IS NULL
`

var mergedNodeViewSubQuery = fmt.Sprintf(`
	WITH
		operator_node AS (%s), 
		artifact_node AS (%s)
	SELECT 
		operator_node.id AS id,
		operator_node.name AS name,
		operator_node.description AS description,
		operator_node.spec AS spec,
		operator_node.execution_environment_id AS execution_environment_id,
		operator_node.dag_id AS dag_id,
		operator_node.inputs AS inputs,
		artifact_node.id AS artifact_id,
		artifact_node.type AS type,
		artifact_node.outputs AS outputs
	FROM 
		operator_node LEFT JOIN 
		artifact_node 
	ON
		artifact_node.input = operator_node.id
`,
	operatorNodeViewSubQuery,
	artifactNodeViewSubQuery,
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

func (r *operatorReader) GetNode(ctx context.Context, ID uuid.UUID, DB database.Database) (*views.OperatorNode, error) {
	nodes, err := r.GetNodeBatch(ctx, []uuid.UUID{ID}, DB)
	if err != nil {
		return nil, err
	}
	return &nodes[0], nil
}

func (*operatorReader) GetNodeBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]views.OperatorNode, error) {
	if len(IDs) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		"WITH %s AS (%s) SELECT %s FROM %s WHERE %s IN (%s)",
		views.OperatorNodeView,
		operatorNodeViewSubQuery,
		views.OperatorNodeCols(),
		views.OperatorNodeView,
		models.OperatorID,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)
	return getOperatorNodes(ctx, DB, query, args...)
}

func (r *operatorReader) GetOperatorWithArtifactNode(ctx context.Context, ID uuid.UUID, DB database.Database) (*views.OperatorWithArtifactNode, error) {
	nodes, err := r.GetOperatorWithArtifactNodeBatch(ctx, []uuid.UUID{ID}, DB)
	if err != nil {
		return nil, err
	}
	return &nodes[0], nil
}

func (*operatorReader) GetOperatorWithArtifactNodeBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]views.OperatorWithArtifactNode, error) {
	if len(IDs) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		"WITH %s AS (%s) SELECT %s FROM %s WHERE %s IN (%s)",
		views.OperatorWithArtifactNodeView,
		mergedNodeViewSubQuery,
		views.OperatorWithArtifactNodeCols(),
		views.OperatorWithArtifactNodeView,
		views.OperatorWithArtifactNodeID,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)
	return getOperatorWithArtifactNodes(ctx, DB, query, args...)
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

func (*operatorReader) GetNodesByDAG(
	ctx context.Context,
	dagID uuid.UUID,
	DB database.Database,
) ([]views.OperatorNode, error) {
	query := fmt.Sprintf(
		"WITH %s AS (%s) SELECT %s FROM %s WHERE %s = $1",
		views.OperatorNodeView,
		operatorNodeViewSubQuery,
		views.OperatorNodeCols(),
		views.OperatorNodeView,
		views.OperatorNodeDagID,
	)
	args := []interface{}{dagID}
	return getOperatorNodes(ctx, DB, query, args...)
}

func (*operatorReader) GetDistinctLoadOPsByWorkflow(
	ctx context.Context,
	workflowID uuid.UUID,
	DB database.Database,
) ([]views.LoadOperator, error) {
	// Get all unique load operator (defined as a unique combination of operator name, resource,
	// and operator spec) that has an edge (in `from_id` or `to_id`) in a DAG
	// belonging to the specified workflow in order of when the operator was last modified.
	query := `
	SELECT
		operator.id AS operator_id,
		operator.name AS operator_name, 
		workflow_dag.created_at AS modified_at,
		resource.name AS resource_name,
		CAST(json_extract(operator.spec, '$.load') AS BLOB) AS spec 	
	FROM 
		operator, resource, workflow_dag_edge, workflow_dag
	WHERE (
		json_extract(operator.spec, '$.type')='load' AND 
		resource.id = json_extract(operator.spec, '$.load.integration_id') AND
		( 
			workflow_dag_edge.from_id = operator.id OR 
			workflow_dag_edge.to_id = operator.id 
		) AND 
		workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
		workflow_dag.workflow_id = $1
	)
	GROUP BY
		operator.name,
		resource.name,
		json_extract(operator.spec, '$.load')	
	ORDER BY modified_at DESC;
	`
	args := []interface{}{workflowID}

	var operators []views.LoadOperator
	err := DB.Query(ctx, &operators, query, args...)
	return operators, err
}

func (*operatorReader) GetExtractAndLoadOPsByResource(
	ctx context.Context,
	resourceID uuid.UUID,
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
	args := []interface{}{resourceID, resourceID}

	return getOperators(ctx, DB, query, args...)
}

// This currently only works with relational and S3 loads!
func (*operatorReader) GetLoadOPsByWorkflowAndResource(
	ctx context.Context,
	workflowID uuid.UUID,
	resourceID uuid.UUID,
	objectName string,
	DB database.Database,
) ([]models.Operator, error) {
	// Get all load operators where table=objectName & integration_id=resourceId
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
	args := []interface{}{objectName, resourceID, workflowID}

	return getOperators(ctx, DB, query, args...)
}

func (*operatorReader) GetLoadOPsByResource(
	ctx context.Context,
	resourceID uuid.UUID,
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
	args := []interface{}{resourceID, resourceID}

	return getOperators(ctx, DB, query, args...)
}

func (*operatorReader) GetLoadOPSpecsByOrg(ctx context.Context, orgID string, DB database.Database) ([]views.LoadOperatorSpec, error) {
	// Get the artifact id, artifact name, operator id, workflow name, workflow id,
	// and operator spec of all load operators (`to_id`s) and the artifact(s) going to
	// that operator (`from_id`s; these artifacts are the objects that will be saved
	// by the operator to the resource) in the workflows owned by the specified
	// organization.
	query := fmt.Sprintf(
		`SELECT DISTINCT 
			workflow_dag_edge.from_id AS artifact_id, 
			artifact.name AS artifact_name, 
		 	operator.id AS load_operator_id, 
			workflow.name AS workflow_name, 
			workflow.id AS workflow_id, 
			workflow_dag_edge.workflow_dag_id AS workflow_dag_id,
			operator.spec 
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

func (*operatorReader) GetByEngineResourceID(
	ctx context.Context,
	resourceID uuid.UUID,
	DB database.Database,
) ([]models.Operator, error) {
	workflow_condition_fragments := make([]string, 0, len(shared.ServiceToEngineConfigField))
	operator_condition_fragments := make([]string, 0, len(shared.ServiceToEngineConfigField))
	for _, field := range shared.ServiceToEngineConfigField {
		workflow_condition_fragments = append(
			workflow_condition_fragments,
			fmt.Sprintf(
				`json_extract(
					workflow_dag.engine_config,
					'$.%s.integration_id'
				) = $1`,
				field),
		)

		operator_condition_fragments = append(
			operator_condition_fragments,
			fmt.Sprintf(
				`json_extract(
					operator.spec,
					'$.engine_config.%s.integration_id'
				) = $1`,
				field),
		)
	}

	workflow_condition := strings.Join(workflow_condition_fragments, " OR ")
	operator_condition := strings.Join(operator_condition_fragments, " OR ")

	query := fmt.Sprintf(`
		SELECT DISTINCT %s FROM
		operator, workflow_dag, workflow_dag_edge
		WHERE
		workflow_dag_edge.workflow_dag_id = workflow_dag.id
		AND (
			workflow_dag_edge.from_id = operator.id
			OR workflow_dag_edge.to_id = operator.id
		)
		AND (
			(
				json_extract(operator.spec, '$.engine_config') IS NULL
				AND (%s)
			)
			OR (%s)
		);`,
		models.OperatorColsWithPrefix(),
		workflow_condition,
		operator_condition,
	)
	args := []interface{}{resourceID}

	var results []models.Operator
	err := DB.Query(ctx, &results, query, args...)
	return results, err
}

func (*operatorReader) GetForAqueductEngine(
	ctx context.Context,
	DB database.Database,
) ([]models.Operator, error) {
	workflowCondition := `
		json_extract(workflow_dag.engine_config, '$.type') == 'aqueduct'
	`
	operatorCondition := `
		json_extract(operator.spec, '$.engine_config.type') == 'aqueduct'
	`

	query := fmt.Sprintf(`
		SELECT DISTINCT %s FROM
		operator, workflow_dag, workflow_dag_edge
		WHERE
		workflow_dag_edge.workflow_dag_id = workflow_dag.id
		AND (
			workflow_dag_edge.from_id = operator.id
			OR workflow_dag_edge.to_id = operator.id
		)
		AND (
			(
				json_extract(operator.spec, '$.engine_config') IS NULL
				AND (%s)
			)
			OR (%s)
		);`,
		models.OperatorColsWithPrefix(),
		workflowCondition,
		operatorCondition,
	)

	var results []models.Operator
	err := DB.Query(ctx, &results, query)
	return results, err
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
		SELECT DISTINCT
			all_env_names.name AS name
		FROM 
			all_env_names, latest_workflow_dag, workflow_dag_edge
		WHERE
			latest_workflow_dag.id = workflow_dag_edge.workflow_dag_id 
			AND 
			workflow_dag_edge.type = '%s' 
			AND 
			workflow_dag_edge.from_id = all_env_names.op_id
	)
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

func (*operatorReader) GetEngineTypesMapByDagIDs(
	ctx context.Context,
	DagIDs []uuid.UUID,
	DB database.Database,
) (map[uuid.UUID][]shared.EngineType, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT
			workflow_dag_edge.workflow_dag_id as dag_id,
			ifnull(
				json_extract(operator.spec, '$.engine_config.type'),
				''
			) as engine_type
		FROM operator, workflow_dag_edge
		WHERE
			(workflow_dag_edge.from_id = operator.id
			OR workflow_dag_edge.to_id = operator.id)
			AND workflow_dag_edge.workflow_dag_id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(DagIDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(DagIDs)
	var resultRows []struct {
		DagID      uuid.UUID         `db:"dag_id"`
		EngineType shared.EngineType `db:"engine_type"`
	}

	err := DB.Query(ctx, &resultRows, query, args...)
	if err != nil {
		return nil, err
	}

	results := make(map[uuid.UUID][]shared.EngineType, len(resultRows))
	for _, row := range resultRows {
		results[row.DagID] = append(results[row.DagID], row.EngineType)
	}

	return results, nil
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

func getOperatorNodes(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]views.OperatorNode, error) {
	var operatorNodes []views.OperatorNode
	err := DB.Query(ctx, &operatorNodes, query, args...)
	return operatorNodes, err
}

func getOperatorWithArtifactNodes(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]views.OperatorWithArtifactNode, error) {
	var mergedNodes []views.OperatorWithArtifactNode
	err := DB.Query(ctx, &mergedNodes, query, args...)
	return mergedNodes, err
}

func getOperator(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.Operator, error) {
	operators, err := getOperators(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(operators) == 0 {
		return nil, database.ErrNoRows()
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
