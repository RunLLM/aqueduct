import math

import pandas as pd
import pytest
from aqueduct.artifacts.bool_artifact import BoolArtifact
from aqueduct.artifacts.generic_artifact import GenericArtifact
from aqueduct.constants.enums import ArtifactType
from aqueduct.error import InvalidUserArgumentException
from data_objects import DataObject
from utils import extract, publish_flow_test, save

from aqueduct import check, global_config, metric, op


def test_lazy_sql_extractor(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT, lazy=True)
    assert table_artifact._get_content() is None
    assert isinstance(table_artifact.get(), pd.DataFrame)
    # After calling get(), artifact's content should be materialized.
    assert table_artifact._get_content() is not None


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


def test_table_artifact_lazy_syntax_sugar(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT, lazy=True)
    num_rows_artifact = table_artifact.number_of_rows(lazy=True)
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


def test_lazy_global_config(client, data_integration):
    with pytest.raises(InvalidUserArgumentException):
        global_config({"lazy": 1234})

    try:
        global_config({"lazy": True})

        # Basic SQL artifact that was lazily computed.
        table_artifact = extract(data_integration, DataObject.SENTIMENT)
        assert table_artifact._get_content() is None
        assert isinstance(table_artifact.get(), pd.DataFrame)
        # After calling get(), artifact's content should be materialized.
        assert table_artifact._get_content() is not None

        # For a lazily-created metric used pre-defined functions.
        table_artifact = extract(data_integration, DataObject.WINE)
        max_metric = table_artifact.max(column_id="fixed_acidity")
        assert max_metric._get_content() is None
        assert math.isclose(max_metric.get(), 15.899, rel_tol=1e-3)
        assert max_metric._get_content() is not None

        # For a workflow defined with the .lazy() decorator attribute.
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
        assert op_result.type() == ArtifactType.UNTYPED
        assert metric_result._get_content() is None
        assert check_result._get_content() is None

        assert op_result.get() == "hello"
        assert metric_result.get() == 2.0
        assert check_result.get()

        # After get(), everything should be materialized on the artifacts.
        assert op_result._get_content() is not None
        assert op_result.type() == ArtifactType.STRING
        assert metric_result._get_content() is not None
        assert check_result._get_content() is not None

    finally:
        global_config({"lazy": False})


def test_lazy_artifacts_backfilled_by_downstream(client):
    @op
    def generate_number():
        return 2.0

    @op
    def double_number(x):
        return 2 * x

    # Eager execution will type the upstream operator, but will not backfill the contents!
    num = generate_number.lazy()
    assert num.type() == ArtifactType.UNTYPED
    assert num._get_content() is None
    output = double_number(num)

    assert num.type() == ArtifactType.NUMERIC
    assert num._get_content() is None
    assert output.type() == ArtifactType.NUMERIC
    assert output._get_content() == 4.0

    # .get() will also type the upstream operator, same as above.
    num = generate_number.lazy()
    output = double_number.lazy(num)
    assert output.type() == ArtifactType.UNTYPED
    assert output._get_content() is None
    assert output.get() == 4.0

    assert num.type() == ArtifactType.NUMERIC
    assert num._get_content() is None
    assert output.type() == ArtifactType.NUMERIC
    assert output._get_content() == 4.0


def test_lazy_artifacts_with_custom_parameters(client):
    """Checks that we do not manifest the contents of a lazy artifact when custom parameters are provided."""
    num = client.create_param("number", default=10)

    @op
    def double_number(num):
        return 2 * num

    doubled = double_number.lazy(num)
    assert doubled.type() == ArtifactType.UNTYPED
    assert doubled._get_content() is None

    assert doubled.get(parameters={"number": 20}) == 40
    assert doubled.type() == ArtifactType.NUMERIC
    assert doubled._get_content() is None  # do not manifest the contents!


def test_lazy_artifact_with_save(client, flow_name, data_integration, engine, validator):
    reviews = extract(data_integration, DataObject.SENTIMENT)

    @op()
    def copy_field(df):
        df["new"] = df["review"]
        return df

    review_copied = copy_field.lazy(reviews)
    save(data_integration, review_copied)

    flow = publish_flow_test(
        client,
        review_copied,
        name=flow_name(),
        engine=engine,
    )
    validator.check_saved_artifact(
        flow, review_copied.id(), expected_data=copy_field.local(reviews)
    )
