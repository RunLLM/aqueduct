from typing import Dict, List

import pytest

from aqueduct.error import InvalidUserArgumentException
from constants import SENTIMENT_SQL_QUERY
from utils import get_integration_name, run_flow_test, generate_new_flow_name
from aqueduct import metric, op
import pandas as pd


@metric
def double_number_input(num: int) -> float:
    if not isinstance(num, int):
        raise Exception("Expected an integer input.")
    print("Result is ", 2 * num)
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
    assert kv_df.get().equals(pd.DataFrame(data=kv))


def test_non_jsonable_parameter(client):
    with pytest.raises(InvalidUserArgumentException):
        _ = client.create_param(name="bad param", default=b"cant serialize me")

    param = client.create_param(name="number", default=8)
    param_doubled = double_number_input(param)
    with pytest.raises(InvalidUserArgumentException):
        _ = param_doubled.get(parameters={"number": b"cant serialize me"})


def test_get_with_custom_parameter(client):
    param = client.create_param(name="number", default=8)
    assert param.get() == 8

    param_doubled = double_number_input(param)
    assert param_doubled.get(parameters={"number": 20}) == 40
    assert param_doubled.get() == 2 * 8

    with pytest.raises(InvalidUserArgumentException):
        param_doubled.get(parameters={"non-existant param": 10})


@op
def append_row_to_df(df, row):
    """`row` is a list of values to append to the input dataframe."""
    df.loc[len(df.index)] = row
    return df


def test_parameter_in_basic_flow(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    row_to_add = ["new hotel", "09-28-1996", "US", "It was new."]
    new_row_param = client.create_param(name="new row", default=row_to_add)
    output = append_row_to_df(sql_artifact, new_row_param)

    input_df = sql_artifact.get()
    input_df.loc[len(input_df.index)] = row_to_add

    output_df = output.get()
    assert output_df.equals(input_df)


@pytest.mark.publish
def test_edit_param_for_flow(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    row_to_add = ["new hotel", "09-28-1996", "US", "It was new."]
    new_row_param = client.create_param(name="new row", default=row_to_add)
    output = append_row_to_df(sql_artifact, new_row_param)

    flow_name = "Edit Parameter Test Flow"
    flow = run_flow_test(client, artifacts=[output], name=flow_name, delete_flow_after=False)
    flow_id = flow.id()

    try:
        # Edit the flow with a different row to append and re-publish
        row_to_add = ["another new hotel", "10-10-1000", "ID", "It was really really new."]
        new_row_param = client.create_param(name="new row", default=row_to_add)
        output = append_row_to_df(sql_artifact, new_row_param)

        # Wait for the first run, then refresh the workflow and verify that it runs at least
        # one more time (two runs total, since the original was manually triggered).
        flow = run_flow_test(
            client, artifacts=[output], name=flow_name, num_runs=2, delete_flow_after=False
        )
    finally:
        client.delete_flow(flow.id())

    assert flow_id == flow.id()


# TODO(ENG-1241): Once we are able to fetch flow results, Need to verify that the actual
@pytest.mark.publish
def test_trigger_flow_with_different_param(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)

    num_param = client.create_param(name="num", default=5)
    output = double_number_input(sql_artifact, num_param)

    flow_name = generate_new_flow_name()
    flow = run_flow_test(client, artifacts=[output], name=flow_name, delete_flow_after=False)

    try:
        client.trigger(flow.id(), parameters={"num": 10})
    finally:
        client.delete_flow(flow.id())
