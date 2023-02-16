from typing import Dict, List

import pandas as pd
import pytest
from aqueduct.artifacts.generic_artifact import GenericArtifact
from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.constants.enums import ExecutionStatus
from aqueduct.error import (
    AqueductError,
    ArtifactNeverComputedException,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from pandas._testing import assert_frame_equal
from PIL import Image

from aqueduct import metric, op

from ..shared.data_objects import DataObject
from ..shared.flow_helpers import publish_flow_test, trigger_flow_test
from ..shared.relational import all_relational_DBs
from .extract import extract


@metric
def double_number_input(num: int) -> float:
    if not isinstance(num, int):
        raise Exception("Expected an integer input.")
    return float(2 * num)


@metric
def len_of_word(word: str) -> int:
    if not isinstance(word, str):
        raise Exception("Expected a string input.")
    return len(word)


@op
def convert_dict_to_df(kv: Dict[str, List[float]]):
    return pd.DataFrame(data=kv)


def test_basic_param_creation(client):
    # Parameter of integer type
    param = client.create_param(name="number", default=8)
    assert param.get() == 8

    param_doubled = double_number_input(param)
    assert param_doubled.get() == 2 * 8

    # Parameter of string type
    param = client.create_param(name="word", default="hello world")
    assert param.get() == "hello world"

    param_length = len_of_word(param)
    assert param_length.get() == len("hello world")

    # Parameter of dictionary type
    kv = {"col 1": [1.23, 4.56], "col 2": [7.89, 1.23]}
    param = client.create_param(name="word", default=kv)
    assert param.get() == kv

    kv_df = convert_dict_to_df(param)
    # We don't use df.equals because when comparing floating point values, our internal serialization
    # may have changed the value's accuracy. assert_frame_equal takes this into account.
    assert_frame_equal(kv_df.get(), pd.DataFrame(data=kv))


def test_get_with_custom_parameter(client):
    param = client.create_param(name="number", default=8)
    assert param.get() == 8

    param_doubled = double_number_input(param)
    assert param_doubled.get(parameters={"number": 20}) == 40
    assert param_doubled.get() == 2 * 8

    with pytest.raises(InvalidUserArgumentException):
        param_doubled.get(parameters={"non-existant param": 10})

    # Check that changing the type of the parameter will error.
    with pytest.raises(AqueductError):
        param_doubled.get(parameters={"number": "NOT A NUMBER"})


def test_implicitly_created_parameter(client, flow_name, engine):
    @op
    def func(foo):
        return foo

    result = func(2)
    assert result.get() == 2
    assert result.get(parameters={"func:foo": 10}) == 10

    # Test with multiple outputs.
    @op(outputs=["output1", "output2"])
    def multi_output(param):
        return param + 1, param + 100

    output1, output2 = multi_output(100)
    assert output1.get() == 101
    assert output2.get() == 200

    # Check that we publish with the implicit parameter default values.
    flow = publish_flow_test(
        client, artifacts=[result, output1, output2], name=flow_name(), engine=engine
    )
    flow_run = flow.latest()

    assert flow_run.artifact("func:foo").get() == 2
    assert flow_run.artifact("multi_output:param").get() == 100
    assert flow_run.artifact("output1").get() == 101
    assert flow_run.artifact("output2").get() == 200


def test_implicitly_created_param_overwrites(client, flow_name, engine):
    @op
    def foo(param):
        return param

    # Test that two implicit parameters colliding will result in an override if they
    # are used by the same function.
    foo_output = foo(123)
    assert foo_output.get() == 123

    foo_output = foo("hello")
    assert foo_output.get() == "hello"
    assert foo_output.get({"foo:param": "custom val"}) == "custom val"

    # Test that two implicit parameters colliding with NOT result in an override if
    # they are used by different functions. In this case, is it because the names are
    # resolved to different values.
    @op
    def different_fn(param):
        return param

    different_fn_output = different_fn("different value")
    assert different_fn_output.get() == "different value"
    assert different_fn_output.get({"different_fn:param": "another val"}) == "another val"

    # Publish and validate the final value of each parameter.
    flow = publish_flow_test(
        client,
        artifacts=[foo_output, different_fn_output],
        name=flow_name(),
        engine=engine,
    )
    flow_run = flow.latest()

    assert flow_run.artifact("foo:param").get() == "hello"
    assert flow_run.artifact("different_fn:param").get() == "different value"


def test_multiple_implicitly_created_param(client):
    @op
    def foo(param1, param2):
        return param1 + param2

    assert foo(100, 50).get() == 150
    assert foo(500, 500).get() == 1000


def test_implicitly_created_param_failures(client):
    # Test that an implicit parameter colliding with a globally created parameter will error.
    @op
    def bar(param):
        return param

    _ = client.create_param("bar:param", default=200)
    with pytest.raises(
        InvalidUserActionException,
        match="there is an existing operator or artifact with the same name",
    ):
        _ = bar(300)

    # Same case as above, but actually attach the global parameter to the operator first.
    @op
    def baz(param):
        return param

    baz_param = client.create_param("baz:param", default=500)
    _ = baz(baz_param)

    with pytest.raises(
        InvalidUserActionException,
        match="there is an existing operator or artifact with the same name",
    ):
        baz(500)

    # Test that an implicit parameter can cannot collide with an existing operator.
    @op(name="qup:param")
    def colliding_fn():
        return 222

    _ = colliding_fn()

    @op
    def qup(param):
        return param

    with pytest.raises(
        InvalidUserActionException,
        match="there is an existing operator or artifact with the same name",
    ):
        _ = qup(500)

    # Test that an explicit parameter colliding with an implicit one will raise an exception.
    @op
    def another_fn(another_param):
        return another_param

    _ = another_fn("another string")
    with pytest.raises(
        InvalidUserActionException,
        match="there is an implicitly created parameter with the same name",
    ):
        client.create_param("another_fn:another_param", default="this should fail")


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


@op
def append_row_to_df(df, row):
    """`row` is a list of values to append to the input dataframe."""
    df.loc[len(df.index)] = row
    return df


def test_parameter_in_basic_flow(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    row_to_add = ["new hotel", "09-28-1996", "US", "It was new."]
    new_row_param = client.create_param(name="new row", default=row_to_add)
    output = append_row_to_df(table_artifact, new_row_param)

    input_df = table_artifact.get()
    input_df.loc[len(input_df.index)] = row_to_add

    output_df = output.get()
    assert output_df.equals(input_df)


def test_edit_param_for_flow(client, flow_name, data_integration, engine):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    row_to_add = ["new hotel", "09-28-1996", "US", "It was new."]
    new_row_param = client.create_param(name="new row", default=row_to_add)
    output = append_row_to_df(table_artifact, new_row_param)

    flow = publish_flow_test(
        client,
        output,
        name=flow_name(),
        engine=engine,
    )

    # Edit the flow with a different row to append and re-publish
    new_row_to_add = ["another new hotel", "10-10-1000", "ID", "It was really really new."]
    new_row_param = client.create_param(name="new row", default=new_row_to_add)
    output = append_row_to_df(table_artifact, new_row_param)

    flow = publish_flow_test(
        client,
        output,
        existing_flow=flow,
        engine=engine,
    )

    # Verify that the parameters were edited as expected.
    flow_runs = flow.list_runs()
    assert len(flow_runs) == 2

    historical_run = flow.fetch(flow_runs[1]["run_id"])
    param_artifact = historical_run.artifact(name="new row")
    assert isinstance(param_artifact, GenericArtifact)
    assert param_artifact.get() == row_to_add

    latest_run = flow.latest()
    param_artifact = latest_run.artifact(name="new row")
    assert isinstance(param_artifact, GenericArtifact)
    assert param_artifact.get() == new_row_to_add


@metric
def add_numbers(sql, num1, num2):
    if not isinstance(num1, int) or not isinstance(num2, int):
        raise Exception("Expected an integer input.")
    return num1 + num2


def test_trigger_flow_with_different_param(client, flow_name, data_integration, engine):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    num1 = client.create_param(name="num1", default=5)
    num2 = client.create_param(name="num2", default=5)
    output = add_numbers(table_artifact, num1, num2)

    flow = publish_flow_test(
        client,
        output,
        name=flow_name(),
        engine=engine,
    )

    # First, check that triggering the flow with a non-existant parameter will error.
    with pytest.raises(InvalidUserArgumentException):
        client.trigger(flow.id(), parameters={"non-existant": 10})

    trigger_flow_test(client, flow, parameters={"num1": 10})

    # Verify the parameters were configured as expected.
    flow_runs = flow.list_runs()
    assert len(flow_runs) == 2

    historical_run = flow.fetch(flow_runs[1]["run_id"])
    num1_artifact = historical_run.artifact(name="num1")
    num2_artifact = historical_run.artifact(name="num2")
    assert isinstance(num1_artifact, NumericArtifact)
    assert isinstance(num2_artifact, NumericArtifact)
    assert num1_artifact.get() == 5
    assert num2_artifact.get() == 5

    latest_run = flow.latest()
    num1_artifact = latest_run.artifact(name="num1")
    num2_artifact = latest_run.artifact(name="num2")
    assert isinstance(num1_artifact, NumericArtifact)
    assert isinstance(num2_artifact, NumericArtifact)
    assert num1_artifact.get() == 10
    assert num2_artifact.get() == 5


@pytest.mark.enable_only_for_data_integration_type(*all_relational_DBs())
def test_trigger_flow_with_different_sql_param(client, flow_name, data_integration, engine):
    _ = client.create_param("table_name", default="hotel_reviews")
    table_artifact = data_integration.sql(query="select * from {{ table_name}}")

    flow = publish_flow_test(
        client,
        table_artifact,
        name=flow_name(),
        engine=engine,
    )

    trigger_flow_test(client, flow, parameters={"table_name": "customer_activity"})

    # Verify the parameters were configured as expected.
    flow_runs = flow.list_runs()
    assert len(flow_runs) == 2

    historical_run = flow.fetch(flow_runs[1]["run_id"])
    param_artifact = historical_run.artifact(name="table_name")
    assert isinstance(param_artifact, GenericArtifact)
    assert param_artifact.get() == "hotel_reviews"

    latest_run = flow.latest()
    param_artifact = latest_run.artifact(name="table_name")
    assert param_artifact.get() == "customer_activity"
    assert isinstance(param_artifact, GenericArtifact)


def test_parameterizing_published_artifact(client, flow_name, engine):
    @op
    def generate_num():
        return 1234

    output = generate_num()

    flow = publish_flow_test(
        client,
        artifacts=[output],
        name=flow_name(),
        engine=engine,
    )

    artifact = flow.latest().artifact(name="generate_num artifact")

    assert artifact.get() == 1234
    assert isinstance(artifact, NumericArtifact)
    with pytest.raises(NotImplementedError):
        artifact.get(parameters={"name": "val"})


def test_materializing_failed_artifact(client, flow_name, engine):
    @op
    def fail_fn():
        5 / 0

    output = fail_fn.lazy()

    flow = publish_flow_test(
        client,
        output,
        name=flow_name(),
        engine=engine,
        expected_statuses=ExecutionStatus.FAILED,
    )
    artifact = flow.latest().artifact(name="fail_fn artifact")
    assert isinstance(artifact, GenericArtifact)
    with pytest.raises(ArtifactNeverComputedException):
        artifact.get()


def test_all_param_types(client, flow_name, engine):
    class EmptyClass:
        """
        For some reason, this class must be nested inside this test,
        otherwise we get a pickle error on the backend: 'No module named `param_test`'.
        """

        def __init__(self):
            pass

    @op
    def must_be_picklable(input):
        """
        Unable to check that the input is picklable, since `pickle.loads()`
        complains about `import of module 'param_test' failed`.
        """
        if input != EmptyClass:
            raise Exception("Expected Class.")
        return input

    picklable_param = client.create_param("pickleable", default=EmptyClass)
    pickle_output = must_be_picklable(picklable_param)

    assert isinstance(pickle_output, GenericArtifact)
    assert pickle_output.get() == EmptyClass

    @op
    def must_be_bytes(input):
        if not isinstance(input, bytes):
            raise Exception("Expected bytes")
        return input

    bytes_param = client.create_param("bytes", default=b"hello world")
    bytes_output = must_be_bytes(bytes_param)

    assert isinstance(bytes_output, GenericArtifact)
    assert bytes_output.get() == b"hello world"

    @op
    def must_be_string(input):
        if not isinstance(input, str):
            raise Exception("Expected string.")
        return input

    string_param = client.create_param("string", default="I am a string")
    string_output = must_be_string(string_param)
    assert isinstance(string_output, GenericArtifact)
    assert string_output.get() == "I am a string"

    @op
    def must_be_tuple(input):
        if not isinstance(input, tuple):
            raise Exception("Expected tuple.")
        return input

    tuple_param = client.create_param("tuple", default=(1, 2, 3))
    tuple_output = must_be_tuple(tuple_param)
    assert isinstance(tuple_output, GenericArtifact)
    assert tuple_output.get() == (1, 2, 3)

    @op
    def must_be_list(input):
        if not isinstance(input, list):
            raise Exception("Expected list.")
        return input

    list_param = client.create_param("list", default=[4, 5, 6])
    list_output = must_be_list(list_param)
    assert isinstance(list_output, GenericArtifact)
    assert list_output.get() == [4, 5, 6]

    @op
    def must_be_image(input):
        if not isinstance(input, Image.Image):
            raise Exception("Expected image.")
        return input

    # Current working directory is one level above.
    image_data = Image.open("aqueduct_tests/data/aqueduct.jpg", "r")

    image_param = client.create_param("image", default=image_data)
    image_output = must_be_image(image_param)
    assert isinstance(image_output, GenericArtifact)
    assert isinstance(image_output.get(), Image.Image)

    publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[
            pickle_output,
            bytes_output,
            string_output,
            tuple_output,
            list_output,
            image_output,
        ],
        engine=engine,
    )


def test_parameter_type_changes(client, flow_name, engine):
    @op
    def noop(input):
        return input

    param = client.create_param("number", default=1234)
    output = noop(param)

    # TODO(ENG-1684): This should be a more specific error.
    with pytest.raises(Exception):
        output.get(parameters={"number": "This is a string."})

    flow = publish_flow_test(
        client,
        output,
        name=flow_name(),
        engine=engine,
    )

    # TODO(ENG-1684): we should not allow the user to trigger successfully with the wrong type.
    trigger_flow_test(
        client,
        flow,
        expected_status=ExecutionStatus.FAILED,
        parameters={"number": "This is a string"},
    )


def test_param_management(client):
    # Create some implicit parameters that are consumed by downstream operators.
    @op
    def foo(param1, param2):
        return 123

    foo_output = foo(1000, "string val")

    @op
    def bar(input):
        return input

    bar_output = bar(foo_output)

    # Create a global parameter with no attachments.
    client.create_param("param3", default="content")

    assert client.list_params() == {
        "foo:param1": 1000,
        "foo:param2": "string val",
        "param3": "content",
    }

    # Delete the unattached parameter.
    client.delete_param("param3")
    assert client.list_params() == {
        "foo:param1": 1000,
        "foo:param2": "string val",
    }

    # Delete one of the attached parameters.
    with pytest.raises(InvalidUserActionException, match="Cannot delete parameter"):
        client.delete_param("foo:param1")

    client.delete_param("foo:param1", force=True)
    client.delete_param("foo:param2", force=True)
    assert client.list_params() == {}

    # Check that bar_output is now invalid
    with pytest.raises(Exception):
        bar_output.get()
