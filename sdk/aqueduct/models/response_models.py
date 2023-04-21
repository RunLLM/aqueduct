import uuid
from datetime import datetime
from typing import Dict, List, Optional

from aqueduct.constants.enums import (
    ArtifactType,
    ExecutionStatus,
    K8sClusterStatusType,
    SerializationType,
)
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import Metadata, RetentionPolicy, Schedule
from aqueduct.models.execution_state import ExecutionState
from aqueduct.models.operators import LoadSpec, Operator, OperatorSpec
from aqueduct.models.utils import human_readable_timestamp
from pydantic import BaseModel


class ArtifactResult(BaseModel):
    serialization_type: SerializationType
    artifact_type: ArtifactType
    content: bytes


class PreviewResponse(BaseModel):
    """This is the response object returned by api_client.preview().

    Attributes:
        status:
            The execution state of preview.
        operator_results:
            A map from an operator id to its OperatorResult object.
            All operators that were run will appear in this map.

        artifact_results:
            A map from an artifact id to its base64 encoded string.
            Artifact results will only appear in this map if explicitly
            specified in the `target_ids` on the request.
    """

    status: ExecutionStatus
    operator_results: Dict[uuid.UUID, ExecutionState]
    artifact_results: Dict[uuid.UUID, ArtifactResult]


class RegisterWorkflowResponse(BaseModel):
    """This is the response object returned by api_client.register_workflow().

    Attributes:
        id:
            The uuid if of the newly registered workflow.
        python_version:
            The Python version in the engine in string format "Python {major_version}.{minor_version}.{patch_level}" e.g. "Python 3.9.11".
    """

    id: uuid.UUID
    python_version: str


class RegisterAirflowWorkflowResponse(BaseModel):
    """This is the response object returned by api_client.register_airflow_workflow().

    Attributes:
        id:
            The uuid if of the newly registered workflow.
    """

    id: uuid.UUID
    # TODO ENG-1481: Return an actual file instead of a string.
    file: str
    is_update: bool


class ListWorkflowResponseEntry(BaseModel):
    """A list of these response objects is returned by api_client.list_workflows()
    and corresponds with a single workflow.

    Attributes:
        id, name, description:
            The id, name, and description of the workflow.
        created_at:
            The unix timestamp in seconds that the workflow was first created at.
        last_run_at:
            The unit timestamp in seconds that the last workflow run was started.
        status:
            The execution status of the latest run of this workflow.
    """

    id: uuid.UUID
    name: str
    description: str
    created_at: int
    last_run_at: int
    status: ExecutionStatus
    engine: str

    def to_readable_dict(self) -> Dict[str, str]:
        return {
            "flow_id": str(self.id),
            "name": self.name,
            "description": self.description,
            "created_at": human_readable_timestamp(self.created_at),
            "last_run_at": human_readable_timestamp(self.last_run_at),
            "last_run_status": str(self.status),
            "engine": self.engine,
        }


class WorkflowDagResponse(BaseModel):
    """Represents a dag structure that could have been executed multiple times.

    It is an essentially a "version" of a flow.

    Attributes:
        id:
            The id of the workflow dag. This is not useful to the user.
        workflow_id:
            The id of the workflow that this dag belongs to.
        metadata:
            This workflow version's metadata like description, schedule, etc.
        operators, artifacts:
            Describes this workflow version's dag structure.

    Excluded Attributes:
        created_at, storage_config
    """

    id: uuid.UUID
    workflow_id: uuid.UUID
    metadata: Metadata
    operators: Dict[str, Operator]
    artifacts: Dict[str, ArtifactMetadata]


class WorkflowDagResultResponse(BaseModel):
    """Represents the result of a single workflow run.

    NOTE: Very confusingly, this is not the response from `get_workflow_dag_result.go`, but instead is
    derived from the `get_workflow.go` response.

    Attributes:
        id:
            The id of the workflow run. This is the same id users can use to fetch
            FlowRuns.
        created_at:
            The unix timestamp in seconds that this workflow run was started at.
        status:
            The execution status of this workflow run.
        workflow_dag_id:
            This id can be used to find the corresponding workflow dag version.

    """

    id: uuid.UUID
    created_at: int

    # TODO(ENG-2665): remove the status field.
    status: ExecutionStatus
    exec_state: ExecutionState
    workflow_dag_id: uuid.UUID

    def to_readable_dict(self) -> Dict[str, str]:
        readable = {
            "run_id": str(self.id),
            "created_at": human_readable_timestamp(self.created_at),
            "status": self.status.value,
        }
        if self.exec_state.error is not None:
            readable["error"] = self.exec_state.error.tip + "\n" + self.exec_state.error.context
        return readable


