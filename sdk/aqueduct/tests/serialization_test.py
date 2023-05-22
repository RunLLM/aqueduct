import json
import uuid

import cloudpickle as pickle
from aqueduct.constants.enums import (
    ArtifactType,
    ExecutionStatus,
    FunctionGranularity,
    FunctionType,
    GoogleSheetsSaveMode,
    LoadUpdateMode,
    OperatorType,
    S3TableFormat,
    SalesforceExtractType,
    SerializationType,
    ServiceType,
)
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG, Metadata
from aqueduct.models.execution_state import ExecutionState, Logs
from aqueduct.models.operators import (
    ExtractSpec,
    FunctionSpec,
    GoogleSheetsExtractParams,
    GoogleSheetsLoadParams,
    LoadSpec,
    Operator,
    OperatorSpec,
    RelationalDBExtractParams,
    RelationalDBLoadParams,
    S3ExtractParams,
    S3LoadParams,
    SalesforceExtractParams,
    SalesforceLoadParams,
)
from aqueduct.models.response_models import ArtifactResult, PreviewResponse
from aqueduct.tests.utils import _construct_dag, _construct_operator
from aqueduct.utils.serialization import (
    PickleableCollectionSerializationFormat,
    _read_image_content,
    _read_pickle_content,
    _read_string_content,
    artifact_type_to_serialization_type,
    deserialize,
    serialize_val,
)
from aqueduct.utils.utils import generate_uuid
from PIL import Image


def test_artifact_serialization():
    artifact_id = uuid.uuid4()
    artifact_name = "Extract Artifact"
    extract_artifact = ArtifactMetadata(
        id=artifact_id,
        name=artifact_name,
        type=ArtifactType.TABLE,
        explicitly_named=True,
    )

    assert extract_artifact.json() == json.dumps(
        {
            "id": str(artifact_id),
            "name": artifact_name,
            "type": ArtifactType.TABLE,
        }
    )


def test_operator_serialization():
    op_id = uuid.uuid4()
    other_ids = [uuid.uuid4(), uuid.uuid4()]
    fn_operator = Operator(
        id=op_id,
        name="Function Operator",
        description="",
        spec=OperatorSpec(
            function=FunctionSpec(
                type=FunctionType.FILE,
                granularity=FunctionGranularity.TABLE,
            )
        ),
        inputs=[other_ids[0]],
        outputs=[other_ids[1]],
        function_file_path="function.zip",
    )
    assert fn_operator.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Function Operator",
            "description": "",
            "spec": {
                "function": {
                    "type": FunctionType.FILE,
                    "granularity": FunctionGranularity.TABLE,
                    "language": "Python",
                },
            },
            "inputs": [str(other_ids[0])],
            "outputs": [str(other_ids[1])],
        }
    )


def test_preview_response_loading():
    op_id = uuid.uuid4()
    op_result = ExecutionState(
        status=ExecutionStatus.SUCCEEDED,
        user_logs=Logs(stdout="These are the operator logs"),
    )
    artifact_id = uuid.uuid4()
    artifact_result = ArtifactResult(
        serialization_type=SerializationType.TABLE,
        artifact_type=ArtifactType.TABLE,
        content="This is a serialized pandas dataframe",
    )
    preview_resp = {
        "status": ExecutionStatus.SUCCEEDED,
        "err_msg": "",
        "operator_results": {
            str(op_id): op_result,
        },
        "artifact_results": {
            str(artifact_id): artifact_result,
        },
    }

    assert PreviewResponse(**preview_resp) == PreviewResponse(
        status=ExecutionStatus.SUCCEEDED,
        operator_results={
            op_id: op_result,
        },
        artifact_results={
            artifact_id: artifact_result,
        },
    )


def test_excluded_fields_cannot_be_compared():
    op_id = generate_uuid()
    artifact_id = generate_uuid()

    op = _construct_operator(
        id=op_id,
        name="Extract",
        operator_type=OperatorType.EXTRACT,
        inputs=[],
        outputs=[artifact_id],
    )
    dag = _construct_dag(
        operators=[op],
        artifacts=[],
    )

    # It doesn't matter what we put as `operator_by_name`.
    assert dag == DAG(
        operators={**{str(op.id): op}},
        artifacts={},
        metadata=Metadata(),
    )
    assert dag == DAG(
        operators={**{str(op.id): op}},
        operator_by_name={**{op.name: op}},
        artifacts={},
        metadata=Metadata(),
    )


