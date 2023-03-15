import pytest
from aqueduct.error import ArtifactNotFoundException, InvalidUserActionException

from aqueduct import check, metric, op

from ..shared.data_objects import DataObject
from ..shared.flow_helpers import publish_flow_test
from .extract import extract
from .test_functions.simple.model import (
    dummy_model,
    dummy_model_2,
    dummy_sentiment_model,
    dummy_sentiment_model_multiple_input,
)


def test_artifact_name_sanitization(client, data_integration):
    "Checks that whitespace in the beginning or end of artifact names is removed."
    param = client.create_param("   whitespace around me  ", default=123)
    assert param.name() == "whitespace around me"

    table = extract(data_integration, DataObject.SENTIMENT, output_name="   whitespace around me  ")
    assert table.name() == "whitespace around me"

    @op(outputs=["  whitespace around me  "])
    def foo():
        return 123
    output = foo()
    assert output.name() == "whitespace around me"

    # Not even .set_name() can save you.
    output.set_name("   whitespace around me  ")
    assert output.name() == "whitespace around me"

    @metric(output="  whitespace around me  ")
    def m(input):
        return 100
    assert m(output).name() == "whitespace around me"

    @check(output="  whitespace around me  ")
    def c(input):
        return True
    assert c(output).name() == "whitespace around me"


def test_extract_with_default_name_collision(client, flow_name, engine, data_integration):
    # In the case where no explicit name is supplied, we expect new extract
    # operators to always be created.
    table_artifact_1 = extract(data_integration, DataObject.SENTIMENT)
    table_artifact_2 = extract(data_integration, DataObject.SENTIMENT)

    assert table_artifact_1.name() == "%s query artifact" % data_integration.name()
    assert table_artifact_2.name() == "%s query artifact" % data_integration.name()

    fn_artifact = dummy_sentiment_model_multiple_input(table_artifact_1, table_artifact_2)
    fn_df = fn_artifact.get()
    assert list(fn_df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
        "positivity",
        "positivity_2",
    ]
    assert fn_df.shape[0] == 100

    # Check that the names were properly deduplicated at publish time.
    flow = publish_flow_test(client, artifacts=[fn_artifact], engine=engine, name=flow_name())
    flow_run = flow.latest()

    # They both have the same data, but the order shouldn't matter.
    assert flow_run.artifact(table_artifact_1.name()).get().equals(table_artifact_1.get())
    assert flow_run.artifact(table_artifact_1.name() + " (1)").get().equals(table_artifact_1.get())


def test_extract_with_op_name_collision(client, data_integration, engine, flow_name):
    """Artifact names are the only collisions we care about. We will deduplicate them, but allow
    for operator name duplicates."""
    table_artifact_1 = extract(data_integration, DataObject.SENTIMENT, op_name="sql query")
    assert table_artifact_1.name() == "sql query artifact"

    table_artifact_2 = extract(data_integration, DataObject.SENTIMENT, op_name="sql query")
    assert table_artifact_2.name() == "sql query artifact"

    # Check that the old operator still exists and works.
    table_1 = table_artifact_1.get()
    table_2 = table_artifact_2.get()
    assert table_1.equals(table_2)
    assert list(table_1) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]
    assert table_1.shape[0] == 100

    flow = publish_flow_test(
        client, artifacts=[table_artifact_1, table_artifact_2], engine=engine, name=flow_name()
    )
    flow_run = flow.latest()
    assert flow_run.artifact("sql query artifact").get().equals(table_1)
    assert flow_run.artifact("sql query artifact (1)").get().equals(table_1)


def test_extract_with_artifact_name_collision(client, data_integration, engine, flow_name):
    output = extract(data_integration, DataObject.SENTIMENT, output_name="hotel reviews")
    assert output.name() == "hotel reviews"

    # Test that custom artifact naming works.
    flow = publish_flow_test(client, artifacts=output, engine=engine, name=flow_name())
    assert flow.latest().artifact("hotel reviews").get().equals(output.get())

    # We can name another output artifact the same, but we can't publish the two together.
    output2 = extract(data_integration, DataObject.SENTIMENT, output_name="hotel reviews")
    with pytest.raises(
        InvalidUserActionException,
        match="Unable to publish flow. You are attempting to publish multiple artifacts explicitly named",
    ):
        client.publish_flow("Test", artifacts=[output, output2], engine=engine)


def test_operator_with_default_artifact_naming_collision(client, engine, flow_name):
    """Also tests that reusing the same operator twice in a flow works as expected."""

    @op(num_outputs=2)
    def foo():
        return 123, "hello"

    output1, output2 = foo()
    output3, output4 = foo()
    assert output1.name() == "foo artifact"
    assert output2.name() == "foo artifact"
    assert output3.name() == "foo artifact"
    assert output4.name() == "foo artifact"

    flow = publish_flow_test(
        client, artifacts=[output1, output2, output3, output4], engine=engine, name=flow_name()
    )
    flow_run = flow.latest()
    assert flow_run.artifact("foo artifact").get() == 123
    assert flow_run.artifact("foo artifact (1)").get() == "hello"
    assert flow_run.artifact("foo artifact (2)").get() == 123
    assert flow_run.artifact("foo artifact (3)").get() == "hello"


