from typing import List

from aqueduct.constants.enums import RelationalDBServices, ServiceType

# Default BigQuery dataset used for integration tests
BIG_QUERY_TEST_DATASET = "integration_test"


def all_relational_DBs() -> List[ServiceType]:
    return [ServiceType(relational_service.value) for relational_service in RelationalDBServices]


def format_table_name(table_name: str, service: ServiceType) -> str:
    """
    Returns the table name so it is formatted according to the integration
    service specified.
    """
    if service == ServiceType.BIGQUERY:
        # BigQuery table names need to be prefixed with the dataset
        return BIG_QUERY_TEST_DATASET + "." + table_name
    return table_name