def test_extract_serialization():
    op_id = uuid.uuid4()
    other_ids = [uuid.uuid4(), uuid.uuid4()]
    resource_id = uuid.uuid4()
    extract_operator = Operator(
        id=op_id,
        name="Extract Operator",
        description="",
        spec=OperatorSpec(
            extract=ExtractSpec(
                service=ServiceType.POSTGRES,
                resource_id=resource_id,
                parameters=RelationalDBExtractParams(query="SELECT * FROM hotel_reviews;"),
            ),
        ),
        outputs=[other_ids[1]],
    )
    assert extract_operator.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Extract Operator",
            "description": "",
            "spec": {
                "extract": {
                    "service": ServiceType.POSTGRES,
                    "resource_id": str(resource_id),
                    "parameters": {
                        "query": "SELECT * FROM hotel_reviews;",
                    },
                }
            },
            "inputs": [],
            "outputs": [str(other_ids[1])],
        }
    )

    extract_operator_sf = Operator(
        id=op_id,
        name="Extract Operator Salesforce",
        description="",
        spec=OperatorSpec(
            extract=ExtractSpec(
                service=ServiceType.SALESFORCE,
                resource_id=resource_id,
                parameters=SalesforceExtractParams(
                    type=SalesforceExtractType.SEARCH, query="FIND joe;"
                ),
            ),
        ),
        outputs=[other_ids[1]],
    )
    assert extract_operator_sf.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Extract Operator Salesforce",
            "description": "",
            "spec": {
                "extract": {
                    "service": ServiceType.SALESFORCE,
                    "resource_id": str(resource_id),
                    "parameters": {
                        "type": "search",
                        "query": "FIND joe;",
                    },
                }
            },
            "inputs": [],
            "outputs": [str(other_ids[1])],
        }
    )

    extract_operator_gs = Operator(
        id=op_id,
        name="Extract Operator Google Sheets",
        description="",
        spec=OperatorSpec(
            extract=ExtractSpec(
                service=ServiceType.GOOGLE_SHEETS,
                resource_id=resource_id,
                parameters=GoogleSheetsExtractParams(spreadsheet_id="0"),
            ),
        ),
        outputs=[other_ids[1]],
    )
    assert extract_operator_gs.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Extract Operator Google Sheets",
            "description": "",
            "spec": {
                "extract": {
                    "service": ServiceType.GOOGLE_SHEETS,
                    "resource_id": str(resource_id),
                    "parameters": {
                        "spreadsheet_id": "0",
                    },
                }
            },
            "inputs": [],
            "outputs": [str(other_ids[1])],
        }
    )

    extract_operator_s3 = Operator(
        id=op_id,
        name="Extract Operator S3",
        description="",
        spec=OperatorSpec(
            extract=ExtractSpec(
                service=ServiceType.S3,
                resource_id=resource_id,
                parameters=S3ExtractParams(
                    filepath=json.dumps("test.csv"),
                    artifact_type=ArtifactType.TABLE,
                    format=S3TableFormat.CSV,
                ),
            ),
        ),
        outputs=[other_ids[1]],
    )
    assert extract_operator_s3.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Extract Operator S3",
            "description": "",
            "spec": {
                "extract": {
                    "service": ServiceType.S3,
                    "resource_id": str(resource_id),
                    "parameters": {
                        "filepath": json.dumps("test.csv"),
                        "artifact_type": ArtifactType.TABLE,
                        "format": "CSV",
                    },
                }
            },
            "inputs": [],
            "outputs": [str(other_ids[1])],
        }
    )


