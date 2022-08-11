import base64
import json
import sys
from io import StringIO
from unittest.mock import MagicMock

from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.decorator import check, metric
from aqueduct.enums import ArtifactType, ExecutionStatus, SerializationType
from aqueduct.responses import ArtifactResult, PreviewResponse
from aqueduct.tests.utils import default_table_artifact
from aqueduct.utils import delete_zip_folder_and_file, generate_uuid

from aqueduct import api_client

metric_op_name = "metric_fn"
description = f"{metric_op_name} description"
metric_artifact_name = f"{metric_op_name} artifact"
zip_folder = f"{metric_op_name}_aqueduct_metric"


@metric(description=description)
def metric_fn(df):
    pass


def test_metric():
    output = 10
    op_name = "op"
    op_id = generate_uuid()
    artifact_id = generate_uuid()
    artifact_name = "artifact"

    metric_input = default_table_artifact(
        operator_name=op_name,
        operator_id=op_id,
        artifact_name=artifact_name,
        artifact_id=artifact_id,
    )
    dag = metric_input._dag

    def mocked_preview(dag):
        output_artifact_id = None
        for id in dag.artifacts:
            if dag.artifacts[id].name == metric_artifact_name:
                output_artifact_id = id
                break

        if output_artifact_id is None:
            raise Exception("Unable to find output artifact from the dag.")

        status = ExecutionStatus.SUCCEEDED
        artifact_results = {
            output_artifact_id: ArtifactResult(
                serialization_type=SerializationType.JSON,
                artifact_type=ArtifactType.NUMERIC,
                content=base64.b64encode(json.dumps(output).encode()),
            ),
        }

        return PreviewResponse(
            status=status,
            operator_results={},
            artifact_results=artifact_results,
        )

    api_client.__GLOBAL_API_CLIENT__.preview = MagicMock(side_effect=mocked_preview)

    try:
        metric_output: NumericArtifact = metric_fn(metric_input)
    finally:
        delete_zip_folder_and_file(zip_folder)

    metric_val = metric_output.get()

    assert len(dag.operators) == len(dag.artifacts)
    assert len(dag.operators) == 2

    artifact_check = {
        artifact_name: ArtifactType.TABULAR,
        metric_artifact_name: ArtifactType.NUMERIC,
    }

    for artifact in dag.artifacts:
        artifact = dag.artifacts[artifact]
        assert artifact.name in artifact_check.keys()
        assert artifact.type == artifact_check[artifact.name]
        if artifact.name == metric_artifact_name:
            metric_artifact_id = artifact.id

    operator_check = {
        op_name: {
            "inputs": [],
            "outputs": [artifact_id],
            "description": "",
        },
        metric_op_name: {
            "inputs": [artifact_id],
            "outputs": [metric_artifact_id],
            "description": description,
        },
    }

    for operator in dag.operators:
        operator = dag.operators[operator]
        assert operator.name in operator_check.keys()
        assert operator.description == operator_check[operator.name]["description"]
        for artifacts, key in [
            (operator.inputs, "inputs"),
            (operator.outputs, "outputs"),
        ]:
            assert len(artifacts) == len(operator_check[operator.name][key])
            for artifact in operator_check[operator.name][key]:
                assert artifact in artifacts

    assert metric_val == output


def test_metrics_and_checks_on_table_describe():
    metric_input = default_table_artifact()

    try:
        metric_output: NumericArtifact = metric_fn(metric_input)
    finally:
        delete_zip_folder_and_file(zip_folder)

    check_description = "Check description"

    @check(description=check_description)
    def check_fn(metric_output):
        return metric_output > 0

    check_name = "check_fn"
    check_artifact_name = f"{check_name} artifact"

    def mocked_preview(dag):
        output_artifact_id = None
        for id in dag.artifacts:
            if dag.artifacts[id].name == check_artifact_name:
                output_artifact_id = id
                break

        if output_artifact_id is None:
            raise Exception("Unable to find output artifact from the dag.")

        status = ExecutionStatus.SUCCEEDED
        artifact_results = {
            output_artifact_id: ArtifactResult(
                serialization_type=SerializationType.JSON,
                artifact_type=ArtifactType.BOOL,
                content=base64.b64encode(json.dumps(True).encode()),
            ),
        }

        return PreviewResponse(
            status=status,
            operator_results={},
            artifact_results=artifact_results,
        )

    api_client.__GLOBAL_API_CLIENT__.preview = MagicMock(side_effect=mocked_preview)

    check_fn(metric_output)

    redirect_stdout = StringIO()
    stdout = sys.stdout
    sys.stdout = redirect_stdout
    metric_input.describe()
    describe_table = redirect_stdout.getvalue()
    sys.stdout = stdout

    output_dict = json.loads("\n".join(describe_table.split("\n")[1:]))

    assert len(output_dict["Metrics"]) == 1
    metric_descr = output_dict["Metrics"][0]
    assert metric_descr["Label"] == metric_op_name
    assert metric_descr["Description"] == description
    assert len(metric_descr["Checks"]) == 1
    metric_check = metric_descr["Checks"][0]
    assert metric_check["Label"] == check_name
    assert metric_check["Description"] == check_description
    assert "Level" in metric_check
