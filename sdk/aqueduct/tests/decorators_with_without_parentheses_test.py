from unittest.mock import MagicMock

import pandas as pd
from aqueduct.artifacts.bool_artifact import BoolArtifact
from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.decorator import check, metric, op
from aqueduct.enums import ArtifactType, SerializationType
from aqueduct.tests.utils import construct_mocked_preview, default_table_artifact
from aqueduct.utils import delete_zip_folder_and_file

from aqueduct import globals


def test_decorators_with_without_parentheses():
    inp = default_table_artifact()

    @op()
    def op_fn_with_parentheses(df):
        pass

    @op
    def op_fn_without_parentheses(df):
        pass

    @metric()
    def metric_fn_with_parentheses(df):
        pass

    @metric
    def metric_fn_without_parentheses(df):
        pass

    @check()
    def check_fn_with_parentheses(df):
        pass

    @check
    def check_fn_without_parentheses(df):
        pass

    w_parentheses = "with parentheses"
    wo_parentheses = "without parentheses"
    output_artifact_type = "output_artifact_type"
    output_serialization_type = "output_serialization_type"
    content = "content"
    artifact_class = "artifact_class"
    label = "label"

    decorators = {
        "op": {
            w_parentheses: op_fn_with_parentheses,
            wo_parentheses: op_fn_without_parentheses,
            output_artifact_type: ArtifactType.TABLE,
            output_serialization_type: SerializationType.TABLE,
            content: pd.DataFrame(),
            artifact_class: TableArtifact,
            label: lambda name: f"{name}_aqueduct",
        },
        "metric": {
            w_parentheses: metric_fn_with_parentheses,
            wo_parentheses: metric_fn_without_parentheses,
            output_artifact_type: ArtifactType.NUMERIC,
            output_serialization_type: SerializationType.JSON,
            content: 1.0,
            artifact_class: NumericArtifact,
            label: lambda name: f"{name}_aqueduct_metric",
        },
        "check": {
            w_parentheses: check_fn_with_parentheses,
            wo_parentheses: check_fn_without_parentheses,
            output_artifact_type: ArtifactType.BOOL,
            output_serialization_type: SerializationType.JSON,
            content: True,
            artifact_class: BoolArtifact,
            label: lambda name: f"{name}_aqueduct_check",
        },
    }

    for decorator in decorators.keys():
        decorator_data = decorators[decorator]
        expected_artifact_type = decorator_data[output_artifact_type]
        expected_serialization_type = decorator_data[output_serialization_type]
        expected_content = decorator_data[content]
        expected_artifact_class = decorator_data[artifact_class]
        for inp_type in [w_parentheses, wo_parentheses]:
            fn = decorator_data[inp_type]
            name = f"{decorator}_fn_{inp_type.replace(' ', '_')}"
            artifact_name = f"{name} artifact"

            globals.__GLOBAL_API_CLIENT__.preview = MagicMock(
                side_effect=construct_mocked_preview(
                    artifact_name,
                    expected_artifact_type,
                    expected_serialization_type,
                    expected_content,  # dummy data
                )
            )

            try:
                fn_output = fn(inp)
            finally:
                delete_zip_folder_and_file(decorator_data[label](name))
            assert isinstance(
                fn_output, expected_artifact_class
            ), f"Expected: {expected_artifact_class}, Got: {type(fn_output)} for decorator {decorator} {inp_type}"