def test_load_serialization():
    op_id = uuid.uuid4()
    other_ids = [uuid.uuid4(), uuid.uuid4()]
    resource_id = uuid.uuid4()
    load_operator = Operator(
        id=op_id,
        name="Load Operator",
        description="",
        spec=OperatorSpec(
            load=LoadSpec(
                service=ServiceType.POSTGRES,
                resource_id=resource_id,
                parameters=RelationalDBLoadParams(
                    table="hotel_reviews", update_mode=LoadUpdateMode.REPLACE
                ),
            ),
        ),
        inputs=[other_ids[0]],
    )
    assert load_operator.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Load Operator",
            "description": "",
            "spec": {
                "load": {
                    "service": ServiceType.POSTGRES,
                    "resource_id": str(resource_id),
                    "parameters": {
                        "table": "hotel_reviews",
                        "update_mode": "replace",
                    },
                }
            },
            "inputs": [str(other_ids[0])],
            "outputs": [],
        }
    )

    load_operator_sf = Operator(
        id=op_id,
        name="Load Operator Salesforce",
        description="",
        spec=OperatorSpec(
            load=LoadSpec(
                service=ServiceType.SALESFORCE,
                resource_id=resource_id,
                parameters=SalesforceLoadParams(object="hotel_reviews"),
            ),
        ),
        inputs=[other_ids[0]],
    )
    assert load_operator_sf.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Load Operator Salesforce",
            "description": "",
            "spec": {
                "load": {
                    "service": ServiceType.SALESFORCE,
                    "resource_id": str(resource_id),
                    "parameters": {
                        "object": "hotel_reviews",
                    },
                }
            },
            "inputs": [str(other_ids[0])],
            "outputs": [],
        }
    )

    load_operator_gs = Operator(
        id=op_id,
        name="Load Operator Google Sheets",
        description="",
        spec=OperatorSpec(
            load=LoadSpec(
                service=ServiceType.GOOGLE_SHEETS,
                resource_id=resource_id,
                parameters=GoogleSheetsLoadParams(
                    filepath="test_sheet.csv",
                    save_mode=GoogleSheetsSaveMode.OVERWRITE,
                ),
            ),
        ),
        inputs=[other_ids[0]],
    )
    assert load_operator_gs.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Load Operator Google Sheets",
            "description": "",
            "spec": {
                "load": {
                    "service": ServiceType.GOOGLE_SHEETS,
                    "resource_id": str(resource_id),
                    "parameters": {
                        "filepath": "test_sheet.csv",
                        "save_mode": "overwrite",
                    },
                }
            },
            "inputs": [str(other_ids[0])],
            "outputs": [],
        }
    )

    load_operator_s3 = Operator(
        id=op_id,
        name="Load Operator S3",
        description="",
        spec=OperatorSpec(
            load=LoadSpec(
                service=ServiceType.S3,
                resource_id=resource_id,
                parameters=S3LoadParams(
                    filepath="test.json",
                    format=S3TableFormat.JSON,
                ),
            ),
        ),
        inputs=[other_ids[0]],
    )
    assert load_operator_s3.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Load Operator S3",
            "description": "",
            "spec": {
                "load": {
                    "service": ServiceType.S3,
                    "resource_id": str(resource_id),
                    "parameters": {
                        "filepath": "test.json",
                        "format": S3TableFormat.JSON,
                    },
                }
            },
            "inputs": [str(other_ids[0])],
            "outputs": [],
        }
    )

    load_operator_s3_without_format = Operator(
        id=op_id,
        name="Load Operator S3",
        description="",
        spec=OperatorSpec(
            load=LoadSpec(
                service=ServiceType.S3,
                resource_id=resource_id,
                parameters=S3LoadParams(
                    filepath="test.json",
                ),
            ),
        ),
        inputs=[other_ids[0]],
    )
    assert load_operator_s3_without_format.json(exclude_none=True) == json.dumps(
        {
            "id": str(op_id),
            "name": "Load Operator S3",
            "description": "",
            "spec": {
                "load": {
                    "service": ServiceType.S3,
                    "resource_id": str(resource_id),
                    "parameters": {
                        "filepath": "test.json",
                    },
                }
            },
            "inputs": [str(other_ids[0])],
            "outputs": [],
        }
    )


def test_serialization_of_pickled_collection_types():
    image_data = Image.open("aqueduct/tests/data/aqueduct.jpg", "r")
    list_input = [image_data, "hello world"]

    assert (
        artifact_type_to_serialization_type(
            ArtifactType.PICKLABLE, derived_from_bson=False, content=list_input
        )
        == SerializationType.PICKLE
    )

    serialized = serialize_val(
        list_input,
        SerializationType.PICKLE,
        False,
    )

    picklable_collection = PickleableCollectionSerializationFormat(
        **_read_pickle_content(serialized)
    )
    assert isinstance(picklable_collection, PickleableCollectionSerializationFormat)

    assert picklable_collection.aqueduct_serialization_types == [
        SerializationType.IMAGE,
        SerializationType.STRING,
    ]
    assert _read_image_content(picklable_collection.data[0]).getbbox() == list_input[0].getbbox()
    assert _read_string_content(picklable_collection.data[1]) == list_input[1]

    original_val = deserialize(
        serialization_type=SerializationType.PICKLE,
        artifact_type=ArtifactType.UNTYPED,  # irrelevant.
        content=serialized,
    )
    assert len(list_input) == len(original_val)
    assert list_input[0].getbbox() == original_val[0].getbbox()
    assert list_input[1] == original_val[1]
