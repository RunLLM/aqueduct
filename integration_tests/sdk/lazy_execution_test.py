import pandas as pd
import pytest
from aqueduct.artifacts.bool_artifact import BoolArtifact
from aqueduct.artifacts.generic_artifact import GenericArtifact
from aqueduct.error import InvalidUserArgumentException
from constants import SENTIMENT_SQL_QUERY
from utils import get_integration_name

from aqueduct import check, global_config, metric, op


def test_lazy_sql_extractor(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY, lazy=True)
    assert sql_artifact._get_content() is None
    assert isinstance(sql_artifact.get(), pd.DataFrame)
    # After calling get(), artifact's content should be materialized.
    assert sql_artifact._get_content() is not None


def test_lazy_operators(client):
    @op
    def dummy_op():
        return "hello"

    @metric
    def dummy_metric(arg):
        return 2.0

    @check
    def dummy_check(arg):
        return True

    op_result = dummy_op.lazy()
    assert op_result._get_content() is None
    assert op_result.get() == "hello"
    # After calling get(), artifact's content should be materialized.
    assert op_result._get_content() == "hello"

    op_result = dummy_op()
    metric_result = dummy_metric.lazy(op_result)
    check_result = dummy_check.lazy(op_result)

    assert metric_result._get_content() is None
    assert check_result._get_content() is None

    assert metric_result.get() == 2.0
    assert check_result.get() == True

    # After calling get(), artifact's content should be materialized.
    assert metric_result._get_content() == 2.0
    assert check_result._get_content() == True


def test_eager_operator_after_lazy(client):
    @op
    def foo():
        return "hello"

    @op
    def bar(arg):
        return 2.0

    foo_result = foo.lazy()
    bar_result = bar(foo_result)

    # bar should be executed eagerly despite that foo is in lazy mode.
    assert bar_result._get_content() == 2.0
    assert bar_result.get() == 2.0


def test_table_artifact_lazy_syntax_sugar(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY, lazy=True)
    num_rows_artifact = sql_artifact.number_of_rows(lazy=True)
    assert num_rows_artifact._get_content() is None
    assert isinstance(num_rows_artifact.get(), float)
    # After calling get(), artifact's content should be materialized.
    assert num_rows_artifact._get_content() is not None


def test_numeric_artifact_lazy_syntax_sugar(client):
    @op
    def generate_number():
        return 2.0

    num_artifact = generate_number()
    bool_artifact = num_artifact.bound(upper=3.0, lazy=True)
    assert isinstance(bool_artifact, BoolArtifact)
    assert bool_artifact._get_content() is None

    assert bool_artifact.get()
    assert bool_artifact._get_content()


def test_lazy_artifact_type(client):
    @op
    def generate_number():
        return 2.0

    output_artifact = generate_number.lazy()
    assert isinstance(output_artifact, GenericArtifact)

    assert output_artifact.get() == 2.0
    # For lazily generated artifact, even after we materialize its value, its type should still be
    # `GenericArtifact` and does not expose type-specific syntax sugars.
    assert isinstance(output_artifact, GenericArtifact)


def test_lazy_global_config(client):
    with pytest.raises(InvalidUserArgumentException):
        global_config({"lazy": 1234})

    try:
        global_config({"lazy": True})

        db = client.integration(name=get_integration_name())
        sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
        assert sql_artifact._get_content() is None
        assert isinstance(sql_artifact.get(), pd.DataFrame)
        # After calling get(), artifact's content should be materialized.
        assert sql_artifact._get_content() is not None

        @op
        def dummy_op():
            return "hello"

        @metric
        def dummy_metric(arg):
            return 2.0

        @check
        def dummy_check(arg):
            return True

        op_result = dummy_op()
        metric_result = dummy_metric(op_result)
        check_result = dummy_check(op_result)

        assert op_result._get_content() is None
        assert metric_result._get_content() is None
        assert check_result._get_content() is None

        assert op_result.get() == "hello"
        assert metric_result.get() == 2.0
        assert check_result.get()

        # After get(), everything should be materialized on the artifacts.
        assert op_result._get_content() is not None
        assert metric_result._get_content() is not None
        assert check_result._get_content() is not None

    finally:
        global_config({"lazy": False})
