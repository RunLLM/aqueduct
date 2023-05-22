import os

import pytest

# Maps the test files in this directory to the allowed data resources for that file.
# If a disallowed data resource is used, all tests in the file will be skipped.
from aqueduct.constants.enums import ServiceType

import aqueduct as aq
from sdk.data_resource_tests.flow_manager import FlowManager

from ..setup_resource import get_aqueduct_config
from ..shared.flow_helpers import delete_all_flows

allowed_data_resources_by_file = {
    "relational_test": [
        ServiceType.BIGQUERY,
        ServiceType.REDSHIFT,
        ServiceType.SQLITE,
        ServiceType.SNOWFLAKE,
        ServiceType.MARIADB,
        ServiceType.MYSQL,
    ],
    "s3_test": [ServiceType.S3],
    "mongo_db_test": [ServiceType.MONGO_DB],
    "athena_test": [ServiceType.ATHENA],
}


@pytest.fixture(autouse=True)
def filter_tests_based_on_data_resources(request, client, data_resource):
    """Does the same thing as `enable_only_for_data_resource_type()`, only over entire files.

    This is because the data resource tests are grouped such that each file is only relevant for
    a specific resource(s).

    All that is required is that every file define a `REQUIRED_INTEGRATION=...` variable, so we know
    which data resources to skip.
    """
    test_file_name = os.path.splitext(os.path.basename(request.fspath))[
        0
    ]  # The extension is stripped out.

    assert test_file_name in allowed_data_resources_by_file, (
        "%s.py has not specified what data resources it's allowed to run with, please add those "
        "to the dict in `data_resource_tests/conftest.py`" % test_file_name
    )

    allowed_data_resources = allowed_data_resources_by_file[test_file_name]
    if data_resource.type() not in allowed_data_resources:
        pytest.skip(
            "Skipped for data resource `%s`, since it is not of type `%s`."
            % (data_resource.name(), ",".join(allowed_data_resources))
        )


@pytest.fixture
def flow_manager(client, flow_name, engine):
    """This a purely a convenience fixture to package some flow-related fields together that
    data resource tests usually don't care about.

    This allows test cases in this suite to import one fixture in order to publish flows,
    instead of three. Data resource tests usually don't care about how flows are published,
    as it is mostly a mechanism by which data can be saved.
    """
    return FlowManager(client, flow_name, engine)


def pytest_sessionfinish(session, exitstatus):
    # hasattr(session.config, "workerinput") ensures
    # this only triggers after all workflow finishes.
    if not hasattr(session.config, "workerinput") and not session.config.getoption("keep_flows"):
        client = aq.Client(*get_aqueduct_config())
        delete_all_flows(client)
