from typing import Dict, List

import pandas as pd
import pytest
from aqueduct.artifacts.generic_artifact import GenericArtifact
from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.enums import ExecutionStatus
from aqueduct.error import (
    AqueductError,
    ArtifactNeverComputedException,
    InvalidUserArgumentException,
)
from constants import SENTIMENT_SQL_QUERY
from pandas._testing import assert_frame_equal
from utils import delete_flow, generate_new_flow_name, run_flow_test, wait_for_flow_runs

from aqueduct import metric, op


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


def test_implicitly_created_parameter(client):
    @op
    def func(foo):
        return foo

    result = func(2)
    assert result.get() == 2
    assert result.get(parameters={"foo": 10}) == 10

    bar_param = client.create_param("bar", default="hello")

    @op
    def another_func(bar):
        return bar

    with pytest.raises(InvalidUserArgumentException):
        # This should error because we try to implicitly create a parameter "bar" but it already exists.
        result = another_func(2)

    result = another_func(bar_param)
    assert result.get() == "hello"


@op
def append_row_to_df(df, row):
    """`row` is a list of values to append to the input dataframe."""
    df.loc[len(df.index)] = row
    return df


def test_parameter_in_basic_flow(client, data_integration):
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    row_to_add = ["new hotel", "09-28-1996", "US", "It was new."]
    new_row_param = client.create_param(name="new row", default=row_to_add)
    output = append_row_to_df(sql_artifact, new_row_param)

    input_df = sql_artifact.get()
    input_df.loc[len(input_df.index)] = row_to_add

    output_df = output.get()
    assert output_df.equals(input_df)


def test_edit_param_for_flow(client, data_integration):
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    row_to_add = ["new hotel", "09-28-1996", "US", "It was new."]
    new_row_param = client.create_param(name="new row", default=row_to_add)
    output = append_row_to_df(sql_artifact, new_row_param)

    flow_name = generate_new_flow_name()

    flow_id = None
    try:
        flow = run_flow_test(client, artifacts=[output], name=flow_name, delete_flow_after=False)
        flow_id = flow.id()

        # Edit the flow with a different row to append and re-publish
        new_row_to_add = ["another new hotel", "10-10-1000", "ID", "It was really really new."]
        new_row_param = client.create_param(name="new row", default=new_row_to_add)
        output = append_row_to_df(sql_artifact, new_row_param)

        # Wait for the first run, then refresh the workflow and verify that it runs at least
        # one more time (two runs total, since the original was manually triggered).
        flow = run_flow_test(
            client, artifacts=[output], name=flow_name, num_runs=2, delete_flow_after=False
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

    finally:
        client.delete_flow(flow_id)

    assert flow_id == flow.id()


@metric
def add_numbers(sql, num1, num2):
    if not isinstance(num1, int) or not isinstance(num2, int):
        raise Exception("Expected an integer input.")
    return num1 + num2


def test_trigger_flow_with_different_param(client, data_integration):
    db = client.integration(data_integration)
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)

    num1 = client.create_param(name="num1", default=5)
    num2 = client.create_param(name="num2", default=5)
    output = add_numbers(sql_artifact, num1, num2)

    flow_name = generate_new_flow_name()
    flow = run_flow_test(client, artifacts=[output], name=flow_name, delete_flow_after=False)

    # First, check that triggering the flow with a non-existant parameter will error.
    with pytest.raises(InvalidUserArgumentException):
        client.trigger(flow.id(), parameters={"non-existant": 10})

    try:
        client.trigger(flow.id(), parameters={"num1": 10})
        wait_for_flow_runs(
            client,
            flow.id(),
            expect_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
        )

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

    finally:
        client.delete_flow(flow.id())


