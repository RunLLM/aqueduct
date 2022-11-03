import sys
from typing import Any

from aqueduct_executor.operators.connectors.data import common, config, connector, extract
from aqueduct_executor.operators.connectors.data.spec import (
    AQUEDUCT_DEMO_NAME,
    AuthenticateSpec,
    DiscoverSpec,
    ExtractSpec,
    LoadSpec,
    LoadTableSpec,
    Spec,
)
from aqueduct_executor.operators.utils import enums, utils
from aqueduct_executor.operators.utils.exceptions import MissingConnectorDependencyException
from aqueduct_executor.operators.utils.execution import (
    TIP_DEMO_CONNECTION,
    TIP_EXTRACT,
    TIP_INTEGRATION_CONNECTION,
    TIP_LOAD,
    TIP_UNKNOWN_ERROR,
    ExecFailureException,
    ExecutionState,
    Logs,
    exception_traceback,
)
from aqueduct_executor.operators.utils.storage.parse import parse_storage
from aqueduct_executor.operators.utils.storage.storage import Storage


def run(spec: Spec) -> None:
    """
    Runs one of the following connector operations:
    - authenticate
    - extract
    - load
    - load-table
    - delete-saved-objects
    - discover

    Arguments:
    - spec: The spec provided for this operator.
    """
    print("Started %s job: %s" % (spec.type, spec.name))

    storage = parse_storage(spec.storage_config)
    exec_state = ExecutionState(user_logs=Logs())

    try:
        _execute(spec, storage, exec_state)
        # Write operator execution metadata
        # Each decorator may set exec_state.status to FAILED, but if none of them did, then we are
        # certain that the operator succeeded.
        if exec_state.status == enums.ExecutionStatus.FAILED:
            print(f"Failed with error. Full Logs:\n{exec_state.json()}")
            utils.write_exec_state(storage, spec.metadata_path, exec_state)
            sys.exit(1)

        exec_state.status = enums.ExecutionStatus.SUCCEEDED
        utils.write_exec_state(storage, spec.metadata_path, exec_state)
    except ExecFailureException as e:
        # We must reconcile the user logs here, since those logs are not captured on the exception.
        from_exception_exec_state = ExecutionState.from_exception(e, user_logs=exec_state.user_logs)

        print(f"Failed with error. Full Logs:\n{from_exception_exec_state.json()}")
        utils.write_exec_state(storage, spec.metadata_path, from_exception_exec_state)
        sys.exit(1)
    except MissingConnectorDependencyException as e:
        exec_state.mark_as_failure(
            enums.FailureType.USER_FATAL, tip=str(e), context=exception_traceback(e)
        )
        utils.write_exec_state(storage, spec.metadata_path, exec_state)
        sys.exit(1)
    except Exception as e:
        exec_state.mark_as_failure(
            enums.FailureType.SYSTEM, tip=TIP_UNKNOWN_ERROR, context=exception_traceback(e)
        )
        print(f"Failed with system error. Full Logs:\n{exec_state.json()}")

        utils.write_exec_state(storage, spec.metadata_path, exec_state)
        sys.exit(1)


def _execute(spec: Spec, storage: Storage, exec_state: ExecutionState) -> None:

    if isinstance(spec.connector_name, dict):
        run_delete_saved_objects(spec, storage, exec_state)
    else:
        # Because constructing certain connectors (eg. Postgres) can also involve authentication,
        # we do both in `run_authenticate()`, and give a more helpful error message on failure.
        if spec.type == enums.JobType.AUTHENTICATE:
            run_authenticate(spec, exec_state, is_demo=(spec.name == AQUEDUCT_DEMO_NAME))
            return

        op = setup_connector(spec.connector_name, spec.connector_config)
        if spec.type == enums.JobType.EXTRACT:
            run_extract(spec, op, storage, exec_state)
        elif spec.type == enums.JobType.LOADTABLE:
            run_load_table(spec, op, storage)
        elif spec.type == enums.JobType.LOAD:
            run_load(spec, op, storage, exec_state)
        elif spec.type == enums.JobType.DISCOVER:
            run_discover(spec, op, storage)
        else:
            raise Exception("Unknown job: %s" % spec.type)


