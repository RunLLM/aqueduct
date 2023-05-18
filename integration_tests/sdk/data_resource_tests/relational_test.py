from typing import List

import pandas as pd
import pytest
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.generic_artifact import GenericArtifact
from aqueduct.constants.enums import ExecutionStatus
from aqueduct.error import AqueductError, InvalidUserActionException, InvalidUserArgumentException
from aqueduct.models.operators import RelationalDBExtractParams
from aqueduct.resources.sql import RelationalDBResource

from aqueduct import LoadUpdateMode, metric, op

from ..shared.demo_db import demo_db_tables
from ..shared.naming import generate_table_name
from ..shared.relational import format_table_name
from ..shared.validation import check_artifact_was_computed
from .relational_data_validator import RelationalDataValidator
from .save import save
from .validation_helpers import check_hotel_reviews_table_artifact


@pytest.fixture(autouse=True)
def assert_data_resource_is_relational(client, data_resource):
    assert isinstance(data_resource, RelationalDBResource)


def _create_successful_sql_artifacts(
    client,
    data_resource,
    wrap_query_in_extract_params_struct: bool = False,
) -> List[BaseArtifact]:
    """Tests and returns artifacts for two types of sql queries: basic and chained.

    Every artifact is saved to a random table with update_mode='replace'.
    """
    hotel_reviews_query = "SELECT * FROM %s" % format_table_name(
        "hotel_reviews", data_resource.type()
    )
    chained_query = [
        "SELECT * FROM %s" % format_table_name("hotel_reviews", data_resource.type()),
        "SELECT review, review_date FROM $ WHERE reviewer_nationality = '$1'",
        "SELECT review FROM $",
    ]

    if wrap_query_in_extract_params_struct:
        hotel_reviews_query = RelationalDBExtractParams(query=hotel_reviews_query)
        chained_query = RelationalDBExtractParams(queries=chained_query)

    # Test a successful basic sql query.
    hotel_reviews_table = data_resource.sql(hotel_reviews_query)
    check_hotel_reviews_table_artifact(hotel_reviews_table)

    # Test a successful chain query.
    nationality = client.create_param("nationality", default=" United Kingdom ")
    chained_query_result = data_resource.sql(chained_query, parameters=[nationality])
    expected_chained_query_result = data_resource.sql(
        "SELECT review FROM %s WHERE reviewer_nationality=' United Kingdom '"
        % format_table_name("hotel_reviews", data_resource.type()),
    )
    assert expected_chained_query_result.get().equals(chained_query_result.get())

    artifacts = [hotel_reviews_table, chained_query_result]
    for artifact in artifacts:
        save(
            data_resource,
            artifact,
            format_table_name(generate_table_name(), data_resource.type()),
            LoadUpdateMode.REPLACE,
        )
    return artifacts


def test_sql_resource_query_and_save(client, flow_manager, data_resource):
    artifacts = _create_successful_sql_artifacts(client, data_resource)

    flow = flow_manager.publish_flow_test(artifacts=artifacts)

    relational_validator = RelationalDataValidator(client, data_resource)
    for artifact in artifacts:
        relational_validator.check_saved_artifact_data(
            flow, artifact.id(), expected_data=artifact.get()
        )


def test_sql_resource_query_and_save_relationaldbextractparams(
    client, flow_manager, data_resource
):
    artifacts = _create_successful_sql_artifacts(
        client, data_resource, wrap_query_in_extract_params_struct=True
    )
    flow = flow_manager.publish_flow_test(artifacts=artifacts)

    relational_validator = RelationalDataValidator(client, data_resource)
    for artifact in artifacts:
        relational_validator.check_saved_artifact_data(
            flow, artifact.id(), expected_data=artifact.get()
        )


def test_sql_resource_artifact_with_custom_metadata(flow_manager, data_resource):
    # TODO: validate custom descriptions once we can fetch descriptions easily.
    artifact = data_resource.sql(
        "SELECT * FROM %s" % format_table_name("hotel_reviews", data_resource.type()),
        name="Test Artifact",
        description="This is a description",
    )
    assert artifact.name() == "Test Artifact artifact"

    flow = flow_manager.publish_flow_test(artifacts=artifact)
    check_artifact_was_computed(flow, "Test Artifact artifact")


def test_sql_resource_failed_query(client, data_resource):
    # Sql query is malformed.
    with pytest.raises(AqueductError, match="Preview Execution Failed"):
        data_resource.sql("SELECT * FROM ")

    # SQL error happens at execution time (table missing).
    with pytest.raises(AqueductError, match="Preview Execution Failed"):
        data_resource.sql(
            "SELECT * FROM %s" % format_table_name("missing_table", data_resource.type())
        )


def test_sql_resource_table_retrieval(client, data_resource):
    df = data_resource.table(name=format_table_name("hotel_reviews", data_resource.type()))
    assert len(df) == 100
    assert list(df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]


