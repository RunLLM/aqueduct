import uuid
from datetime import datetime, timedelta

import pandas as pd
import pytest
from aqueduct.constants.enums import ExecutionStatus, ServiceType
from aqueduct.error import InvalidRequestError, InvalidUserArgumentException
from aqueduct.integrations.airflow_integration import AirflowIntegration
from aqueduct.models.config import FlowConfig
from aqueduct.models.integration import IntegrationInfo

import aqueduct
from aqueduct import check, metric, op

from ..shared.data_objects import DataObject
from ..shared.flow_helpers import publish_flow_test, trigger_flow_test, wait_for_flow_runs
from ..shared.naming import generate_new_flow_name, generate_table_name
from .extract import extract
from .save import save
from .test_functions.sentiment.model import sentiment_model
from .test_functions.simple.model import (
    dummy_model,
    dummy_sentiment_model,
    dummy_sentiment_model_multiple_input,
)
from .test_metrics.constant.model import constant_metric


def test_basic_flow(client, flow_name, data_integration, engine, data_validator):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_integration, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
    )

    data_validator.check_saved_artifact_data(
        flow, output_artifact.id(), expected_data=output_artifact.get()
    )


def test_sentiment_flow(client, flow_name, data_integration, engine, data_validator):
    """Actually run the full sentiment model (with nltk dependency)."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = sentiment_model(table_artifact)
    save(data_integration, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
    )
    data_validator.check_saved_artifact_data(
        flow, output_artifact.id(), expected_data=output_artifact.get()
    )


def test_complex_flow(client, flow_name, data_integration, engine, data_validator):
    table_artifact1 = extract(data_integration, DataObject.SENTIMENT)
    table_artifact2 = extract(data_integration, DataObject.SENTIMENT)

    fn_artifact = dummy_sentiment_model_multiple_input(table_artifact1, table_artifact2)
    output_artifact = dummy_model(fn_artifact)
    save(data_integration, output_artifact)

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


def test_multiple_output_artifacts(client, flow_name, data_integration, engine):
    table_artifact1 = extract(data_integration, DataObject.SENTIMENT)
    table_artifact2 = extract(data_integration, DataObject.SENTIMENT)

    fn_artifact1 = dummy_sentiment_model(table_artifact1)
    fn_artifact2 = dummy_model(table_artifact2)
    save(data_integration, fn_artifact1)
    save(data_integration, fn_artifact2)

    publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[fn_artifact1, fn_artifact2],
        engine=engine,
    )


def test_publish_with_schedule(client, flow_name, data_integration, engine):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_integration, output_artifact)

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


def test_publish_flow_with_cascading_trigger(client, flow_name, data_integration, engine):
    """Tests publishing a flow that is set to run on a cascading trigger."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
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


def test_publish_with_schedule_and_source_flow(client, flow_name, data_integration, engine):
    """Tests publishing an invalid flow that has both a schedule and a source flow."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
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


def test_publish_with_source_flow_cyclic(client, flow_name, data_integration, engine):
    """Tests publishing an invalid flow, because it would cause a cycle amongst cascading workflows."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

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


def test_publish_flow_with_same_name(client, flow_name, data_integration, engine):
    """Tests flow editing behavior."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
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


def test_refresh_flow(client, flow_name, data_integration, engine):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_integration, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
        schedule=aqueduct.hourly(),
    )

    # Trigger the workflow again verify that it runs one more time.
    trigger_flow_test(
        client,
        flow,
    )


def test_get_artifact_from_flow(client, flow_name, data_integration, engine):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_integration, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
    )

    artifact_return = flow.latest().artifact(output_artifact.name())
    assert artifact_return.name() == output_artifact.name()
    assert artifact_return.get().equals(output_artifact.get())


def test_get_artifact_reuse_for_computation(client, flow_name, data_integration, engine):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    save(data_integration, output_artifact)

    flow = publish_flow_test(
        client,
        output_artifact,
        name=flow_name(),
        engine=engine,
    )

    artifact_return = flow.latest().artifact(output_artifact.name())
    with pytest.raises(Exception):
        output_artifact = dummy_sentiment_model(artifact_return)


def test_multiple_flows_with_same_schedule(client, flow_name, data_integration, engine):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
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


def test_fetching_historical_flows_uses_old_data(client, flow_name, data_integration, engine):
    # Write a new table into the data integration.
    initial_table = pd.DataFrame([1, 2, 3, 4, 5, 6], columns=["numbers"])

    @op
    def generate_initial_table():
        return initial_table

    table = generate_initial_table()
    saved_table_identifier = generate_table_name()
    save(data_integration, table, name=saved_table_identifier)

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
    output = extract(data_integration, saved_table_identifier, op_name="Test Table Query")
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
    save(data_integration, table, name=saved_table_identifier)
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

# TODO
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


def test_flow_list_saved_objects_none(client, flow_name, engine):
    """Check that flow.list_saved_objects() works when no objects were actually saved."""

    @op
    def noop():
        return 123

    output = noop()
    flow = publish_flow_test(client, artifacts=output, name=flow_name(), engine=engine)
    assert len(flow.list_saved_objects()) == 0