def run_authenticate(
    spec: AuthenticateSpec,
    exec_state: ExecutionState,
    is_demo: bool,
) -> None:
    @exec_state.user_fn_redirected(
        failure_tip=TIP_DEMO_CONNECTION if is_demo else TIP_INTEGRATION_CONNECTION
    )
    def _authenticate() -> None:
        op = setup_connector(spec.connector_name, spec.connector_config)
        op.authenticate()

    _authenticate()


def run_extract(
    spec: ExtractSpec, op: connector.DataConnector, storage: Storage, exec_state: ExecutionState
) -> None:
    extract_params = spec.parameters

    # Search for user-defined placeholder if this is a relational query, and replace them with
    # the appropriate values.
    if isinstance(extract_params, extract.RelationalParams) or isinstance(
        extract_params, extract.MongoDBParams
    ):
        assert len(spec.input_param_names) == len(spec.input_content_paths)
        input_vals, _ = utils.read_artifacts(
            storage,
            spec.input_content_paths,
            spec.input_metadata_paths,
        )
        assert all(
            isinstance(param_val, str) for param_val in input_vals
        ), "Parameter value must be a string."

        parameters = dict(zip(spec.input_param_names, input_vals))
        extract_params.compile(parameters)

    @exec_state.user_fn_redirected(failure_tip=TIP_EXTRACT)
    def _extract() -> Any:
        return op.extract(spec.parameters)

    output = _extract()

    output_artifact_type = enums.ArtifactType.TABLE
    if isinstance(extract_params, extract.S3Params):
        output_artifact_type = extract_params.artifact_type
        # If the type of the output is tuple, then it could be a multi-file S3 request so we
        # overwrite the output type to tuple.
        if isinstance(output, tuple):
            output_artifact_type = enums.ArtifactType.TUPLE

    if exec_state.status != enums.ExecutionStatus.FAILED:
        utils.write_artifact(
            storage,
            output_artifact_type,
            spec.output_content_path,
            spec.output_metadata_path,
            output,
            system_metadata={},
        )


def run_delete_saved_objects(spec: Spec, storage: Storage, exec_state: ExecutionState) -> None:
    results = {}
    assert isinstance(spec.connector_name, dict)
    for integration in spec.connector_name:
        op = setup_connector(spec.connector_name[integration], spec.connector_config[integration])
        results[integration] = op.delete(spec.integration_to_object[integration])
    utils.write_delete_saved_objects_results(storage, spec.output_content_path, results)


def run_load(
    spec: LoadSpec, op: connector.DataConnector, storage: Storage, exec_state: ExecutionState
) -> None:
    inputs, input_types = utils.read_artifacts(
        storage,
        [spec.input_content_path],
        [spec.input_metadata_path],
    )
    if len(inputs) != 1:
        raise Exception("Expected 1 input artifact, but got %d" % len(inputs))

    @exec_state.user_fn_redirected(failure_tip=TIP_LOAD)
    def _load() -> None:
        op.load(spec.parameters, inputs[0], input_types[0])

    _load()


def run_load_table(spec: LoadTableSpec, op: connector.DataConnector, storage: Storage) -> None:
    df = utils._read_csv(storage.get(spec.csv))
    op.load(spec.load_parameters.parameters, df, enums.ArtifactType.TABLE)


def run_discover(spec: DiscoverSpec, op: connector.DataConnector, storage: Storage) -> None:
    tables = op.discover()
    utils.write_discover_results(storage, spec.output_content_path, tables)


