import time
import uuid
from datetime import datetime, timedelta

import pandas as pd
import pytest
from aqueduct.constants.enums import ExecutionStatus, LoadUpdateMode
from aqueduct.error import InvalidRequestError, InvalidUserArgumentException

import aqueduct
from aqueduct import check, metric, op

from ..shared.data_objects import DataObject
from ..shared.flow_helpers import publish_flow_test, trigger_flow_test, wait_for_flow_runs
from ..shared.naming import generate_new_flow_name, generate_table_name
from ..shared.relational import all_relational_DBs
from .extract import extract
from .save import save
from .test_functions.sentiment.model import sentiment_model
from .test_functions.simple.model import (
    dummy_model,
    dummy_sentiment_model,
    dummy_sentiment_model_multiple_input,
)
from .test_metrics.constant.model import constant_metric


def test_basic_flow(client, flow_name, data_resource, engine, data_validator):
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_resource, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
    )

    data_validator.check_saved_artifact_data(
        flow, output_artifact.id(), expected_data=output_artifact.get()
    )


@pytest.mark.skip_for_spark_engines(reason="Uses sentiment model with pandas-specific code.")
def test_sentiment_flow(client, flow_name, data_resource, engine, data_validator):
    """Actually run the full sentiment model (with nltk dependency)."""
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = sentiment_model(table_artifact)
    save(data_resource, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
    )
    data_validator.check_saved_artifact_data(
        flow, output_artifact.id(), expected_data=output_artifact.get()
    )


def test_complex_flow(client, flow_name, data_resource, engine, data_validator):
    table_artifact1 = extract(data_resource, DataObject.SENTIMENT)
    table_artifact2 = extract(data_resource, DataObject.SENTIMENT)

    fn_artifact = dummy_sentiment_model_multiple_input(table_artifact1, table_artifact2)
    output_artifact = dummy_model(fn_artifact)
    save(data_resource, output_artifact)

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
    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
    )
    data_validator.check_saved_artifact_data(
        flow, output_artifact.id(), expected_data=output_artifact.get()
    )

    # Metrics and checks should have been computed.
    flow_run = flow.latest()
    assert flow_run.artifact("dummy_metric artifact") is not None
    assert flow_run.artifact("successful_check artifact") is not None
    assert flow_run.artifact("failing_check artifact") is not None

    flow = publish_flow_test(
        client,
        output_artifact,
        existing_flow=flow,
        engine=engine,
        checks=[success_check],  # failing_check will no longer be included.
    )
    data_validator.check_saved_artifact_data(
        flow, output_artifact.id(), expected_data=output_artifact.get()
    )

    # Only the explicitly defined metrics and checks should have been included in this second run.
    flow_run = flow.latest()
    assert flow_run.artifact("dummy_metric artifact") is not None
    assert flow_run.artifact("successful_check artifact") is not None
    assert flow_run.artifact("failing_check artifact") is None


def test_publish_with_schedule(client, flow_name, data_resource, engine):
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_resource, output_artifact)

    # Execute the flow 1 minute from now.
    execute_at = datetime.now() + timedelta(minutes=1)
    publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[output_artifact],
        engine=engine,
        schedule=aqueduct.hourly(minute=aqueduct.Minute(execute_at.minute)),
        # Wait for two runs because registering a workflow always triggers an immediate run first.
        expected_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
    )


@pytest.mark.skip_for_spark_engines(
    reason="Multiple Databricks cluster spinup takes longer than timeout."
)
def test_publish_flow_with_cascading_trigger(client, flow_name, data_resource, engine):
    """Tests publishing a flow that is set to run on a cascading trigger."""
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)

    # Create a source flow
    source_name = flow_name()
    source_flow = publish_flow_test(
        client,
        name=source_name,
        artifacts=output_artifact,
        engine=engine,
        schedule=aqueduct.daily(),
    )

    # Create a flow that is set to run after the above source_flow
    name = flow_name()
    flow = publish_flow_test(
        client,
        name=name,
        artifacts=output_artifact,
        engine=engine,
        source_flow=source_flow,
    )

    # Trigger a run of the source flow
    trigger_flow_test(
        client,
        source_flow,
    )

    # Verify that there are now 2 runs of flow
    wait_for_flow_runs(
        client,
        flow.id(),
        expected_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
    )


