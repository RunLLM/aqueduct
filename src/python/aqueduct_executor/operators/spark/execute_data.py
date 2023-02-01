import sys
from typing import Any

from aqueduct_executor.operators.connectors.data import common, config, connector, extract
from aqueduct_executor.operators.connectors.data.execute import (
    run_authenticate,
    run_delete_saved_objects,
    run_discover,
    setup_connector,
)
from aqueduct_executor.operators.connectors.data.spec import (
    AQUEDUCT_DEMO_NAME,
    AuthenticateSpec,
    DiscoverSpec,
    ExtractSpec,
    LoadSpec,
    LoadTableSpec,
    Spec,
)
from aqueduct_executor.operators.spark.utils import read_artifacts_spark, write_artifact_spark
from aqueduct_executor.operators.utils import utils
from aqueduct_executor.operators.utils.enums import (
    ArtifactType,
    ExecutionStatus,
    FailureType,
    JobType,
    SerializationType,
)
from aqueduct_executor.operators.utils.exceptions import (
    MissingConnectorDependencyException,
    UnsupportedConnectorExecption,
)
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
from pyspark.sql import SparkSession


def run(spec: Spec, spark_session_obj: SparkSession) -> None:
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
        _execute_spark(spec, storage, exec_state, spark_session_obj)
        # Write operator execution metadata
        # Each decorator may set exec_state.status to FAILED, but if none of them did, then we are
        # certain that the operator succeeded.
        if exec_state.status == ExecutionStatus.FAILED:
            print(f"Failed with error. Full Logs:\n{exec_state.json()}")
            utils.write_exec_state(storage, spec.metadata_path, exec_state)
            sys.exit(1)

        exec_state.status = ExecutionStatus.SUCCEEDED
        utils.write_exec_state(storage, spec.metadata_path, exec_state)
    except ExecFailureException as e:
        # We must reconcile the user logs here, since those logs are not captured on the exception.
        from_exception_exec_state = ExecutionState.from_exception(e, user_logs=exec_state.user_logs)

        print(f"Failed with error. Full Logs:\n{from_exception_exec_state.json()}")
        utils.write_exec_state(storage, spec.metadata_path, from_exception_exec_state)
        sys.exit(1)
    except MissingConnectorDependencyException as e:
        exec_state.mark_as_failure(
            FailureType.USER_FATAL, tip=str(e), context=exception_traceback(e)
        )
        utils.write_exec_state(storage, spec.metadata_path, exec_state)
        sys.exit(1)
    except Exception as e:
        exec_state.mark_as_failure(
            FailureType.SYSTEM, tip=TIP_UNKNOWN_ERROR, context=exception_traceback(e)
        )
        print(f"Failed with system error. Full Logs:\n{exec_state.json()}")

        utils.write_exec_state(storage, spec.metadata_path, exec_state)
        sys.exit(1)


def _execute_spark(
    spec: Spec, storage: Storage, exec_state: ExecutionState, spark_session_obj: SparkSession
) -> None:
    if spec.type == JobType.DELETESAVEDOBJECTS:
        run_delete_saved_objects(spec, storage, exec_state)

    # Because constructing certain connectors (eg. Postgres) can also involve authentication,
    # we do both in `run_authenticate()`, and give a more helpful error message on failure.
    elif spec.type == JobType.AUTHENTICATE:
        run_authenticate(spec, exec_state, is_demo=(spec.name == AQUEDUCT_DEMO_NAME))

    else:
        op = setup_connector_spark(spec.connector_name, spec.connector_config)
        if spec.type == JobType.EXTRACT:
            run_extract_spark(spec, op, storage, exec_state, spark_session_obj)
        elif spec.type == JobType.LOADTABLE:
            run_load_table_spark(spec, op, storage, spark_session_obj)
        elif spec.type == JobType.LOAD:
            run_load_spark(spec, op, storage, exec_state, spark_session_obj)
        elif spec.type == JobType.DISCOVER:
            run_discover(spec, op, storage)
        else:
            raise Exception("Unknown job: %s" % spec.type)


