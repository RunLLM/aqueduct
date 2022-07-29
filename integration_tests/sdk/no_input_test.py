import pandas as pd

from aqueduct import op


@op
def no_input() -> pd.DataFrame:
    d = {'col1': [1, 2], 'col2': [3, 4]}
    return pd.DataFrame(data=d)


@op
def join(x: pd.DataFrame, y: pd.DataFrame) -> pd.DataFrame:
    return x


def test_basic_no_input_function(client):
    expected = pd.DataFrame(data={'col1': [1, 2], 'col2': [3, 4]})
    result = no_input().get()
    assert result.equals(expected)


def test_flow_with_no_input_function(client):
    warehouse = client.integration(name="aqueduct_demo")
    customers_table = warehouse.sql(query="SELECT * FROM customers;")

    result = join(no_input(), customers_table)
    expected = pd.DataFrame(data={'col1': [1, 2], 'col2': [3, 4]})
    assert result.get().equals(expected)