def test_publish_with_schedule_and_source_flow(client, flow_name, data_resource, engine):
    """Tests publishing an invalid flow that has both a schedule and a source flow."""
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)

    with pytest.raises(InvalidUserArgumentException):
        publish_flow_test(
            client,
            name=generate_new_flow_name(),
            artifacts=output_artifact,
            engine=engine,
            schedule=aqueduct.daily(),
            source_flow=uuid.uuid4(),
        )


def test_publish_with_source_flow_cyclic(client, flow_name, data_resource, engine):
    """Tests publishing an invalid flow, because it would cause a cycle amongst cascading workflows."""
    table_artifact = extract(data_resource, DataObject.SENTIMENT)

    @op
    def noop(input):
        return input

    output_artifact = noop(table_artifact)

    # First, create 3 workflows with the following dependencies: A --> B --> C
    a_flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=output_artifact,
        engine=engine,
        schedule=aqueduct.daily(),
    )

    b_flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=output_artifact,
        engine=engine,
        source_flow=a_flow,
    )

    c_flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=output_artifact,
        engine=engine,
        source_flow=b_flow,
    )

    # Now, change a_flow to have c_flow as its source, which would create a cyle
    with pytest.raises(InvalidRequestError):
        client.publish_flow(
            name=a_flow.name(),
            artifacts=[output_artifact],
            engine=engine,
            source_flow=c_flow,
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


def test_publish_flow_with_same_name(client, flow_name, data_resource, engine):
    """Tests flow editing behavior."""
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)

    flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=output_artifact,
        engine=engine,
        schedule=aqueduct.daily(),
    )

    # Add a metric to the flow and re-publish under the same name.
    metric = constant_metric(output_artifact)

    publish_flow_test(
        client,
        metric,
        engine=engine,
        schedule=aqueduct.daily(),
        existing_flow=flow,
    )


def test_refresh_flow(client, flow_name, data_resource, engine):
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_resource, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
        # Schedule the test far enough in the future so that the test will
        # finish before the next scheduled run.
        schedule=aqueduct.hourly(minute=(datetime.now().minute + 30) % 60),
    )

    # Trigger the workflow again verify that it runs one more time.
    trigger_flow_test(
        client,
        flow,
    )


def test_publish_flow_without_triggering(client, flow_name, data_resource, engine):
    @op
    def foo():
        return "results"

    output = foo()
    name = flow_name()
    flow = client.publish_flow(
        name=name,
        artifacts=output,
        engine=engine,
        run_now=False,
    )

    # flow.describe() should run without issue.
    flow.describe()
    assert len(flow.list_runs()) == 0
    assert flow.latest() is None
    assert flow.name() == name


@pytest.mark.skip_for_spark_engines(
    reason="Spark converts column names to capital, .equals doesn't work."
)
def test_get_artifact_from_flow(client, flow_name, data_resource, engine):
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_resource, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
    )

    artifact_return = flow.latest().artifact(output_artifact.name())
    assert artifact_return.name() == output_artifact.name()
    assert artifact_return.get().equals(output_artifact.get())


def test_get_artifact_reuse_for_computation(client, flow_name, data_resource, engine):
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_resource, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
    )

    artifact_return = flow.latest().artifact(output_artifact.name())
    with pytest.raises(Exception):
        output_artifact = dummy_sentiment_model(artifact_return)