def test_sql_resource_list_tables(client, data_resource):
    tables = data_resource.list_tables()

    for expected_table in demo_db_tables():
        assert tables["tablename"].str.contains(expected_table, case=False).sum() > 0


def test_sql_today_tag(client, data_resource):
    table_artifact_today = data_resource.sql(
        query="select * from %s where review_date = {{today}}"
        % format_table_name("hotel_reviews", data_resource.type())
    )
    assert table_artifact_today.get().empty
    table_artifact_not_today = data_resource.sql(
        query="select * from %s where review_date < {{today}}"
        % format_table_name("hotel_reviews", data_resource.type())
    )
    assert len(table_artifact_not_today.get()) == 100


def test_sql_query_with_parameters(client, data_resource, flow_manager):
    table_name = client.create_param(
        "table name", default=format_table_name("hotel_reviews", data_resource.type())
    )
    column_name = client.create_param("column name", default="reviewer_nationality")
    column_value = client.create_param("column value", default=" United Kingdom ")
    parameterized_output = data_resource.sql(
        query="Select * from $1 where $2 = '$3'", parameters=[table_name, column_name, column_value]
    )
    expanded_output = data_resource.sql(
        query="Select * from %s where reviewer_nationality = ' United Kingdom '"
        % format_table_name("hotel_reviews", data_resource.type())
    )
    assert parameterized_output.get().equals(expanded_output.get())

    # Test that .get(parameters={...}) works.
    expanded_custom_output = data_resource.sql(
        query="Select * from %s where reviewer_nationality = ' Australia '"
        % format_table_name("hotel_reviews", data_resource.type())
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


def test_sql_query_invalid_parameters(client, data_resource, flow_manager):
    country = client.create_param("country", default=" United Kingdom ")

    # Error if provided parameters are not all used.
    with pytest.raises(
        InvalidUserArgumentException,
        match="Unused parameter `country`.* must contain the placeholder \$1",
    ):
        data_resource.sql(
            query="Select * from %s where reviewer_nationality = $2"
            % format_table_name("hotel_reviews", data_resource.type()),
            parameters=[country],
        )

    # Error if we use the {{built-in tag}} syntax improperly.
    with pytest.raises(
        InvalidUserActionException, match="`something` is not a valid Aqueduct placeholder"
    ):
        data_resource.sql(query="Select * from {{something }}")

    # Error if the parameter is not a string type.
    num = client.create_param("num", default=1234)
    with pytest.raises(InvalidUserArgumentException, match="must be defined as a string"):
        data_resource.sql(
            query="Select * from %s where reviewer_nationality = '$1'"
            % format_table_name("hotel_reviews", data_resource.type()),
            parameters=[num],
        )

    # Error if the parameter we attempt to set a custom parameter that is not a string.
    output = data_resource.sql(
        query="Select * from %s where reviewer_nationality = '$1'"
        % format_table_name("hotel_reviews", data_resource.type()),
        parameters=[country],
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


def test_sql_resource_save_wrong_data_type(client, flow_manager, data_resource):
    # Try to save a numeric artifact.
    num_param = client.create_param("number", default=123)
    with pytest.raises(
        InvalidUserActionException,
        match="Unable to save non-relational data into relational data store",
    ):
        save(data_resource, num_param, generate_table_name(), LoadUpdateMode.REPLACE)

    # Save a generic artifact that is actually a string. This won't fail at save() time,
    # but instead when the flow is published.
    @op
    def foo():
        return "asdf"

    string_artifact = foo.lazy()
    assert isinstance(string_artifact, GenericArtifact)
    save(data_resource, string_artifact, generate_table_name(), LoadUpdateMode.REPLACE)
    flow_manager.publish_flow_test(
        artifacts=string_artifact,
        expected_statuses=ExecutionStatus.FAILED,
    )


def test_sql_resource_save_with_different_update_modes(client, flow_manager, data_resource):
    table_1_save_name = format_table_name(generate_table_name(), data_resource.type())
    table_2_save_name = format_table_name(generate_table_name(), data_resource.type())

    table = data_resource.sql(
        "select * from %s limit 5" % format_table_name("hotel_reviews", data_resource.type())
    )
    extracted_table_data = table.get()
    save(data_resource, table, table_1_save_name, LoadUpdateMode.REPLACE)

    # This will create the table.
    relational_validator = RelationalDataValidator(client, data_resource)
    flow = flow_manager.publish_flow_test(artifacts=table)
    relational_validator.check_saved_artifact_data(
        flow, table.id(), expected_data=extracted_table_data
    )

    # Change to append mode.
    save(data_resource, table, table_1_save_name, LoadUpdateMode.APPEND)
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
    save(data_resource, table, table_1_save_name, LoadUpdateMode.APPEND)
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
    save(data_resource, table, table_2_save_name, LoadUpdateMode.REPLACE)
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
