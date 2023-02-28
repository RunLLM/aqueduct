package _000024_migrate_exec_env_to_conda_engine

const downPostgresScript = `
UPDATE operator SET execution_environment_id = NULL
WHERE jsonb_extract_path_text(spec, '{engine_config, type}') != 'aqueduct_conda';

UPDATE operator SET spec = jsonb_set(spec, '{engine_config}', jsonb_build_object(
	'type', 'aqueduct'
)) WHERE jsonb_extract_path_text(spec, '{engine_config, type}') = 'aqueduct_conda';
`