@pytest.mark.skip_for_spark_engines(
    reason="Wating for 4 workflows takes longer than timeout period."
)
def test_multiple_flows_with_same_schedule(client, flow_name, data_resource, engine):
    table_artifact = extract(data_resource, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    output_artifact_2 = dummy_model(table_artifact)

    flow_1 = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
        schedule="* * * * *",
        should_block=False,
    )

    flow_2 = publish_flow_test(
        client,
        output_artifact_2,
        name=flow_name(),
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


@pytest.mark.skip_for_spark_engines(reason="Requires implicit Pandas requirement.")
def test_fetching_historical_flows_uses_old_data(client, flow_name, data_resource, engine):
    # Write a new table into the data resource.
    initial_table = pd.DataFrame([1, 2, 3, 4, 5, 6], columns=["numbers"])

    @op
    def generate_initial_table():
        # We copy the definition to bypass a known potential package version issue.
        # TODO (ENG-2814): Resolve the pkg issue and replace this with `initial_table` variable
        return pd.DataFrame([1, 2, 3, 4, 5, 6], columns=["numbers"])

    table = generate_initial_table()
    saved_table_identifier = generate_table_name()
    save(data_resource, table, name=saved_table_identifier)

    setup_flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=table,
        engine=engine,
    )

    @op
    def noop(df):
        return df

    # Create a new flow that extracts this data.
    output = extract(data_resource, saved_table_identifier, op_name="Test Table Query")
    assert output.get().equals(initial_table)

    flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=output,
        engine=engine,
    )

    # Now, change the data that the new flow relies on, by populating data the same way as the setup flow.
    @op
    def generate_new_table():
        return pd.DataFrame([9, 9, 9, 9, 9, 9], columns=["numbers"])

    table = generate_new_table()
    save(data_resource, table, name=saved_table_identifier)
    publish_flow_test(
        client,
        artifacts=table,
        existing_flow=setup_flow,
        engine=engine,
    )

    # Fetching the historical flow and materializing the data will not use the new data
    # that was just written. It will use a snapshot of the old data instead.
    artifact = flow.latest().artifact(name="Test Table Query artifact")
    assert artifact.get().equals(initial_table)


@pytest.mark.skip_for_spark_engines(reason="Cannot use implicit imports with Spark engines.")
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
        output_artifact = foo_with_args(*[str_val, num_val])
        output_artifact.get()


def test_flow_list_saved_objects_none(client, flow_name, engine):
    """Check that flow.list_saved_objects() works when no objects were actually saved."""

    @op
    def noop():
        return 123

    output = noop()
    flow = publish_flow_test(client, artifacts=output, name=flow_name(), engine=engine)
    assert len(flow.list_saved_objects()) == 0


def test_flow_with_disabled_snapshots(client, flow_name, data_resource, engine):
    @op
    def op_disabled_by_wf(df):
        return df

    @metric
    def metric_enabled(df):
        return df.shape[0]

    @metric
    def metric_disabled(df):
        return df.shape[0]

    @check
    def check_enabled(count):
        return count > 10

    reviews = extract(data_resource, DataObject.SENTIMENT, output_name="extract artifact")
    op_artf = op_disabled_by_wf(reviews)
    metric_enabled(op_artf)
    metric_disabled_artf = metric_disabled(op_artf)
    metric_disabled_artf.disable_snapshot()
    check_enabled(metric_disabled_artf)

    op_artf.enable_snapshot()
    save(data_resource, op_artf)

    flow = publish_flow_test(
        client, op_artf, name=flow_name(), engine=engine, disable_snapshots=True
    )
    latest_run = flow.latest()
    assert (
        latest_run.artifact("extract artifact").execution_state().status == ExecutionStatus.DELETED
    )
    assert latest_run.artifact("extract artifact").get() is None
    assert (
        latest_run.artifact("op_disabled_by_wf artifact").execution_state().status
        == ExecutionStatus.DELETED
    )
    assert latest_run.artifact("op_disabled_by_wf artifact").get() is None
    assert (
        latest_run.artifact("metric_enabled artifact").execution_state().status
        == ExecutionStatus.SUCCEEDED
    )
    assert latest_run.artifact("metric_enabled artifact").get() == 100
    assert (
        latest_run.artifact("metric_disabled artifact").execution_state().status
        == ExecutionStatus.DELETED
    )
    assert latest_run.artifact("metric_disabled artifact").get() is None
    assert (
        latest_run.artifact("check_enabled artifact").execution_state().status
        == ExecutionStatus.SUCCEEDED
    )
    assert latest_run.artifact("check_enabled artifact").get()


