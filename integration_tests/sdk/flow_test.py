import uuid
from datetime import datetime, timedelta

import pandas as pd
import pytest
from aqueduct.enums import ExecutionStatus, ServiceType
from aqueduct.error import InvalidUserArgumentException
from aqueduct.integrations.airflow_integration import AirflowIntegration
from aqueduct.integrations.integration import IntegrationInfo
from constants import SENTIMENT_SQL_QUERY
from test_functions.sentiment.model import sentiment_model
from test_functions.simple.model import (
    dummy_model,
    dummy_sentiment_model,
    dummy_sentiment_model_multiple_input,
)
from test_metrics.constant.model import constant_metric
from utils import (
    delete_flow,
    generate_new_flow_name,
    generate_table_name,
    run_flow_test,
    wait_for_flow_runs, publish_flow_test, trigger_flow_test,
)

import aqueduct
from aqueduct import FlowConfig, LoadUpdateMode, check, metric, op


def test_basic_flow(client, flow_name, data_integration, engine):
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    publish_flow_test(
        client,
        flow_name(),
        output_artifact,
        engine=engine,
    )


def test_sentiment_flow(client, flow_name, data_integration, engine):
    """Actually run the full sentiment model (with nltk dependency)."""
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    publish_flow_test(
        client,
        flow_name(),
        output_artifact,
        engine=engine,
    )


def test_complex_flow(client, flow_name, data_integration, engine):
    db = client.integration(data_integration)
    sql_artifact1 = db.sql(name="Query 1", query=SENTIMENT_SQL_QUERY)
    sql_artifact2 = db.sql(name="Query 2", query=SENTIMENT_SQL_QUERY)

    fn_artifact = dummy_sentiment_model_multiple_input(sql_artifact1, sql_artifact2)
    output_artifact = dummy_model(fn_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    @check()
    def successful_check(df):
        return True

    @check()
    def failing_check(df):
        return False

    @metric
    def dummy_metric(df):
        return 123

    success_check = successful_check(output_artifact)
    _ = failing_check(output_artifact)
    _ = dummy_metric(output_artifact)

    # Test that metrics and checks can be implicitly included, and that a non-error check
    # failing does not fail the flow.
    name = flow_name()
    flow = publish_flow_test(
        client,
        name,
        output_artifact,
        engine=engine,
    )

    # Metrics and checks should have been computed.
    flow_run = flow.latest()
    assert flow_run.artifact("dummy_metric artifact") is not None
    assert flow_run.artifact("successful_check artifact") is not None
    assert flow_run.artifact("failing_check artifact") is not None

    flow = publish_flow_test(
        client,
        name,
        output_artifact,
        engine=engine,
        checks=[success_check],  # failing_check will no longer be included.
        existing_flow=flow,
    )

    # Only the explicitly defined metrics and checks should have been included in this second run.
    flow_run = flow.latest()
    assert flow_run.artifact("dummy_metric artifact") is not None
    assert flow_run.artifact("successful_check artifact") is not None
    assert flow_run.artifact("failing_check artifact") is None


def test_multiple_output_artifacts(client, flow_name, data_integration, engine):
    db = client.integration(data_integration)
    sql_artifact1 = db.sql(name="Query 1", query=SENTIMENT_SQL_QUERY)
    sql_artifact2 = db.sql(name="Query 2", query=SENTIMENT_SQL_QUERY)

    fn_artifact1 = dummy_sentiment_model(sql_artifact1)
    fn_artifact2 = dummy_model(sql_artifact2)
    fn_artifact1.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )
    fn_artifact2.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    publish_flow_test(
        client,
        flow_name(),
        artifacts=[fn_artifact1, fn_artifact2],
        engine=engine,
    )


def test_publish_with_schedule(client, flow_name, data_integration, engine):
    db = client.integration(
        data_integration,
    )
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    # Execute the flow 1 minute from now.
    execute_at = datetime.now() + timedelta(minutes=1)
    publish_flow_test(
        client,
        flow_name(),
        artifacts=[output_artifact],
        engine=engine,
        schedule=aqueduct.hourly(minute=aqueduct.Minute(execute_at.minute)),

        # Wait for two runs because registering a workflow always triggers an immediate run first.
        expected_status=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
    )


def test_invalid_flow(client):
    with pytest.raises(InvalidUserArgumentException):
        client.publish_flow(
            name=generate_new_flow_name(),
            artifacts=[],
        )

    with pytest.raises(Exception):
        client.publish_flow(
            name=generate_new_flow_name(),
            artifacts=["123"],
        )


def test_publish_flow_with_same_name(client, flow_name, data_integration, engine):
    """Tests flow editing behavior."""
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)

    name = flow_name()
    flow = publish_flow_test(
        client,
        name,
        output_artifact,
        engine=engine,
        schedule=aqueduct.daily(),
    )

    # Add a metric to the flow and re-publish under the same name.
    metric = constant_metric(output_artifact)

    publish_flow_test(
        client,
        name,
        metric,
        engine=engine,
        schedule=aqueduct.daily(),
        existing_flow=flow,
    )


