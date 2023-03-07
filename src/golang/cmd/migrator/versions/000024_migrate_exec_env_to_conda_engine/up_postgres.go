package _000024_migrate_exec_env_to_conda_engine

const upPostgresScript = `
UPDATE operator SET spec = jsonb_set(spec, '{engine_config}', jsonb_build_object(
	'type', 'aqueduct_conda', 'aqueduct_conda_config', jsonb_build_object(
		'env', 'aqueduct_' || execution_environment_id
	)
)) WHERE execution_environment_id IS NOT NULL;
`
