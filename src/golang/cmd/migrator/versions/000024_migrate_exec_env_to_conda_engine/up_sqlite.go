package _000024_migrate_exec_env_to_conda_engine

const upSqliteScript = `
UPDATE operator SET spec = CAST(
	json_set(spec, '$.engine_config', json_object(
		'type', 'aqueduct_conda', 'aqueduct_conda_config', json_object(
			'env', 'aqueduct_' || execution_environment_id
		)
	)) AS BLOB
) WHERE execution_environment_id IS NOT NULL;
`
