package _000026_add_operator_artifact_node_views

const upSqliteScript = `
CREATE VIEW IF NOT EXISTS operator_node
AS 
	WITH edge_ordered AS (
		SELECT * FROM workflow_dag_edge ORDER BY idx ASC
	),
	op_with_outputs AS ( -- Aggregate outputs
		SELECT
			operator.id AS id,
			workflow_dag.id AS dag_id,
			operator.name AS name,
			operator.description AS description,
			operator.spec AS spec,
			operator.execution_environment_id AS execution_environment_id,
			CAST( json_group_array(
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
			CAST( json_group_array(
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
    WHERE op_with_outputs.outputs IS NULL;

CREATE VIEW IF NOT EXISTS artifact_node
AS 
	WITH artf_with_outputs AS ( -- Aggregate outputs
		SELECT
			artifact.id AS id,
			workflow_dag.id AS dag_id,
			artifact.name AS name,
			artifact.description AS description,
			artifact.type as type,
			CAST( json_group_array(
				json_object(
					'value', workflow_dag_edge.to_id,
					'idx', workflow_dag_edge.idx
				)
			) AS BLOB) AS outputs
		FROM
			artifact, workflow_dag, workflow_dag_edge
		WHERE
			workflow_dag.id = workflow_dag_edge.workflow_dag_id
			AND artifact.id = workflow_dag_edge.from_id
		GROUP BY
			workflow_dag.id, artifact.id
	),
	artf_with_inputs AS ( -- Aggregate inputs
		SELECT
			artifact.id AS id,
			workflow_dag.id AS dag_id,
			artifact.name AS name,
			artifact.description AS description,
			artifact.type as type,
			CAST( json_group_array(
				json_object(
					'value', workflow_dag_edge.from_id,
					'idx', workflow_dag_edge.idx
				)
			) AS BLOB) AS inputs
		FROM
			artifact, workflow_dag, workflow_dag_edge
		WHERE
			workflow_dag.id = workflow_dag_edge.workflow_dag_id
			AND artifact.id = workflow_dag_edge.to_id
		GROUP BY
			workflow_dag.id, artifact.id
	)
	SELECT -- just do input LEFT JOIN outputs as all artifacts have inputs
		artf_with_inputs.id AS id,
		artf_with_inputs.dag_id AS dag_id,
		artf_with_inputs.name AS name,
		artf_with_inputs.description AS description,
		artf_with_inputs.type AS type,
		artf_with_outputs.outputs AS outputs,
		artf_with_inputs.inputs AS inputs
	FROM
		artf_with_inputs LEFT JOIN artf_with_outputs
	ON
		artf_with_outputs.id = artf_with_inputs.id
		AND artf_with_outputs.dag_id = artf_with_inputs.dag_id
	WHERE artf_with_outputs.outputs IS NULL;
`
