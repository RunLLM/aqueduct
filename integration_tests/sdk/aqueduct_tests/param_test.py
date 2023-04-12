import os
import pickle
from typing import Dict, List

import pandas as pd
import pytest
from aqueduct.artifacts.generic_artifact import GenericArtifact
from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.constants.enums import ArtifactType, ExecutionStatus,ServiceType
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


@pytest.mark.skip_for_spark_engines(
    reason="Expect a Spark Dataframe as return type, not Pandas Dataframe."
)
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

    @op
    def func(a, b="world"):
        return a + " " + b

    with pytest.raises(
        InvalidUserArgumentException,
        match="No input was provided for argument `a` of function `func`, and no default value was specified.",
    ):
        result = func()

    result = func("hello")
    assert result.get() == "hello world"

    flow = publish_flow_test(client, artifacts=[result], name=flow_name(), engine=engine)
    flow_run = flow.latest()
    assert flow_run.artifact("func:a").get() == "hello"
    # Test that we implicitly created a parameter called "func:b" with the default value "world".
    assert flow_run.artifact("func:b").get() == "world"


@op
def append_row_to_df(df, row):
    """`row` is a list of values to append to the input dataframe."""
    df.loc[len(df.index)] = row
    return df


@pytest.mark.skip_for_spark_engines(reason="append_row_to_df doesn't work for Spark Dataframes")
def test_parameter_in_basic_flow(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    row_to_add = ["new hotel", "09-28-1996", "US", "It was new."]
    new_row_param = client.create_param(name="new row", default=row_to_add)
    output = append_row_to_df(table_artifact, new_row_param)

    input_df = table_artifact.get()
    input_df.loc[len(input_df.index)] = row_to_add

    output_df = output.get()
    assert output_df.equals(input_df)


@pytest.mark.skip_for_spark_engines(reason="append_row_to_df doesn't work for Spark Dataframes")
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
    table_name_param = client.create_param("table_name", default="hotel_reviews")
    table_artifact = data_integration.sql(query="select * from $1", parameters=[table_name_param])

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
    image_data = Image.open("data/aqueduct.jpg", "r")

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


@pytest.mark.skip_for_spark_engines(reason="append_row_to_df doesn't work for Spark Dataframes")
def test_local_table_data_parameter(client, flow_name, engine):
    row_to_add = ["new hotel", "09-28-1996", "US", "It was new."]

    file_type = ["csv", "json", "parquet"]
    output_artifact_list = []
    input_data_list = []
    for extension in file_type:
        data_param = client.create_param(
            name="data_" + extension,
            default="data/hotel_reviews." + extension,
            use_local=True,
            as_type=ArtifactType.TABLE,
            format=extension,
        )

        if extension == "csv":
            input_df = pd.read_csv("data/hotel_reviews." + extension)
        elif extension == "json":
            input_df = pd.read_json("data/hotel_reviews." + extension, orient="table")
        else:
            input_df = pd.read_parquet("data/hotel_reviews." + extension)
        assert input_df.equals(data_param.get())

        @op(name=extension, outputs=["output_" + extension])
        def append_row_to_df(df, row):
            """`row` is a list of values to append to the input dataframe."""
            df.loc[len(df.index)] = row
            return df

        output = append_row_to_df(data_param, row_to_add)
        input_df.loc[len(input_df.index)] = row_to_add
        output_df = output.get()
        assert output_df.equals(input_df)

        output_artifact_list.append(output)
        input_data_list.append(input_df)

    with pytest.raises(
        InvalidUserActionException,
        match="Cannot create a flow with local data. Consider setting `use_local` to True to publish a workflow with local data parameters.",
    ):
        flow = client.publish_flow(
            name=flow_name(),
            artifacts=output_artifact_list,
            engine=engine,
        )
    flow = publish_flow_test(
        client, artifacts=output_artifact_list, name=flow_name(), engine=engine, use_local=True
    )
    flow_run = flow.latest()
    assert flow_run.artifact("output_csv").get().equals(input_data_list[0])
    assert flow_run.artifact("output_json").get().equals(input_data_list[1])
    assert flow_run.artifact("output_parquet").get().equals(input_data_list[2])


def test_invalid_local_data(client):
    # check Local Data with file path that does not exist will fail
    with pytest.raises(
        InvalidUserArgumentException,
        match="Given path file 'data/hotel_reviews' to local data does not exist.",
    ):
        client.create_param(
            name="data", default="data/hotel_reviews", use_local=True, as_type=ArtifactType.IMAGE
        )

    # Check that format is supplied when Artifact type is table
    with pytest.raises(
        InvalidUserArgumentException,
        match="Specify format in order to use local data as TableArtifact.",
    ):
        client.create_param(
            name="data",
            default="data/hotel_reviews.json",
            use_local=True,
            as_type=ArtifactType.TABLE,
        )

# TODO: Remove this pytest fixture on next release.
@pytest.mark.enable_only_for_engine_type(ServiceType.AQUEDUCT_ENGINE)
def test_all_local_data_types(client, flow_name, engine):
    @op
    def must_be_picklable(input):
        """
        Unable to check that the input is picklable, since `pickle.loads()`
        complains about `import of module 'param_test' failed`.
        """
        if input != ArtifactType:
            raise Exception("Expected Class.")
        return input

    with open("data/test.pickle", "wb") as file:
        pickle.dump(ArtifactType, file)
    picklable_param = client.create_param(
        "pickleable", default="data/test.pickle", use_local=True, as_type=ArtifactType.PICKLABLE
    )
    pickle_output = must_be_picklable(picklable_param)

    assert isinstance(pickle_output, GenericArtifact)
    assert pickle_output.get() == ArtifactType

    @op
    def must_be_bytes(input):
        if not isinstance(input, bytes):
            raise Exception("Expected bytes")
        return input

    bytes_param = client.create_param(
        "bytes", default="data/test_bytes.txt", use_local=True, as_type=ArtifactType.BYTES
    )
    bytes_output = must_be_bytes(bytes_param)

    assert isinstance(bytes_output, GenericArtifact)
    assert bytes_output.get() == b"hello world"

    @op
    def must_be_string(input):
        if not isinstance(input, str):
            raise Exception("Expected string.")
        return input

    string_param = client.create_param(
        "string", default="data/test_bytes.txt", use_local=True, as_type=ArtifactType.STRING
    )
    string_output = must_be_string(string_param)
    assert isinstance(string_output, GenericArtifact)
    assert string_output.get() == "hello world"

    @op
    def must_be_tuple(input):
        if not isinstance(input, tuple):
            raise Exception("Expected tuple.")
        return input

    tuple_param = client.create_param(
        "tuple", default="data/test_tuple", use_local=True, as_type=ArtifactType.TUPLE
    )
    tuple_output = must_be_tuple(tuple_param)
    assert isinstance(tuple_output, GenericArtifact)
    assert tuple_output.get() == ("hello", "world")

    @op
    def must_be_list(input):
        if not isinstance(input, list):
            raise Exception("Expected list.")
        return input

    list_param = client.create_param(
        "list", default="data/test_list", use_local=True, as_type=ArtifactType.LIST
    )
    list_output = must_be_list(list_param)
    assert isinstance(list_output, GenericArtifact)
    assert list_output.get() == ["hello", "world"]

    @op
    def must_be_image(input):
        if not isinstance(input, Image.Image):
            raise Exception("Expected image.")
        return input

    image_param = client.create_param(
        "image", default="data/aqueduct.jpg", use_local=True, as_type=ArtifactType.IMAGE
    )
    image_output = must_be_image(image_param)
    assert isinstance(image_output, GenericArtifact)
    assert isinstance(image_output.get(), Image.Image)

    from tensorflow import keras

    model = keras.models.load_model("data/tf_model")

    @op
    def must_be_tf_keras(input):
        if not isinstance(input, keras.Model):
            raise Exception("Tensorflow keras model config does not match.")
        return input

    tf_keras_param = client.create_param(
        "tf_keras", default="data/tf_model", use_local=True, as_type=ArtifactType.TF_KERAS
    )
    tf_keras_output = must_be_tf_keras(tf_keras_param)
    assert isinstance(tf_keras_output.get(), keras.Model)
    assert tf_keras_output.get().get_config() == model.get_config()

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
            tf_keras_output,
        ],
        engine=engine,
        use_local=True,
    )