class GetWorkflowResponse(BaseModel):
    """This is the response object returned by api_client.get_workflow().

    Attributes:
        workflow_dags:
            All the historical workflow dags.
        workflow_dag_results:
            All the historical workflow results.

    Excluded Attributes:
        watcher_auth_ids
    """

    workflow_dags: Dict[uuid.UUID, WorkflowDagResponse]
    workflow_dag_results: List[WorkflowDagResultResponse]


class WorkflowDagResultResultResponse(BaseModel):
    """A workflow dag result's execution state, derived from api_client.get_workflow_dag_result()."""

    id: uuid.UUID
    exec_state: ExecutionState


class OperatorRawResultResponse(BaseModel):
    """An operator result's execution state, derived from api_client.get_workflow_dag_result()."""

    id: uuid.UUID
    exec_state: ExecutionState


class ArtifactRawResultResponse(BaseModel):
    """An artifact result's execution state, derived from api_client.get_workflow_dag_result()."""

    id: uuid.UUID
    exec_state: ExecutionState


class OperatorResultResponse(BaseModel):
    """An operator result, derived from the returned response from api_client.get_workflow_dag_result()."""

    # Copied from the Operator class, because the golang response nests that class
    # in a way that is hard to represent in pydantic.
    id: uuid.UUID
    name: str
    description: str
    spec: OperatorSpec
    inputs: List[uuid.UUID] = []
    outputs: List[uuid.UUID] = []

    # The operator execution result.
    result: Optional[OperatorRawResultResponse]

    def to_operator(self) -> Operator:
        """Convert to an operator class."""
        return Operator(**self.dict())


class ArtifactResultResponse(BaseModel):
    """An artifact result, derived from the returned response from api_client.get_workflow_dag_result().

    Excluded attributes:
    - from: the upstream operator ID. Also, "from" is a reserved keyword in pydantic.
    - to: the downstream operator IDs.
    """

    # Copied from the ArtifactMetadata class, because the golang response nests that class
    # in a way that is hard to represent in pydantic.
    id: uuid.UUID
    name: str
    type: ArtifactType

    def to_artifact(self) -> ArtifactMetadata:
        """Convert to an artifact class."""
        return ArtifactMetadata(**self.dict())


class GetWorkflowDagResultResponse(BaseModel):
    """This is the response object returned by api_client.get_workflow_dag_result().

    Excluded attributes:
        Only the necessary metadata fields are included.
    """

    # This is a subset of what metadata is available on the backend.
    # Allows for this object to be cast into a Metadata object.
    name: Optional[str]
    description: Optional[str]
    schedule: Optional[Schedule]
    retention_policy: Optional[RetentionPolicy]
    dag_created_at: datetime

    result: WorkflowDagResultResultResponse
    operators: Dict[uuid.UUID, OperatorResultResponse]
    artifacts: Dict[uuid.UUID, ArtifactResultResponse]

    def metadata(self) -> Metadata:
        return Metadata(**self.dict())


class SavedObjectDelete(BaseModel):
    """This is an item in the list returned by DeleteWorkflowResponse."""

    name: str
    exec_state: ExecutionState


class DeleteWorkflowResponse(BaseModel):
    """This is the response object returned by api_client.delete_workflow().

    Attributes:
        saved_object_deletion_results:
            Results of deleting saved objects.
            Key: Integration name
            Value: List of SavedObjectDelete belonging to that integration
    """

    saved_object_deletion_results: Dict[str, List[SavedObjectDelete]]


class SavedObjectUpdate(BaseModel):
    """This is an item in the list returned by ListWorkflowSavedObjectsResponse."""

    operator_name: str
    modified_at: str
    integration_name: str
    spec: LoadSpec


class ListWorkflowSavedObjectsResponse(BaseModel):
    """This is the response object returned by api_client.get_workflow_writes().

    Attributes:
        table_details:
            List of objects written by the workflow.
    """

    object_details: Optional[List[SavedObjectUpdate]]


class GetVersionResponse(BaseModel):
    """This is the response object returned by /api/version."""

    version: str


class DynamicEngineStatusResponse(BaseModel):
    id: uuid.UUID
    name: str
    status: K8sClusterStatusType
