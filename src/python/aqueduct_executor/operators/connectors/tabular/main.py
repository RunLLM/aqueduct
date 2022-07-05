import argparse
import base64
import json
import sys

from pydantic import parse_obj_as

from aqueduct_executor.operators.connectors.tabular import (
    common,
    config,
    connector,
    extract,
    spec,
)
from aqueduct_executor.operators.utils import enums, utils
from aqueduct_executor.operators.utils.execution import (
    Error,
    ExecutionState,
    Logs,
    TIP_EXTRACT,
    TIP_INTEGRATION_CONNECTION,
    TIP_DEMO_CONNECTION,
    TIP_LOAD,
    TIP_UNKNOWN_ERROR,
    exception_traceback,
)
from aqueduct_executor.operators.utils.storage.parse import parse_storage
from aqueduct_executor.operators.utils.storage.storage import Storage

from aqueduct_executor.operators.connectors.tabular import common, config
from aqueduct_executor.operators.utils import enums


def run(spec: spec.Spec, storage: Storage, exec_state: ExecutionState):
    """
    Runs one of the following connector operations:
    - authenticate
    - extract
    - load
    - load-table
    - discover

    Arguments:
    - spec: The spec provided for this operator.
    - storage: An execution storage to use for reading or writing artifacts.
    """
    op = setup_connector(spec.connector_name, spec.connector_config)

    if spec.type == enums.JobType.AUTHENTICATE:
        run_authenticate(op, exec_state, is_demo=(spec.name == spec.AQUEDUCT_DEMO_NAME))
    elif spec.type == enums.JobType.EXTRACT:
        run_extract(spec, op, storage, exec_state)
    elif spec.type == enums.JobType.LOADTABLE:
        run_load_table(spec, op, storage)
    elif spec.type == enums.JobType.LOAD:
        run_load(spec, op, storage, exec_state)
    elif spec.type == enums.JobType.DISCOVER:
        run_discover(spec, op, storage)
    else:
        raise Exception("Unknown job: %s" % spec.type)


def run_authenticate(op: connector.TabularConnector, exec_state: ExecutionState, is_demo: bool):
    @exec_state.user_fn_redirected(
        failure_tip=TIP_DEMO_CONNECTION if is_demo else TIP_INTEGRATION_CONNECTION
    )
    def _authenticate():
        op.authenticate()

    _authenticate()


def run_extract(spec: spec.ExtractSpec, op: connector.TabularConnector, storage: Storage):
    extract_params = spec.parameters

    # Search for user-defined placeholder if this is a relational query, and replace them with
    # the appropriate values.
    if isinstance(extract_params, extract.RelationalParams):
        assert len(spec.input_param_names) == len(spec.input_content_paths)
        input_vals = utils.read_artifacts(
            storage,
            spec.input_content_paths,
            spec.input_metadata_paths,
            [utils.InputArtifactType.JSON] * len(spec.input_content_paths),
        )
        assert all(
            isinstance(param_val, str) for param_val in input_vals
        ), "Parameter value must be a string."

        parameters = dict(zip(spec.input_param_names, input_vals))
        extract_params.expand_placeholders(parameters)

    @exec_state.user_fn_redirected(failure_tip=TIP_EXTRACT)
    def _extract():
        return op.extract(spec.parameters)

    df = _extract()
    if exec_state.status != enums.ExecutionStatus.FAILED:
        utils.write_artifacts(
            storage,
            [utils.OutputArtifactType.TABLE],
            [spec.output_content_path],
            [spec.output_metadata_path],
            [df],
            system_metadata={},
        )


def run_load(
    spec: spec.LoadSpec,
    op: connector.TabularConnector,
    storage: Storage,
    exec_state: ExecutionState,
):
    inputs = utils.read_artifacts(
        storage,
        [spec.input_content_path],
        [spec.input_metadata_path],
        [utils.InputArtifactType.TABLE],
    )
    if len(inputs) != 1:
        raise Exception("Expected 1 input artifact, but got %d" % len(inputs))

    @exec_state.user_fn_redirected(failure_tip=TIP_LOAD)
    def _load():
        op.load(spec.parameters, inputs[0])

    _load()