def test_trigger_flow_with_different_sql_param(client, data_integration):
    db = client.integration(data_integration)

    _ = client.create_param("table_name", default="hotel_reviews")
    sql_artifact = db.sql(query="select * from {{ table_name}}")

    flow_name = generate_new_flow_name()

    flow_id = None
    try:
        flow = run_flow_test(
            client, artifacts=[sql_artifact], name=flow_name, delete_flow_after=False
        )
        flow_id = flow.id()

        client.trigger(flow.id(), parameters={"table_name": "customer_activity"})
        wait_for_flow_runs(
            client,
            flow.id(),
            expect_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.SUCCEEDED],
        )

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
    finally:
        client.delete_flow(flow_id)


def test_parameterizing_published_artifact(client):
    @op
    def generate_num():
        return 1234

    output = generate_num()

    flow_id = None
    try:
        flow = run_flow_test(client, artifacts=[output], delete_flow_after=False)
        flow_id = flow.id()

        artifact = flow.latest().artifact(name="generate_num artifact")

        assert artifact.get() == 1234
        assert isinstance(artifact, NumericArtifact)
        with pytest.raises(NotImplementedError):
            artifact.get(parameters={"name": "val"})

    finally:
        client.delete_flow(flow_id)


def test_materializing_failed_artifact(client):
    @op
    def fail_fn():
        5 / 0

    output = fail_fn.lazy()
    flow_id = None
    try:
        flow = run_flow_test(
            client, artifacts=[output], expect_success=False, delete_flow_after=False
        )
        flow_id = flow.id()

        artifact = flow.latest().artifact(name="fail_fn artifact")
        assert isinstance(artifact, GenericArtifact)
        with pytest.raises(ArtifactNeverComputedException):
            artifact.get()

    finally:
        client.delete_flow(flow_id)


def test_all_param_types(client):
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
        Unable to check that the input is pickleabe, since `pickle.loads()`
        complains about `import of module 'param_test' failed`.
        """
        assert input == EmptyClass
        return input

    picklable_param = client.create_param("pickleable", default=EmptyClass)
    pickle_output = must_be_picklable(picklable_param)

    assert isinstance(pickle_output, GenericArtifact)
    assert pickle_output.get() == EmptyClass

    @op
    def must_be_bytes(input):
        assert isinstance(input, bytes)
        return input

    bytes_param = client.create_param("bytes", default=b"hello world")
    bytes_output = must_be_bytes(bytes_param)

    assert isinstance(bytes_output, GenericArtifact)
    assert bytes_output.get() == b"hello world"

    @op
    def must_be_string(input):
        assert isinstance(input, str)
        return input

    string_param = client.create_param("string", default="I am a string")
    string_output = must_be_string(string_param)
    assert isinstance(string_output, GenericArtifact)
    assert string_output.get() == "I am a string"

    @op
    def must_be_tuple(input):
        assert isinstance(input, tuple)
        return input

    tuple_param = client.create_param("tuple", default=(1, 2, 3))
    tuple_output = must_be_tuple(tuple_param)
    assert isinstance(tuple_output, GenericArtifact)
    assert tuple_output.get() == (1, 2, 3)

    @op
    def must_be_list(input):
        assert isinstance(input, list)
        return input

    list_param = client.create_param("list", default=[4, 5, 6])
    list_output = must_be_list(list_param)
    assert isinstance(list_output, GenericArtifact)
    assert list_output.get() == [4, 5, 6]

    run_flow_test(client, artifacts=[pickle_output, bytes_output, string_output, tuple_output, list_output])


def test_parameter_type_changes(client):
    @op
    def noop(input):
        return input

    param = client.create_param("number", default=1234)
    output = noop(param)

    # TODO(ENG-1684): This should be a more specific error.
    with pytest.raises(Exception):
        output.get(parameters={"number": "This is a string."})

    flow_id = None
    try:
        flow = run_flow_test(client, artifacts=[output], delete_flow_after=False)
        flow_id = flow.id()

        # TODO(ENG-1684): we should not allow the user to trigger successfully with the wrong type.
        client.trigger(flow_id, parameters={"number": "This is a string"})
        wait_for_flow_runs(
            client,
            flow_id,
            expect_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.FAILED],
        )
    finally:
        delete_flow(client, flow_id)