def test_artifact_set_name(client, flow_name, engine):
    @op
    def foo():
        return 123

    output = foo()
    assert output.name() == "foo artifact"

    output.set_name("bar")
    assert output.name() == "bar"

    # Check that the artifact can be fetched by the new name after publishing.
    flow = publish_flow_test(client, output, engine=engine, name=flow_name())
    assert flow.latest().artifact("bar").get() == 123


def test_operators_with_custom_output_names(client, flow_name, engine):
    @op(outputs=["output1", "output2"])
    def foo():
        return 123, "hello"

    output1, output2 = foo()
    assert output1.name() == "output1"
    assert output2.name() == "output2"

    @metric(output="metric output")
    def my_metric(input):
        return 99999

    @check(output="check output")
    def passed(input):
        return True

    m = my_metric(output1)
    assert m.name() == "metric output"

    c = passed(output2)
    assert c.name() == "check output"

    flow = publish_flow_test(
        client,
        artifacts=[output1, output2],
        name=flow_name(),
        engine=engine,
    )
    flow_run = flow.latest()

    assert flow_run.artifact("output1").get() == 123
    assert flow_run.artifact("output2").get() == "hello"
    assert flow_run.artifact("metric output").get() == 99999
    assert flow_run.artifact("check output").get()

    # Failure case: mismatches between num_outputs and len(outputs)
    with pytest.raises(InvalidUserArgumentException):

        @op(num_outputs=2, outputs=["output"])
        def bar():
            return 222


def test_get_flow_with_name(client, flow_name, engine):
    """Tests performing flow read operations using the flow name."""

    @op
    def noop():
        return 123

    output = noop()

    flow = publish_flow_test(
        client,
        artifacts=[output],
        name=flow_name(),
        engine=engine,
    )

    fetched_with_id = client.flow(flow.id())
    # fetched_with_name = client.flow(flow_identifier=flow.name())

    # assert fetched_with_id.id() == fetched_with_name.id()

    # Failure case: flow name does not match
    with pytest.raises(InvalidUserArgumentException):
        client.flow(flow_identifier="not a real flow")


def test_refresh_flow_with_name(client, flow_name, engine):
    """Tests triggering new run using the flow name."""

    @op
    def noop():
        return 123

    output = noop()

    flow = publish_flow_test(
        client,
        artifacts=[output],
        name=flow_name(),
        engine=engine,
    )

    # Failure case: name does not match
    with pytest.raises(InvalidUserArgumentException):
        client.trigger(flow_identifier="not a real flow")

    client.trigger(flow_identifier=flow.id())


def test_delete_flow_with_name(client, flow_name, engine):
    """Tests deleting flow using name."""

    @op
    def noop():
        return 123

    output = noop()

    flow = publish_flow_test(
        client,
        artifacts=[output],
        name=flow_name(),
        engine=engine,
    )

    # Failure case: flow id and name do not match
    with pytest.raises(InvalidUserArgumentException):
        client.delete_flow(flow_identifier="not a real flow")

    client.delete_flow(flow_identifier=flow.id())


@pytest.mark.enable_only_for_data_integration_type(*all_relational_DBs())
def test_flow_with_failed_compute_operators(
    client, flow_name, data_resource, engine, data_validator
):
    """
    Test if one or more compute operators fail, then the save/load operator does not succeed also.
    """

    @op
    def bar(arg):
        return 5 / 0

    @op
    def baz(arg):
        time.sleep(10)
        return arg

    table_name = generate_table_name()
    result = data_resource.sql("select * from hotel_reviews limit 5")
    test_data = bar.lazy(baz.lazy(result))
    save(data_resource, result, name=table_name, update_mode=LoadUpdateMode.REPLACE)

    publish_flow_test(
        client,
        artifacts=[test_data, result],
        name=flow_name(),
        engine=engine,
        expected_statuses=[ExecutionStatus.FAILED],
    )

    data_validator.check_saved_artifact_data_does_not_exist(result.id())
