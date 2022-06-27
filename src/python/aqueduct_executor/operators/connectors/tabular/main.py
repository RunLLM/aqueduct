import argparse
import base64
import json
import sys

from pydantic import parse_obj_as

from aqueduct_executor.operators.connectors.tabular import common, config, connector
from aqueduct_executor.operators.connectors.tabular.spec import (
    AQUEDUCT_DEMO_NAME,
    DiscoverSpec,
    ExtractSpec,
    LoadSpec,
    LoadTableSpec,
    Spec,
)
from aqueduct_executor.operators.utils import enums, utils
from aqueduct_executor.operators.utils.logging import (
    Error,
    Logger,
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


def run(spec: Spec, storage: Storage, logger: Logger):
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
        run_authenticate(op, logger, is_demo=(spec.name == AQUEDUCT_DEMO_NAME))
    elif spec.type == enums.JobType.EXTRACT:
        run_extract(spec, op, storage, logger)
    elif spec.type == enums.JobType.LOADTABLE:
        run_load_table(spec, op, storage)
    elif spec.type == enums.JobType.LOAD:
        run_load(spec, op, storage, logger)
    elif spec.type == enums.JobType.DISCOVER:
        run_discover(spec, op, storage)
    else:
        raise Exception("Unknown job: %s" % spec.type)


def run_authenticate(op: connector.TabularConnector, logger: Logger, is_demo: bool):
    @logger.user_fn_redirected(
        failure_tip=TIP_DEMO_CONNECTION if is_demo else TIP_INTEGRATION_CONNECTION
    )
    def _authenticate():
        op.authenticate()

    _authenticate()


def run_extract(
    spec: ExtractSpec, op: connector.TabularConnector, storage: Storage, logger: Logger
):
    @logger.user_fn_redirected(failure_tip=TIP_EXTRACT)
    def _extract():
        return op.extract(spec.parameters)

    df = _extract()
    utils.write_artifacts(
        storage,
        [spec.output_content_path],
        [spec.output_metadata_path],
        [df],
        {},
        [utils.OutputArtifactType.TABLE],
    )


def run_load(spec: LoadSpec, op: connector.TabularConnector, storage: Storage, logger: Logger):
    inputs = utils.read_artifacts(
        storage,
        [spec.input_content_path],
        [spec.input_metadata_path],
        [utils.InputArtifactType.TABLE],
    )
    if len(inputs) != 1:
        raise Exception("Expected 1 input artifact, but got %d" % len(inputs))

    @logger.user_fn_redirected(failure_tip=TIP_LOAD)
    def _load():
        op.load(spec.parameters, inputs[0])

    _load()


def run_load_table(spec: LoadTableSpec, op: connector.TabularConnector, storage: Storage):
    df = utils._read_csv(storage, spec.csv)
    op.load(spec.load_parameters.parameters, df)


def run_discover(spec: DiscoverSpec, op: connector.TabularConnector, storage: Storage):
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
        from aqueduct_executor.operators.connectors.tabular.snowflake import (  # type: ignore
            SnowflakeConnector as OpConnector,
        )
    elif connector_name == common.Name.BIG_QUERY:
        from aqueduct_executor.operators.connectors.tabular.bigquery import (  # type: ignore
            BigQueryConnector as OpConnector,
        )
    elif connector_name == common.Name.REDSHIFT:
        from aqueduct_executor.operators.connectors.tabular.redshift import (  # type: ignore
            RedshiftConnector as OpConnector,
        )
    elif connector_name == common.Name.SQL_SERVER:
        from aqueduct_executor.operators.connectors.tabular.sql_server import (  # type: ignore
            SqlServerConnector as OpConnector,
        )
    elif connector_name == common.Name.MYSQL:
        from aqueduct_executor.operators.connectors.tabular.mysql import (  # type: ignore
            MySqlConnector as OpConnector,
        )
    elif connector_name == common.Name.MARIA_DB:
        from aqueduct_executor.operators.connectors.tabular.maria_db import (  # type: ignore
            MariaDbConnector as OpConnector,
        )
    elif connector_name == common.Name.AZURE_SQL:
        from aqueduct_executor.operators.connectors.tabular.azure_sql import (  # type: ignore
            AzureSqlConnector as OpConnector,
        )
    elif connector_name == common.Name.S3:
        from aqueduct_executor.operators.connectors.tabular.s3 import S3Connector as OpConnector  # type: ignore
    elif connector_name == common.Name.SQLITE:
        from aqueduct_executor.operators.connectors.tabular.sqlite import (  # type: ignore
            SqliteConnector as OpConnector,
        )
    else:
        raise Exception("Unknown connector name: %s" % connector_name)

    return OpConnector(config=connector_config)  # type: ignore


def _parse_spec(spec_json: bytes) -> Spec:
    """
    Parses a JSON string into a spec.Spec.
    """
    data = json.loads(spec_json)

    print("Job Spec: \n{}".format(json.dumps(data, indent=4)))

    # TODO: The following line is working, but mypy complains:
    # Argument 1 to "parse_obj_as" has incompatible type "object"; expected "Type[<nothing>]"
    # We ignore the error for now.
    return parse_obj_as(Spec, data)  # type: ignore


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = _parse_spec(spec_json)

    print("Started %s job: %s" % (spec.type, spec.name))

    storage = parse_storage(spec.storage_config)
    logger = Logger(user_logs=Logs())

    try:
        run(spec, storage, logger)
        # Write operator execution metadata
        if not logger.failed():
            logger.code = enums.ExecutionCode.SUCCEEDED
        utils.write_logs(storage, spec.metadata_path, logger)
    except Exception as e:
        logger.code = enums.ExecutionCode.SYSTEM_FAILURE
        logger.error = Error(context=exception_traceback(e), tip=TIP_UNKNOWN_ERROR)
        print(f"Failed with system error. Full Logs:\n{logger.json()}")
        utils.write_logs(storage, spec.metadata_path, logger)
        sys.exit(1)