def test_refresh_flow(client, flow_name, data_integration, engine):
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    flow = publish_flow_test(
        client,
        flow_name(),
        output_artifact,
        engine=engine,
        schedule=aqueduct.hourly(),
    )

    # Trigger the workflow again verify that it runs one more time.
    trigger_flow_test(
        client,
        flow,
    )


def test_get_artifact_from_flow(client, flow_name, data_integration, engine):
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    flow = publish_flow_test(
        client,
        flow_name(),
        output_artifact,
        engine=engine,
    )

    artifact_return = flow.latest().artifact(output_artifact.name())
    assert artifact_return.name() == output_artifact.name()
    assert artifact_return.get().equals(output_artifact.get())


def test_get_artifact_reuse_for_computation(client, flow_name, data_integration, engine):
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )
    flow = publish_flow_test(
        client,
        flow_name(),
        output_artifact,
        engine=engine,
    )

    artifact_return = flow.latest().artifact(output_artifact.name())
    with pytest.raises(Exception):
        output_artifact = dummy_sentiment_model(artifact_return)


def test_multiple_flows_with_same_schedule(client, flow_name, data_integration, engine):
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact_2 = dummy_model(sql_artifact)

    flow_1 = publish_flow_test(
        client,
        flow_name(),
        output_artifact,
        engine=engine,
        schedule="* * * * *",
        should_block=False,
    )

    flow_2 = publish_flow_test(
        client,
        flow_name(),
        output_artifact_2,
        engine=engine,
        schedule="* * * * *",
        should_block=False,
    )

    wait_for_flow_runs(
        client,
        flow_1.id(),
        expected_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
    )
    wait_for_flow_runs(
        client,
        flow_2.id(),
        expected_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
    )


def test_fetching_historical_flows_uses_old_data(client, flow_name, data_integration, engine):
    db = client.integration(data_integration)

    # Write a new table into the demo db.
    initial_table = pd.DataFrame([1, 2, 3, 4, 5, 6], columns=["numbers"])

    @op
    def generate_initial_table():
        return initial_table

    setup_flow_name = flow_name()
    table = generate_initial_table()
    table.save(db.config(table="test_table", update_mode=LoadUpdateMode.REPLACE))

    setup_flow = publish_flow_test(
        client,
        setup_flow_name,
        artifacts=table,
        engine=engine,
    )

    @op
    def noop(df):
        return df

    # Create a new flow that extracts this data.
    output = db.sql("Select * from test_table", name="Test Table Query")
    assert output.get().equals(initial_table)

    flow = publish_flow_test(
        client,
        flow_name(),
        artifacts=output,
        engine=engine,
    )

    # Now, change the data that the new flow relies on, by populating data the same way as the setup flow.
    @op
    def generate_new_table():
        return pd.DataFrame([9, 9, 9, 9, 9, 9], columns=["numbers"])

    table = generate_new_table()
    table.save(db.config(table="test_table", update_mode=LoadUpdateMode.REPLACE))
    publish_flow_test(
        client,
        setup_flow_name,
        artifacts=table,
        engine=engine,
        existing_flow=setup_flow,
    )

    # Fetching the historical flow and materializing the data will not use the new data
    # that was just written. It will use a snapshot of the old data instead.
    artifact = flow.latest().artifact(name="Test Table Query artifact")
    assert artifact.get().equals(initial_table)


def test_flow_with_args(client):
    str_val = "this is a string"
    num_val = 1234

    @op
    def foo_with_args(*args):
        args_list = list(args)
        assert args_list == [str_val, num_val]
        return args_list

    @op
    def generate_str():
        return str_val

    @op
    def generate_num():
        return num_val

    output = foo_with_args(generate_str(), generate_num())
    assert output.get() == [str_val, num_val]

    # Implicit parameter creation is disallowed for variable-length parameters.
    with pytest.raises(InvalidUserArgumentException):
        foo_with_args(*[str_val, num_val])


def test_publish_with_redundant_config_fields(client):
    """Once the user-facing `FlowConfig` struct is deprecated, we can get rid of this test."""

    @op
    def noop():
        return 123

    output = noop()

    # Test redundant engine field.
    dummy_integration_info = IntegrationInfo(
        id=uuid.uuid4(),
        name="dummy",
        service=ServiceType.LAMBDA,
        createdAt=123,
        validated=True,
    )
    with pytest.raises(InvalidUserArgumentException):
        client.publish_flow(
            generate_new_flow_name(),
            artifacts=[output],
            engine="something",
            config=FlowConfig(engine=AirflowIntegration(dummy_integration_info)),
        )

    # Test redundant `k_latest_runs` field.
    with pytest.raises(InvalidUserArgumentException):
        client.publish_flow(
            generate_new_flow_name(),
            artifacts=[output],
            k_latest_runs=10,
            config=FlowConfig(k_latest_runs=123),
        )
