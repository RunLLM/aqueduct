from typing import List

import pandas as pd
import pytest
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.generic_artifact import GenericArtifact
from aqueduct.constants.enums import ExecutionStatus
from aqueduct.error import AqueductError, InvalidUserActionException, InvalidUserArgumentException
from aqueduct.integrations.sql_integration import RelationalDBIntegration
from aqueduct.models.operators import RelationalDBExtractParams

from aqueduct import LoadUpdateMode, metric, op

from ..shared.demo_db import demo_db_tables
from ..shared.naming import generate_table_name
from ..shared.relational import SHORT_SENTIMENT_SQL_QUERY
from ..shared.validation import check_artifact_was_computed
from .relational_data_validator import RelationalDataValidator
from .save import save
from .validation_helpers import check_hotel_reviews_table_artifact


@pytest.fixture(autouse=True)
def assert_data_integration_is_relational(client, data_integration):
    assert isinstance(data_integration, RelationalDBIntegration)


def _create_successful_sql_artifacts(
    client,
    data_integration,
    wrap_query_in_extract_params_struct: bool = False,
) -> List[BaseArtifact]:
    """Tests and returns artifacts for two types of sql queries: basic and chained.

    Every artifact is saved to a random table with update_mode='replace'.
    """
    hotel_reviews_query = "SELECT * FROM hotel_reviews"
    chained_query = [
        "SELECT * FROM hotel_reviews",
        "SELECT review, review_date FROM $ WHERE reviewer_nationality = '$1'",
        "SELECT review FROM $",
    ]

    if wrap_query_in_extract_params_struct:
        hotel_reviews_query = RelationalDBExtractParams(query=hotel_reviews_query)
        chained_query = RelationalDBExtractParams(queries=chained_query)

    # Test a successful basic sql query.
    hotel_reviews_table = data_integration.sql(hotel_reviews_query)
    check_hotel_reviews_table_artifact(hotel_reviews_table)

    # Test a successful chain query.
    nationality = client.create_param("nationality", default=" United Kingdom ")
    chained_query_result = data_integration.sql(chained_query, parameters=[nationality])
    expected_chained_query_result = data_integration.sql(
        "SELECT review FROM hotel_reviews WHERE reviewer_nationality=' United Kingdom '",
    )
    assert expected_chained_query_result.get().equals(chained_query_result.get())

    artifacts = [hotel_reviews_table, chained_query_result]
    for artifact in artifacts:
        save(data_integration, artifact, generate_table_name(), LoadUpdateMode.REPLACE)
    return artifacts


def test_sql_integration_query_and_save(client, flow_manager, data_integration):
    artifacts = _create_successful_sql_artifacts(client, data_integration)

    flow = flow_manager.publish_flow_test(artifacts=artifacts)

    relational_validator = RelationalDataValidator(client, data_integration)
    for artifact in artifacts:
        relational_validator.check_saved_artifact_data(
            flow, artifact.id(), expected_data=artifact.get()
        )


def test_sql_integration_query_and_save_relationaldbextractparams(
    client, flow_manager, data_integration
):
    artifacts = _create_successful_sql_artifacts(
        client, data_integration, wrap_query_in_extract_params_struct=True
    )
    flow = flow_manager.publish_flow_test(artifacts=artifacts)

    relational_validator = RelationalDataValidator(client, data_integration)
    for artifact in artifacts:
        relational_validator.check_saved_artifact_data(
            flow, artifact.id(), expected_data=artifact.get()
        )


def test_sql_integration_artifact_with_custom_metadata(flow_manager, data_integration):
    # TODO: validate custom descriptions once we can fetch descriptions easily.
    artifact = data_integration.sql(
        "SELECT * FROM hotel_reviews", name="Test Artifact", description="This is a description"
    )
    assert artifact.name() == "Test Artifact artifact"

    flow = flow_manager.publish_flow_test(artifacts=artifact)
    check_artifact_was_computed(flow, "Test Artifact artifact")


def test_sql_integration_failed_query(client, data_integration):
    # Sql query is malformed.
    with pytest.raises(AqueductError, match="Preview Execution Failed"):
        data_integration.sql("SELECT * FROM ")

    # SQL error happens at execution time (table missing).
    with pytest.raises(AqueductError, match="Preview Execution Failed"):
        data_integration.sql("SELECT * FROM missing_table")


def test_sql_integration_table_retrieval(client, data_integration):
    df = data_integration.table(name="hotel_reviews")
    assert len(df) == 100
    assert list(df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]


def test_sql_integration_list_tables(client, data_integration):
    tables = data_integration.list_tables()

    for expected_table in demo_db_tables():
        assert tables["tablename"].str.contains(expected_table, case=False).sum() > 0


def test_sql_today_tag(client, data_integration):
    table_artifact_today = data_integration.sql(
        query="select * from hotel_reviews where review_date = {{today}}"
    )
    assert table_artifact_today.get().empty
    table_artifact_not_today = data_integration.sql(
        query="select * from hotel_reviews where review_date < {{today}}"
    )
    assert len(table_artifact_not_today.get()) == 100


