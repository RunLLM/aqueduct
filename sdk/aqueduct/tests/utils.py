import json
import uuid
from typing import List

import pandas as pd
from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.dag import DAG, Metadata
from aqueduct.enums import (
    ArtifactType,
    CheckSeverity,
    ExecutionStatus,
    FunctionGranularity,
    FunctionType,
    LoadUpdateMode,
    OperatorType,
    SerializationType,
    ServiceType,
)
from aqueduct.operators import (
    CheckSpec,
    ExtractSpec,
    FunctionSpec,
    LoadSpec,
    MetricSpec,
    Operator,
    OperatorSpec,
    RelationalDBExtractParams,
    RelationalDBLoadParams,
)
from aqueduct.responses import ArtifactResult, PreviewResponse
from aqueduct.utils import generate_uuid

from aqueduct import dag as dag_module


def generate_uuids(num: int) -> List[uuid.UUID]:
    return [generate_uuid() for _ in range(num)]


def _construct_dag(
    operators: List[Operator],
    artifacts: List[ArtifactMetadata],
):
    return DAG(
        operators={**{str(op.id): op for op in operators}},
        operator_by_name={**{op.name: op for op in operators}},
        artifacts={**{str(artifact.id): artifact for artifact in artifacts}},
        metadata=Metadata(),
    )


def _construct_operator(
    id: uuid.UUID,
    name: str,
    operator_type: OperatorType,
    inputs: List[uuid.UUID],
    outputs: List[uuid.UUID],
):
    """Only sets the fields needed to figure out the DAG structure, not to actually execute the operator."""
    if operator_type == OperatorType.EXTRACT:
        spec = default_extract_spec()
    elif operator_type == OperatorType.FUNCTION:
        spec = default_function_spec()
    elif operator_type == OperatorType.CHECK:
        spec = default_check_spec()
    elif operator_type == OperatorType.METRIC:
        spec = default_metric_spec()
    else:
        spec = default_load_spec()

    return Operator(
        id=id,
        name=name,
        description="",
        spec=spec,
        inputs=inputs,
        outputs=outputs,
    )


def default_extract_spec() -> OperatorSpec:
    return OperatorSpec(
        extract=ExtractSpec(
            service=ServiceType.POSTGRES,
            integration_id=generate_uuid(),
            parameters=RelationalDBExtractParams(query="This is a SQL Query"),
        )
    )


def default_function_spec() -> OperatorSpec:
    return OperatorSpec(
        function=FunctionSpec(
            type=FunctionType.FILE,
            granularity=FunctionGranularity.TABLE,
        )
    )


def default_check_spec() -> OperatorSpec:
    return OperatorSpec(
        check=CheckSpec(
            level=CheckSeverity.WARNING,
            function=FunctionSpec(
                type=FunctionType.FILE,
                granularity=FunctionGranularity.TABLE,
            ),
        )
    )


def default_metric_spec() -> OperatorSpec:
    return OperatorSpec(
        metric=MetricSpec(
            function=FunctionSpec(
                type=FunctionType.FILE,
                granularity=FunctionGranularity.TABLE,
            ),
        )
    )


def default_load_spec() -> OperatorSpec:
    return OperatorSpec(
        load=LoadSpec(
            service=ServiceType.POSTGRES,
            integration_id=generate_uuid(),
            parameters=RelationalDBLoadParams(table="output", update_mode=LoadUpdateMode.REPLACE),
        )
    )


def default_artifact(id: uuid.UUID, name: str) -> ArtifactMetadata:
    return ArtifactMetadata(id=id, name=name, type=ArtifactType.TABLE)


def default_table_artifact(
    operator_name="extract_operator",
    operator_id=None,
    artifact_name="table_artifact",
    artifact_id=None,
) -> TableArtifact:
    if not operator_id:
        operator_id = generate_uuid()
    if not artifact_id:
        artifact_id = generate_uuid()
    artifact = ArtifactMetadata(id=artifact_id, name=artifact_name, type=ArtifactType.TABLE)
    op = _construct_operator(
        id=operator_id,
        name=operator_name,
        operator_type=OperatorType.EXTRACT,
        inputs=[],
        outputs=[artifact_id],
    )
    dag_module.__GLOBAL_DAG__ = _construct_dag(
        operators=[op],
        artifacts=[artifact],
    )
    return TableArtifact(
        dag=dag_module.__GLOBAL_DAG__, artifact_id=artifact_id, content=pd.DataFrame()
    )


# This helper function is used to mock our preview call so that it 1) captures the randomly generated
# output artifact id, and 2) returns the mocked preview response result keyed by that artifact id.
def construct_mocked_preview(
    artifact_name: str,
    artifact_type: ArtifactType,
    serialization_type: SerializationType,
    content: any,
):
    def mocked_preview(dag: DAG):
        output_artifact_id = None
        for id in dag.artifacts:
            if dag.artifacts[id].name == artifact_name:
                output_artifact_id = id
                break

        if output_artifact_id is None:
            raise Exception("Unable to find output artifact from the dag.")

        status = ExecutionStatus.SUCCEEDED

        if serialization_type == SerializationType.TABLE:
            serialized_content = content.to_json(
                orient="table", date_format="iso", index=False
            ).encode()
        elif serialization_type == SerializationType.JSON:
            serialized_content = json.dumps(content).encode()
        else:
            raise Exception("Unexpected serialization type %s." % serialization_type)

        artifact_results = {
            output_artifact_id: ArtifactResult(
                serialization_type=serialization_type,
                artifact_type=artifact_type,
                content=serialized_content,
            ),
        }

        return PreviewResponse(
            status=status,
            operator_results={},
            artifact_results=artifact_results,
        )

    return mocked_preview