def run_extract_spark(
    spec: ExtractSpec,
    op: connector.DataConnector,
    storage: Storage,
    exec_state: ExecutionState,
    spark_session_obj: SparkSession,
) -> None:
    extract_params = spec.parameters

    # Search for user-defined placeholder if this is a relational query, and replace them with
    # the appropriate values.
    if isinstance(extract_params, extract.RelationalParams) or isinstance(
        extract_params, extract.MongoDBParams
    ):
        assert len(spec.input_param_names) == len(spec.input_content_paths)
        input_vals, _, _ = read_artifacts_spark(
            storage,
            spec.input_content_paths,
            spec.input_metadata_paths,
            spark_session_obj,
        )
        assert all(
            isinstance(param_val, str) for param_val in input_vals
        ), "Parameter value must be a string."

        parameters = dict(zip(spec.input_param_names, input_vals))
        extract_params.compile(parameters)

    @exec_state.user_fn_redirected(failure_tip=TIP_EXTRACT)
    def _extract() -> Any:
        return op.extract_spark(spec.parameters, spark_session_obj)  # type: ignore

    output = _extract()

    output_artifact_type = ArtifactType.TABLE
    derived_from_bson = isinstance(extract_params, extract.MongoDBParams)
    if isinstance(extract_params, extract.S3Params):
        output_artifact_type = extract_params.artifact_type
        # If the type of the output is tuple, then it could be a multi-file S3 request so we
        # overwrite the output type to tuple.
        if isinstance(output, tuple):
            output_artifact_type = ArtifactType.TUPLE

    if exec_state.status != ExecutionStatus.FAILED:
        write_artifact_spark(
            storage,
            output_artifact_type,
            derived_from_bson,
            spec.output_content_path,
            spec.output_metadata_path,
            output,
            system_metadata={},
            spark_session_obj=spark_session_obj,
        )


def run_load_spark(
    spec: LoadSpec,
    op: connector.DataConnector,
    storage: Storage,
    exec_state: ExecutionState,
    spark_session_obj: SparkSession,
) -> None:
    inputs, input_types, _ = read_artifacts_spark(
        storage,
        [spec.input_content_path],
        [spec.input_metadata_path],
        spark_session_obj,
    )
    if len(inputs) != 1:
        raise Exception("Expected 1 input artifact, but got %d" % len(inputs))

    @exec_state.user_fn_redirected(failure_tip=TIP_LOAD)
    def _load() -> None:
        op.load_spark(spec.parameters, inputs[0], input_types[0])  # type: ignore

    _load()


def run_load_table_spark(
    spec: LoadTableSpec,
    op: connector.DataConnector,
    storage: Storage,
    spark_session_obj: SparkSession,
) -> None:
    df = utils._read_csv(storage.get(spec.csv))
    op.load_spark(spec.load_parameters.parameters, df, ArtifactType.TABLE)  # type: ignore


def setup_connector_spark(
    connector_name: common.Name, connector_config: config.Config
) -> connector.DataConnector:
    # prevent isort from moving around type: ignore comments which will cause mypy issues.
    # isort: off
    if connector_name == common.Name.SNOWFLAKE:
        try:
            from pyspark.sql import SparkSession
            from snowflake import sqlalchemy
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the Spark Snowflake connector. Have you run `aqueduct install spark-snowflake`?"
            )

        from aqueduct_executor.operators.connectors.data.spark.snowflake import (
            SparkSnowflakeConnector as OpConnector,
        )
    elif connector_name == common.Name.S3:
        try:
            import pyarrow
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the Spark S3 connector. Have you run `aqueduct install s3`?"
            )

        from aqueduct_executor.operators.connectors.data.spark.s3 import (  # type: ignore
            SparkS3Connector as OpConnector,
        )
    else:
        raise UnsupportedConnectorExecption(
            "Unable to initialize connector. This connector is not yet supported for Aqueduct on Spark."
        )
    # isort: on
    return OpConnector(config=connector_config)  # type: ignore
