from datetime import datetime, timedelta

import pandas as pd
import pytest
from aqueduct.enums import ExecutionStatus
from aqueduct.error import InvalidUserArgumentException
from constants import SENTIMENT_SQL_QUERY
from test_functions.sentiment.model import sentiment_model
from test_functions.simple.model import dummy_model, dummy_sentiment_model, dummy_sentiment_model_multiple_input
from test_metrics.constant.model import constant_metric
from utils import (
    delete_flow,
    generate_new_flow_name,
    generate_table_name,
    get_integration_name,
    run_flow_test,
    wait_for_flow_runs,
)

import aqueduct
from aqueduct import LoadUpdateMode, check, metric, op


def test_basic_flow(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    run_flow_test(client, artifacts=[output_artifact])


def test_sentiment_flow(client):
    """Actually run the full sentiment model (with nltk dependency)."""
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    run_flow_test(client, artifacts=[output_artifact])


def test_complex_flow(client):
    db = client.integration(name=get_integration_name())
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
    dummy_metric = dummy_metric(output_artifact)

    # Test that publish_flow can implicitly and explicitly include metrics and checks.
    flow_name = generate_new_flow_name()

    flow_id = None
    try:
        # Test that metrics and checks can be implicitly included, and that a non-error check
        # failing does not fail the flow.
        flow = run_flow_test(
            client,
            name=flow_name,
            artifacts=[output_artifact],
            delete_flow_after=False,
        )
        flow_id = flow.id()

        # Metrics and checks should have been computed.
        flow_run = flow.latest()
        assert flow_run.artifact("dummy_metric artifact") is not None
        assert flow_run.artifact("successful_check artifact") is not None
        assert flow_run.artifact("failing_check artifact") is not None

        flow = run_flow_test(
            client,
            name=flow_name,
            artifacts=[output_artifact],
            checks=[success_check],  # failing_check will no longer be included.
            num_runs=2,
            delete_flow_after=False,
        )

        # Only the explicitly defined metrics and checks should have been included in this second run.
        flow_run = flow.latest()
        assert flow_run.artifact("dummy_metric artifact") is not None
        assert flow_run.artifact("successful_check artifact") is not None
        assert flow_run.artifact("failing_check artifact") is None
    finally:
        delete_flow(client, flow_id)


def test_multiple_output_artifacts(client):
    db = client.integration(name=get_integration_name())
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

    run_flow_test(
        client,
        artifacts=[fn_artifact1, fn_artifact2],
    )


def test_publish_with_schedule(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )

    # Execute the flow 1 minute from now.
    execute_at = datetime.now() + timedelta(minutes=1)
    run_flow_test(
        client,
        artifacts=[output_artifact],
        schedule=aqueduct.hourly(minute=aqueduct.Minute(execute_at.minute)),
        num_runs=2,  # Wait for two runs because registering a workflow always triggers an immediate run first.
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


def test_publish_flow_with_same_name(client):
    """Tests flow editing behavior."""
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)

    # Remember to cleanup any created test data.
    flow_ids_to_delete = set()
    try:
        flow_name = generate_new_flow_name()
        flow = run_flow_test(
            client,
            artifacts=[output_artifact],
            name=flow_name,
            schedule=aqueduct.daily(),
            delete_flow_after=False,
        )
        flow_ids_to_delete.add(flow.id())

        # Add a metric to the flow and re-publish under the same name.
        metric = constant_metric(output_artifact)
        flow = run_flow_test(
            client,
            artifacts=[metric],
            name=flow_name,
            schedule=aqueduct.daily(),
            num_runs=2,
            delete_flow_after=False,
        )
        flow_ids_to_delete.add(flow.id())
    finally:
        for flow_id in flow_ids_to_delete:
            delete_flow(client, flow_id)


def test_refresh_flow(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )
    flow = client.publish_flow(
        name=generate_new_flow_name(),
        artifacts=output_artifact,
        schedule=aqueduct.hourly(),
    )

    # Wait for the first run, then refresh the workflow and verify that it runs at least
    # one more time.
    try:
        wait_for_flow_runs(client, flow.id(), expect_statuses=[ExecutionStatus.SUCCEEDED])
        client.trigger(flow.id())
        wait_for_flow_runs(
            client,
            flow.id(),
            expect_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
        )
    finally:
        client.delete_flow(flow.id())


def test_get_artifact_from_flow(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )
    flow = client.publish_flow(
        name=generate_new_flow_name(),
        artifacts=output_artifact,
    )
    try:
        wait_for_flow_runs(client, flow.id(), expect_statuses=[ExecutionStatus.SUCCEEDED])
        artifact_return = flow.latest().artifact(output_artifact.name())
        assert artifact_return.name() == output_artifact.name()
        assert artifact_return.get().equals(output_artifact.get())
    finally:
        client.delete_flow(flow.id())


def test_get_artifact_reuse_for_computation(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_artifact.save(
        config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
    )
    flow = client.publish_flow(
        name=generate_new_flow_name(),
        artifacts=output_artifact,
    )
    try:
        wait_for_flow_runs(client, flow.id(), expect_statuses=[ExecutionStatus.SUCCEEDED])
        artifact_return = flow.latest().artifact(output_artifact.name())
        with pytest.raises(Exception):
            output_artifact = dummy_sentiment_model(artifact_return)
    finally:
        client.delete_flow(flow.id())


def test_multiple_flows_with_same_schedule(client):
    try:
        db = client.integration(name=get_integration_name())
        sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
        output_artifact = dummy_sentiment_model(sql_artifact)
        output_artifact_2 = dummy_model(sql_artifact)

        flow_1 = client.publish_flow(
            name=generate_new_flow_name(),
            artifacts=output_artifact,
            schedule="* * * * *",
        )

        flow_2 = client.publish_flow(
            name=generate_new_flow_name(),
            artifacts=output_artifact_2,
            schedule="* * * * *",
        )

        wait_for_flow_runs(
            client,
            flow_1.id(),
            expect_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
        )
        wait_for_flow_runs(
            client,
            flow_2.id(),
            expect_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
        )
    finally:
        delete_flow(client, flow_1.id())
        delete_flow(client, flow_2.id())


def test_fetching_historical_flows_uses_old_data(client):
    db = client.integration(name=get_integration_name())

    # Write a new table into the demo db.
    flows_to_delete = []
    try:
        initial_table = pd.DataFrame([1, 2, 3, 4, 5, 6], columns=["numbers"])

        @op
        def generate_initial_table():
            return initial_table

        setup_flow_name = generate_new_flow_name()
        table = generate_initial_table()
        table.save(db.config(table="test_table", update_mode=LoadUpdateMode.REPLACE))
        setup_flow = run_flow_test(
            client, name=setup_flow_name, artifacts=[table], delete_flow_after=False
        )
        flows_to_delete.append(setup_flow.id())

        @op
        def noop(df):
            return df

        # Create a new flow that extracts this data.
        output = db.sql("Select * from test_table", name="Test Table Query")
        assert output.get().equals(initial_table)

        flow = run_flow_test(client, artifacts=[output], delete_flow_after=False)
        flows_to_delete.append(flow.id())

        # Now, change the data that the new flow relies on, by populating data the same way as the setup flow.
        @op
        def generate_new_table():
            return pd.DataFrame([9, 9, 9, 9, 9, 9], columns=["numbers"])

        table = generate_new_table()
        table.save(db.config(table="test_table", update_mode=LoadUpdateMode.REPLACE))
        run_flow_test(
            client, name=setup_flow_name, artifacts=[table], num_runs=2, delete_flow_after=False
        )

        # Fetching the historical flow and materializing the data will not use the new data
        # that was just written. It will use a snapshot of the old data instead.
        artifact = flow.latest().artifact(name="Test Table Query artifact")
        assert artifact.get().equals(initial_table)

    finally:
        for flow_id in flows_to_delete:
            delete_flow(client, flow_id)
