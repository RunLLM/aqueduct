import base64
import json
from unittest.mock import MagicMock

import pandas as pd
from aqueduct.artifacts.bool_artifact import BoolArtifact
from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.decorator import check, metric, op
from aqueduct.enums import ArtifactType, ExecutionStatus, SerializationType
from aqueduct.responses import ArtifactResult, PreviewResponse
from aqueduct.tests.utils import default_table_artifact
from aqueduct.utils import delete_zip_folder_and_file

from aqueduct import api_client


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
            output_artifact_type: ArtifactType.TABULAR,
            output_serialization_type: SerializationType.TABULAR,
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

            def mocked_preview(dag):
                output_artifact_id = None
                for id in dag.artifacts:
                    if dag.artifacts[id].name == artifact_name:
                        output_artifact_id = id
                        break

                if output_artifact_id is None:
                    raise Exception("Unable to find output artifact from the dag.")

                status = ExecutionStatus.SUCCEEDED
                if expected_artifact_type == ArtifactType.TABULAR:
                    serialized_data = expected_content.to_json(
                        orient="table", date_format="iso", index=False
                    ).encode()
                else:
                    serialized_data = json.dumps(expected_content).encode()
                artifact_results = {
                    output_artifact_id: ArtifactResult(
                        serialization_type=expected_serialization_type,
                        artifact_type=expected_artifact_type,
                        content=base64.b64encode(serialized_data),
                    ),
                }

                return PreviewResponse(
                    status=status,
                    operator_results={},
                    artifact_results=artifact_results,
                )

            api_client.__GLOBAL_API_CLIENT__.preview = MagicMock(side_effect=mocked_preview)

            try:
                fn_output = fn(inp)
            finally:
                delete_zip_folder_and_file(decorator_data[label](name))
            assert isinstance(
                fn_output, expected_artifact_class
            ), f"Expected: {expected_artifact_class}, Got: {type(fn_output)} for decorator {decorator} {inp_type}"