def test_operator_with_explicit_artifact_naming_collision(client, engine, flow_name):
    @op(outputs=["output1", "output2"])
    def foo():
        return 123, "hello"

    output1, output2 = foo()
    output3, output4 = foo()

    with pytest.raises(
        InvalidUserActionException,
        match="Unable to publish flow. You are attempting to publish multiple artifacts explicitly named `output1`",
    ):
        client.publish_flow("Test", artifacts=[output1, output2, output3, output4], engine=engine)

    # Trigger another collision case with the artifact.set_name() method.
    @op
    def bar():
        return 555

    bar_output = bar()
    bar_output.set_name("output1")
    with pytest.raises(
        InvalidUserActionException,
        match="Unable to publish flow. You are attempting to publish multiple artifacts explicitly named `output1`",
    ):
        client.publish_flow("Test", artifacts=[output1, bar_output], engine=engine)


def test_explicit_and_implicit_artifact_name_collisions(client, engine, flow_name):
    """Test that if such a collision occurs, the explicit name always wins."""

    @op
    def foo():
        return 123

    @op(outputs=["foo artifact"])
    def bar():
        return "hello"

    foo_output = foo()
    bar_output = bar()

    # Regardless of which order we publish them, the explicit name should always win.
    flow = publish_flow_test(
        client, artifacts=[foo_output, bar_output], engine=engine, name=flow_name()
    )
    flow_run = flow.latest()
    assert flow_run.artifact("foo artifact").get() == "hello"  # bar's output
    assert flow_run.artifact("foo artifact (1)").get() == 123  # foo's output

    flow = publish_flow_test(
        client, artifacts=[bar_output, foo_output], engine=engine, name=flow_name()
    )
    flow_run = flow.latest()
    assert flow_run.artifact("foo artifact").get() == "hello"  # bar's output
    assert flow_run.artifact("foo artifact (1)").get() == 123  # foo's output


def _run_noop_op(table_output):
    """Returns artifact with name `foo artifact`"""

    @op
    def foo(table):
        return table

    return foo(table_output)


def _run_noop_metric(input):
    """Returns artifact with name `foo artifact`"""

    @metric
    def foo(input):
        return 100

    return foo(input)


def _run_noop_check(input):
    """Returns artifact with name `foo artifact`"""

    @check
    def foo(input):
        return True

    return foo(input)


def test_artifact_name_collisions_across_operator_types(
    client, data_integration, engine, flow_name
):
    """Tests that the same naming policy holds regardless of the operator type."""
    extract_output = extract(data_integration, DataObject.SENTIMENT, output_name="foo artifact")
    op_output = _run_noop_op(extract_output)
    metric_output = _run_noop_metric(op_output)
    check_output = _run_noop_check(metric_output)

    assert extract_output.name() == "foo artifact"
    assert op_output.name() == "foo artifact"
    assert metric_output.name() == "foo artifact"
    assert check_output.name() == "foo artifact"

    # The artifact's should be bumped in the following order: extract (explicitly named), op, metric, check.
    flow = publish_flow_test(client, artifacts=[op_output], engine=engine, name=flow_name())
    flow_run = flow.latest()
    assert flow_run.artifact("foo artifact").get().equals(extract_output.get())
    assert flow_run.artifact("foo artifact (1)").get().equals(extract_output.get())
    assert flow_run.artifact("foo artifact (2)").get() == 100
    assert flow_run.artifact("foo artifact (3)").get() == True

    # Making any of the downstream operators explicit should error at publish time due to new collision against the extract artifact.
    metric_output.set_name("foo artifact")
    with pytest.raises(
        InvalidUserActionException,
        match="Unable to publish flow. You are attempting to publish multiple artifacts explicitly named `foo artifact`",
    ):
        client.publish_flow("Test", artifacts=op_output, engine=engine)


def test_param_naming_collisions(client, engine, flow_name):
    # Two explicit parameter names colliding is not allowed.
    param1 = client.create_param("foo:param", default=123)
    param2 = client.create_param("foo:param", default="hello")

    with pytest.raises(
        InvalidUserActionException,
        match="Unable to publish flow. You are attempting to publish multiple artifacts explicitly named `foo:param`",
    ):
        client.publish_flow("Test", artifacts=[param1, param2], engine=engine)

    # An implicit parameter is allowed to collide with an explicit one, however.
    @op
    def foo(param):
        return param

    foo_output = foo("hello")  # creates an implicit parameter named "foo:param"

    flow = publish_flow_test(
        client, artifacts=[foo_output, param1], engine=engine, name=flow_name()
    )
    flow_run = flow.latest()
    assert flow_run.artifact("foo:param").get() == 123
    assert flow_run.artifact("foo:param (1)").get() == "hello"

    # Two implicit parameters are allowed and will be deduplicated at publish time.
    @op
    def bar(param):
        return param

    output1 = bar(100)
    output2 = bar(200)
    flow = publish_flow_test(client, artifacts=[output1, output2], engine=engine, name=flow_name())
    flow_run = flow.latest()
    assert flow_run.artifact("bar:param").get() in [100, 200]
    assert flow_run.artifact("bar:param (1)").get() in [100, 200]


def test_change_param_artifact_name(client, flow_name, engine):
    """Test that changing a parameter artifact name is possible."""
    param = client.create_param("param", default=123)
    param.set_name("new param name")
    new_param = param  # Move the parameter to a different variable

    # The operator name collides with the old param name, but we already moved it out.
    @op
    def param():
        return "value"

    fn_output = param()

    flow = publish_flow_test(
        client, artifacts=[new_param, fn_output], name=flow_name(), engine=engine
    )
    flow_run = flow.latest()
    assert flow_run.artifact("new param name").get() == 123
    assert flow_run.artifact("param artifact").get() == "value"