def test_sql_query_with_parameters(client, data_integration, flow_manager):
    table_name = client.create_param("table name", default="hotel_reviews")
    column_name = client.create_param("column name", default="reviewer_nationality")
    column_value = client.create_param("column value", default=" United Kingdom ")
    parameterized_output = data_integration.sql(
        query="Select * from $1 where $2 = '$3'", parameters=[table_name, column_name, column_value]
    )
    expanded_output = data_integration.sql(
        query="Select * from hotel_reviews where reviewer_nationality = ' United Kingdom '"
    )
    assert parameterized_output.get().equals(expanded_output.get())

    # Test that .get(parameters={...}) works.
    expanded_custom_output = data_integration.sql(
        query="Select * from hotel_reviews where reviewer_nationality = ' Australia '"
    )
    assert parameterized_output.get(parameters={"column value": " Australia "}).equals(
        expanded_custom_output.get()
    )

    # Test that publishing this sql query works.
    parameterized_output.set_name("query output")
    flow = flow_manager.publish_flow_test(
        artifacts=parameterized_output,
    )
    flow_run = flow.latest()
    flow_run.artifact("query output").get().equals(expanded_output.get())

    # Test that client.trigger(parameters={...}) works.
    flow_manager.trigger_flow_test(flow, parameters={"column value": " Australia "})
    flow_run = flow.latest()
    flow_run.artifact("query output").get().equals(expanded_custom_output.get())

    # Check that `.get(parameters={...})` works even if the parameter is not directly fed into the operator
    # that produces the artifact.
    @metric
    def count_reviews(df):
        return len(df)

    len_df = count_reviews(parameterized_output)
    assert (
        len_df.get(parameters={"column value": " Australia "})
        == count_reviews(expanded_custom_output).get()
    )

    # Use the parameters in another operator.
    @metric
    def noop(sql_output, param):
        return len(param)


def test_sql_query_invalid_parameters(client, data_integration, flow_manager):
    country = client.create_param("country", default=" United Kingdom ")

    # Error if provided parameters are not all used.
    with pytest.raises(
        InvalidUserArgumentException,
        match="Unused parameter `country`.* must contain the placeholder \$1",
    ):
        data_integration.sql(
            query="Select * from hotel_reviews where reviewer_nationality = $2",
            parameters=[country],
        )

    # Error if we use the {{built-in tag}} syntax improperly.
    with pytest.raises(
        InvalidUserActionException, match="`something` is not a valid Aqueduct placeholder"
    ):
        data_integration.sql(query="Select * from {{something }}")

    # Error if the parameter is not a string type.
    num = client.create_param("num", default=1234)
    with pytest.raises(InvalidUserArgumentException, match="must be defined as a string"):
        data_integration.sql(
            query="Select * from hotel_reviews where reviewer_nationality = '$1'", parameters=[num]
        )

    # Error if the parameter we attempt to set a custom parameter that is not a string.
    output = data_integration.sql(
        query="Select * from hotel_reviews where reviewer_nationality = '$1'", parameters=[country]
    )
    with pytest.raises(
        InvalidUserArgumentException,
        match="Parameter `country` is used by a sql query, so it must be a string type, not type int",
    ):
        output.get(parameters={"country": 1234})

    flow = flow_manager.publish_flow_test(
        artifacts=output,
    )
    with pytest.raises(
        InvalidUserArgumentException,
        match="Parameter `country` is used by a sql query, so it must be a string type, not type int",
    ):
        client.trigger(flow.id(), parameters={"country": 1234})


def test_sql_integration_save_wrong_data_type(client, flow_manager, data_integration):
    # Try to save a numeric artifact.
    num_param = client.create_param("number", default=123)
    with pytest.raises(
        InvalidUserActionException,
        match="Unable to save non-relational data into relational data store",
    ):
        save(data_integration, num_param, generate_table_name(), LoadUpdateMode.REPLACE)

    # Save a generic artifact that is actually a string. This won't fail at save() time,
    # but instead when the flow is published.
    @op
    def foo():
        return "asdf"

    string_artifact = foo.lazy()
    assert isinstance(string_artifact, GenericArtifact)
    save(data_integration, string_artifact, generate_table_name(), LoadUpdateMode.REPLACE)
    flow_manager.publish_flow_test(
        artifacts=string_artifact,
        expected_statuses=ExecutionStatus.FAILED,
    )


def test_sql_integration_save_with_different_update_modes(client, flow_manager, data_integration):
    table_1_save_name = generate_table_name()
    table_2_save_name = generate_table_name()

    table = data_integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    extracted_table_data = table.get()
    save(data_integration, table, table_1_save_name, LoadUpdateMode.REPLACE)

    # This will create the table.
    relational_validator = RelationalDataValidator(client, data_integration)
    flow = flow_manager.publish_flow_test(artifacts=table)
    relational_validator.check_saved_artifact_data(
        flow, table.id(), expected_data=extracted_table_data
    )

    # Change to append mode.
    save(data_integration, table, table_1_save_name, LoadUpdateMode.APPEND)
    flow_manager.publish_flow_test(
        existing_flow=flow,
        artifacts=table,
    )
    relational_validator.check_saved_artifact_data(
        flow,
        table.id(),
        expected_data=pd.concat([extracted_table_data, extracted_table_data], ignore_index=True),
    )

    # Redundant append mode change
    save(data_integration, table, table_1_save_name, LoadUpdateMode.APPEND)
    flow_manager.publish_flow_test(
        existing_flow=flow,
        artifacts=table,
    )
    relational_validator.check_saved_artifact_data(
        flow,
        table.id(),
        expected_data=pd.concat(
            [extracted_table_data, extracted_table_data, extracted_table_data], ignore_index=True
        ),
    )

    # Create a different table from the same artifact.
    save(data_integration, table, table_2_save_name, LoadUpdateMode.REPLACE)
    flow_manager.publish_flow_test(
        existing_flow=flow,
        artifacts=table,
    )
    relational_validator.check_saved_artifact_data(
        flow,
        table.id(),
        expected_data=extracted_table_data,
    )

    relational_validator.check_saved_update_mode_changes(
        flow,
        expected_updates=[
            (table_2_save_name, LoadUpdateMode.REPLACE),
            (table_1_save_name, LoadUpdateMode.APPEND),
            (table_1_save_name, LoadUpdateMode.REPLACE),
        ],
    )
