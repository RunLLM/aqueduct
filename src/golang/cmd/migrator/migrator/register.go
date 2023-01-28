package migrator

import (
	"context"

	_000001 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000001_base"
	_000002 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000002_add_user_id_to_integration"
	_000003 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000003_add_storage_column"
	_000004 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000004_storage_interface_backfill"
	_000005 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000005_storage_interface_not_null"
	_000006 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000006_add_retention_policy_column"
	_000007 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000007_workflow_dag_edge_pk"
	_000008 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000008_delete_s3_config"
	_000009 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000009_metadata_interface_backfill"
	_000010 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000010_add_exec_state_column"
	_000011 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000011_exec_state_column_backfill"
	_000012 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000012_drop_metadata_column"
	_000013 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000013_add_workflow_dag_engine_config"
	_000014 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000014_add_exec_state_column_to_artifact_result"
	_000015 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000015_artifact_result_exec_state_column_backfill"
	_000016 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000016_add_artifact_type_column_to_artifact"
	_000017 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000017_update_to_canceled_status"
	_000018 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000018_add_dag_result_exec_state_column"
	_000019 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000019_add_serialization_type_value_to_param_op"
	_000020 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000020_add_execution_environment_table"
	_000021 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000021_add_gc_column_to_env_table"
	_000022 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000022_backfill_python_type"
	_000023 "github.com/aqueducthq/aqueduct/cmd/migrator/versions/000023_add_notification_settings_column"
	"github.com/aqueducthq/aqueduct/lib/database"
)

var registeredMigrations = map[int64]*migration{}

type migrationFunc func(context.Context, database.Database) error

type migration struct {
	upPostgres   migrationFunc
	upSqlite     migrationFunc
	downPostgres migrationFunc
	name         string
}

func init() {
	registeredMigrations[1] = &migration{
		upPostgres: _000001.UpPostgres, upSqlite: _000001.UpSqlite,
		name: "base",
	}
	registeredMigrations[2] = &migration{
		upPostgres: _000002.UpPostgres, upSqlite: _000002.UpSqlite,
		downPostgres: _000002.DownPostgres,
		name:         "add integration.user_id",
	}
	registeredMigrations[3] = &migration{
		upPostgres: _000003.UpPostgres, upSqlite: _000003.UpSqlite,
		downPostgres: _000003.DownPostgres,
		name:         "add workflow_dag.storage_config",
	}
	registeredMigrations[4] = &migration{
		upPostgres: _000004.Up, upSqlite: _000004.Up,
		downPostgres: _000004.Down,
		name:         "backfill workflow_dag.storage_config and operator.spec.storage_path",
	}
	registeredMigrations[5] = &migration{
		upPostgres: _000005.UpPostgres, upSqlite: _000005.UpSqlite,
		downPostgres: _000005.DownPostgres,
		name:         "add not null constraint to workflow_dag.storage_config",
	}
	registeredMigrations[6] = &migration{
		upPostgres: _000006.UpPostgres, upSqlite: _000006.UpSqlite,
		downPostgres: _000006.DownPostgres,
		name:         "add workflow.retention_policy",
	}
	registeredMigrations[7] = &migration{
		upPostgres: _000007.UpPostgres, upSqlite: _000007.UpSqlite,
		downPostgres: _000007.DownPostgres,
		name:         "add primary key constraint to workflow_dag_edge on workflow_dag_id, from_id, to_id",
	}

	registeredMigrations[8] = &migration{
		upPostgres: _000008.UpPostgres, upSqlite: _000008.UpSqlite,
		downPostgres: _000008.DownPostgres,
		name:         "delete outdated s3_config column",
	}

	registeredMigrations[9] = &migration{
		upPostgres: _000009.Up, upSqlite: _000009.Up,
		downPostgres: _000009.Down,
		name:         "backfill metadata in artifact_results",
	}

	registeredMigrations[10] = &migration{
		upPostgres: _000010.UpPostgres, upSqlite: _000010.UpSqlite,
		downPostgres: _000010.DownPostgres,
		name:         "add exec state column to operator_result",
	}

	registeredMigrations[11] = &migration{
		upPostgres: _000011.Up, upSqlite: _000011.Up,
		downPostgres: _000011.Down,
		name:         "backfill exec state column in operator_result",
	}

	registeredMigrations[12] = &migration{
		upPostgres: _000012.UpPostgres, upSqlite: _000012.UpSqlite,
		downPostgres: _000012.DownPostgres,
		name:         "remove metadata in operator_result",
	}

	registeredMigrations[13] = &migration{
		upPostgres: _000013.UpPostgres, upSqlite: _000013.UpSqlite,
		downPostgres: _000013.DownPostgres,
		name:         "add workflow_dag.engine_config",
	}

	registeredMigrations[14] = &migration{
		upPostgres: _000014.UpPostgres, upSqlite: _000014.UpSqlite,
		downPostgres: _000014.DownPostgres,
		name:         "add exec state column to artifact result",
	}

	registeredMigrations[15] = &migration{
		upPostgres: _000015.Up, upSqlite: _000015.Up,
		downPostgres: _000015.Down,
		name:         "backfill exec state column in artifact result",
	}

	registeredMigrations[16] = &migration{
		upPostgres: _000016.UpPostgres, upSqlite: _000016.UpSqlite,
		downPostgres: _000016.DownPostgres,
		name:         "add type column to artifact",
	}

	registeredMigrations[17] = &migration{
		upPostgres: _000017.Up, upSqlite: _000017.Up,
		downPostgres: _000017.Down,
		name:         "add canceled status to results",
	}

	registeredMigrations[18] = &migration{
		upPostgres: _000018.UpPostgres, upSqlite: _000018.UpSqlite,
		downPostgres: _000018.DownPostgres,
		name:         "add exec state to workflow_dag_result",
	}

	registeredMigrations[19] = &migration{
		upPostgres: _000019.Up, upSqlite: _000019.Up,
		downPostgres: _000019.Down,
		name:         "add serialization type and value to param op",
	}

	registeredMigrations[20] = &migration{
		upPostgres: _000020.UpPostgres, upSqlite: _000020.UpSqlite,
		downPostgres: _000020.DownPostgres,
		name:         "add execution environment table",
	}

	registeredMigrations[21] = &migration{
		upPostgres: _000021.UpPostgres, upSqlite: _000021.UpSqlite,
		downPostgres: _000021.DownPostgres,
		name:         "add gc column to the execution environment table",
	}

	registeredMigrations[22] = &migration{
		upPostgres: _000022.Up, upSqlite: _000022.Up,
		downPostgres: _000022.Down,
		name:         "backfill python type to the artifact result table",
	}

	registeredMigrations[23] = &migration{
		upPostgres: _000023.UpPostgres, upSqlite: _000023.UpSqlite,
		downPostgres: _000023.DownPostgres,
		name:         "add notification_settings column to workflow table",
	}
}