def setup_connector(
    connector_name: common.Name, connector_config: config.Config
) -> connector.DataConnector:
    # prevent isort from moving around type: ignore comments which will cause mypy issues.
    # isort: off
    if connector_name == common.Name.AQUEDUCT_DEMO or connector_name == common.Name.POSTGRES:
        try:
            import psycopg2
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the Postgres connector. Have you run `aqueduct install postgres`?"
            )

        from aqueduct_executor.operators.connectors.data.postgres import (
            PostgresConnector as OpConnector,
        )
    elif connector_name == common.Name.SNOWFLAKE:
        try:
            from snowflake import sqlalchemy
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the Snowflake connector. Have you run `aqueduct install snowflake`?"
            )

        from aqueduct_executor.operators.connectors.data.snowflake import (  # type: ignore
            SnowflakeConnector as OpConnector,
        )
    elif connector_name == common.Name.BIG_QUERY:
        try:
            from google.cloud import bigquery
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the BigQuery connector. Have you run `aqueduct install bigquery`?"
            )

        from aqueduct_executor.operators.connectors.data.bigquery import (  # type: ignore
            BigQueryConnector as OpConnector,
        )
    elif connector_name == common.Name.REDSHIFT:
        try:
            import psycopg2
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the Redshift connector. Have you run `aqueduct install redshift`?"
            )

        from aqueduct_executor.operators.connectors.data.redshift import (  # type: ignore
            RedshiftConnector as OpConnector,
        )
    elif connector_name == common.Name.SQL_SERVER:
        try:
            import pyodbc
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the SQL Server connector. Have you run `aqueduct install sqlserver`?"
            )

        from aqueduct_executor.operators.connectors.data.sql_server import (  # type: ignore
            SqlServerConnector as OpConnector,
        )
    elif connector_name == common.Name.MYSQL:
        try:
            import MySQLdb
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the MySQL connector. Have you run `aqueduct install mysql`?"
            )

        from aqueduct_executor.operators.connectors.data.mysql import (  # type: ignore
            MySqlConnector as OpConnector,
        )
    elif connector_name == common.Name.MARIA_DB:
        try:
            import MySQLdb
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the MariaDB connector. Have you run `aqueduct install mariadb`?"
            )

        from aqueduct_executor.operators.connectors.data.maria_db import (  # type: ignore
            MariaDbConnector as OpConnector,
        )
    elif connector_name == common.Name.AZURE_SQL:
        try:
            import pyodbc
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the Azure SQL connector. Have you run `aqueduct install azuresql`?"
            )

        from aqueduct_executor.operators.connectors.data.azure_sql import (  # type: ignore
            AzureSqlConnector as OpConnector,
        )
    elif connector_name == common.Name.S3:
        try:
            import pyarrow
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the S3 connector. Have you run `aqueduct install s3`?"
            )

        from aqueduct_executor.operators.connectors.data.s3 import (  # type: ignore
            S3Connector as OpConnector,
        )
    elif connector_name == common.Name.SQLITE:
        from aqueduct_executor.operators.connectors.data.sqlite import (  # type: ignore
            SqliteConnector as OpConnector,
        )
    elif connector_name == common.Name.GCS:
        from aqueduct_executor.operators.connectors.data.gcs import (  # type: ignore
            GCSConnector as OpConnector,
        )
    elif connector_name == common.Name.ATHENA:
        try:
            import awswrangler
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the Athena connector. Have you run `aqueduct install athena`?"
            )

        from aqueduct_executor.operators.connectors.data.athena import (  # type: ignore
            AthenaConnector as OpConnector,
        )
    elif connector_name == common.Name.MONGO_DB:
        try:
            import pymongo
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize MongoDB connector. Have you run `aqueduct install mongodb`?"
            )

        from aqueduct_executor.operators.connectors.data.mongodb import (  # type: ignore
            MongoDBConnector as OpConnector,
        )
    else:
        raise Exception("Unknown connector name: %s" % connector_name)

    # isort: on
    return OpConnector(config=connector_config)  # type: ignore