def run_load_table(spec: spec.LoadTableSpec, op: connector.TabularConnector, storage: Storage):
    df = utils._read_csv(storage, spec.csv)
    op.load(spec.load_parameters.parameters, df)


def run_discover(spec: spec.DiscoverSpec, op: connector.TabularConnector, storage: Storage):
    tables = op.discover()
    utils.write_discover_results(storage, spec.output_content_path, tables)


def setup_connector(
    connector_name: common.Name, connector_config: config.Config
) -> connector.TabularConnector:
    if connector_name == common.Name.AQUEDUCT_DEMO or connector_name == common.Name.POSTGRES:
        from aqueduct_executor.operators.connectors.tabular.postgres import (
            PostgresConnector as OpConnector,
        )
    elif connector_name == common.Name.SNOWFLAKE:
        from aqueduct_executor.operators.connectors.tabular.snowflake import (
            SnowflakeConnector as OpConnector,
        )
    elif connector_name == common.Name.BIG_QUERY:
        from aqueduct_executor.operators.connectors.tabular.bigquery import (
            BigQueryConnector as OpConnector,
        )
    elif connector_name == common.Name.REDSHIFT:
        from aqueduct_executor.operators.connectors.tabular.redshift import (
            RedshiftConnector as OpConnector,
        )
    elif connector_name == common.Name.SQL_SERVER:
        from aqueduct_executor.operators.connectors.tabular.sql_server import (
            SqlServerConnector as OpConnector,
        )
    elif connector_name == common.Name.MYSQL:
        from aqueduct_executor.operators.connectors.tabular.mysql import (
            MySqlConnector as OpConnector,
        )
    elif connector_name == common.Name.MARIA_DB:
        from aqueduct_executor.operators.connectors.tabular.maria_db import (
            MariaDbConnector as OpConnector,
        )
    elif connector_name == common.Name.AZURE_SQL:
        from aqueduct_executor.operators.connectors.tabular.azure_sql import (
            AzureSqlConnector as OpConnector,
        )
    elif connector_name == common.Name.S3:
        from aqueduct_executor.operators.connectors.tabular.s3 import S3Connector as OpConnector
    elif connector_name == common.Name.SQLITE:
        from aqueduct_executor.operators.connectors.tabular.sqlite import (
            SqliteConnector as OpConnector,
        )
    else:
        raise Exception("Unknown connector name: %s" % connector_name)

    return OpConnector(config=connector_config)


def _parse_spec(spec_json: str) -> spec.Spec:
    """
    Parses a JSON string into a spec.Spec.
    """
    data = json.loads(spec_json)

    print("Job Spec: \n{}".format(json.dumps(data, indent=4)))

    return parse_obj_as(spec.Spec, data)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = _parse_spec(spec_json)

    print("Started %s job: %s" % (spec.type, spec.name))

    storage = parse_storage(spec.storage_config)
    exec_state = ExecutionState(user_logs=Logs())

    try:
        run(spec, storage, exec_state)
        # Write operator execution metadata
        if exec_state.status != enums.ExecutionStatus.FAILED:
            exec_state.status = enums.ExecutionStatus.SUCCEEDED
        utils.write_exec_state(storage, spec.metadata_path, exec_state)
    except Exception as e:
        exec_state.status = enums.ExecutionStatus.FAILED
        exec_state.failure_type = enums.FailureType.SYSTEM
        exec_state.error = Error(context=exception_traceback(e), tip=TIP_UNKNOWN_ERROR)
        print(f"Failed with system error. Full Logs:\n{exec_state.json()}")
        utils.write_exec_state(storage, spec.metadata_path, exec_state)
        sys.exit(1)
