import uuid
from typing import Dict, List, Optional

from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.config import EngineConfig
from aqueduct.enums import ArtifactType, OperatorType, TriggerType
from aqueduct.error import ArtifactNotFoundException, InternalAqueductError
from aqueduct.operators import (
    Operator,
    OperatorSpec,
    get_operator_type,
    get_operator_type_from_spec,
)
from pydantic import BaseModel


class Schedule(BaseModel):
    trigger: Optional[TriggerType] = None
    cron_schedule: str = ""
    disable_manual_trigger: bool = False


class RetentionPolicy(BaseModel):
    k_latest_runs: int = -1


class Metadata(BaseModel):
    """These fields should always set when writing/reading from the backend."""

    name: Optional[str]
    description: Optional[str]
    schedule: Optional[Schedule]
    retention_policy: Optional[RetentionPolicy]


class DAG(BaseModel):
    operators: Dict[str, Operator] = {}
    artifacts: Dict[str, ArtifactMetadata] = {}

    # Allows for quick operator lookup by name.
    # Is excluded from json serialization.
    operator_by_name: Dict[str, Operator] = {}

    # These fields must be set when publishing the workflow
    metadata: Metadata
    engine_config: EngineConfig = EngineConfig()

    class Config:
        fields = {
            "operators_by_name": {"exclude": ...},
        }

    def must_get_operator(
        self,
        with_id: Optional[uuid.UUID] = None,
        with_name: Optional[str] = None,
        with_output_artifact_id: Optional[uuid.UUID] = None,
    ) -> Operator:
        op = self.get_operator(with_id, with_name, with_output_artifact_id)
        if op is None:
            raise InternalAqueductError(
                "Unable to find operator: with_id %s, with_name %s, with_output_artifact_id %s"
                % (str(with_id), with_name, str(with_output_artifact_id)),
            )
        return op

    def get_operator(
        self,
        with_id: Optional[uuid.UUID] = None,
        with_name: Optional[str] = None,
        with_output_artifact_id: Optional[uuid.UUID] = None,
    ) -> Optional[Operator]:
        if (
            int(with_id is not None)
            + int(with_name is not None)
            + int(with_output_artifact_id is not None)
        ) != 1:
            raise InternalAqueductError(
                "Cannot fetch operator with multiple search parameters set."
            )

        if with_id is not None:
            return self.operators.get(str(with_id))

        elif with_name is not None:
            return self.operator_by_name.get(with_name)

        # Search with output artifact id
        for _, op in self.operators.items():
            if with_output_artifact_id in op.outputs:
                return op
        return None

    def list_operators(
        self,
        filter_to: Optional[List[OperatorType]] = None,
        on_artifact_id: Optional[uuid.UUID] = None,
    ) -> List[Operator]:
        """Multiple conditions can be applied to filter down the list of operators."""
        operators = list(self.operators.values())

        if filter_to is not None:
            operators = [op for op in operators if get_operator_type(op) in filter_to]

        if on_artifact_id is not None:
            operators = [op for op in operators if on_artifact_id in op.inputs]
        return operators

    def list_downstream_operators(
        self,
        op_id: uuid.UUID,
    ) -> List[uuid.UUID]:
        """Returns a list of all operators that depend on the given operator. Includes the given operator."""
        downstream_ops = []

        q = [op_id]
        seen_op_ids = set(q)
        while len(q) > 0:
            curr_op_id = q.pop(0)
            downstream_ops.append(curr_op_id)

            curr_op = self.must_get_operator(with_id=curr_op_id)
            for output_artifact_id in curr_op.outputs:
                next_op_ids = [
                    op.id
                    for op in self.list_operators(on_artifact_id=output_artifact_id)
                    if op.id not in seen_op_ids
                ]
                seen_op_ids.union(set(next_op_ids))
                q.extend(next_op_ids)

        return downstream_ops

    def list_root_operators(
        self, for_artifact_ids: Optional[List[uuid.UUID]] = None
    ) -> List[Operator]:
        all_root_operators = [op for op in self.operators.values() if len(op.inputs) == 0]
        if for_artifact_ids is None:
            return all_root_operators

        # Perform a DFS in the reverse direction to find all upstream operators, starting at the given artifacts.
        root_operators = []
        q: List[Operator] = [
            self.must_get_operator(with_output_artifact_id=artifact_id)
            for artifact_id in for_artifact_ids
        ]
        seen_op_ids = set(op.id for op in q)
        while len(q):
            curr_op = q.pop(0)
            if get_operator_type(curr_op) == OperatorType.EXTRACT:
                root_operators.append(curr_op)
                continue

            input_operators = [
                self.must_get_operator(with_output_artifact_id=input_artifact_id)
                for input_artifact_id in curr_op.inputs
            ]
            previous_operators = [op for op in input_operators if op.id not in seen_op_ids]
            q.extend(previous_operators)
            seen_op_ids.union(set(op.id for op in previous_operators))

        return root_operators

    def must_get_artifact(self, artifact_id: uuid.UUID) -> ArtifactMetadata:
        if str(artifact_id) not in self.artifacts:
            raise ArtifactNotFoundException("Unable to find artifact.")
        return self.artifacts[str(artifact_id)]

    def must_get_artifacts(self, artifact_ids: List[uuid.UUID]) -> List[ArtifactMetadata]:
        return [self.must_get_artifact(artifact_id) for artifact_id in artifact_ids]

    def get_artifact_by_name(self, name: str) -> Optional[ArtifactMetadata]:
        for artifact in self.list_artifacts():
            if artifact.name == name:
                return artifact

        return None

    def list_artifacts(
        self,
        on_op_ids: Optional[List[uuid.UUID]] = None,
        filter_to: Optional[List[ArtifactType]] = None,
    ) -> List[ArtifactMetadata]:
        """Returns all artifacts in the DAG with the following optional filters:

        Args:
            `on_op_ids`: only artifacts that are the outputs of these operators are included.
        """
        artifacts = [artifact for artifact in self.artifacts.values()]

        if on_op_ids is not None:
            operators = [self.must_get_operator(op_id) for op_id in on_op_ids]
            artifact_ids = set()
            for op in operators:
                artifact_ids.update(op.outputs)
            artifacts = self.must_get_artifacts(list(artifact_ids))

        if filter_to is not None:
            artifacts = [artifact for artifact in artifacts if artifact.type in filter_to]

        return artifacts

    def get_unclaimed_op_name(self, prefix: str) -> str:
        """Returns an operator name that is guaranteed to not collide with any existing name in the dag.

        Starts with the operator name `<prefix> 1`. If it is taken, we continue to increment the suffix counter
        until we hit an unclaimed name.
        """
        curr_suffix = 1
        while True:
            candidate_name = prefix + " %d" % curr_suffix
            colliding_op = self.get_operator(with_name=candidate_name)
            if colliding_op is None:
                # We've found an unallocated name!
                op_name = candidate_name
                break
            curr_suffix += 1

        assert op_name is not None
        return op_name

    ######################## DAG WRITES #############################

    def add_operator(self, op: Operator) -> None:
        self.add_operators([op])

    def add_operators(self, ops: List[Operator]) -> None:
        for op in ops:
            self.operators[str(op.id)] = op
            self.operator_by_name[op.name] = op

    def add_artifacts(self, artifacts: List[ArtifactMetadata]) -> None:
        for artifact in artifacts:
            self.artifacts[str(artifact.id)] = artifact

    def update_artifact_type(self, artifact_id: uuid.UUID, artifact_type: ArtifactType) -> None:
        self.must_get_artifact(artifact_id).type = artifact_type

    def update_operator_spec(self, name: str, spec: OperatorSpec) -> None:
        """Replaces an operator's spec in the dag.

        The assumption validated within the method is that the caller has already validated
        both that the operator exists, and that the spec type will be unchanged.
        """
        assert name in self.operator_by_name, "Operator %s does not exist." % name
        op = self.operator_by_name[name]
        assert get_operator_type(op) == get_operator_type_from_spec(
            spec
        ), "New spec has a different type."

        self.operators[str(op.id)].spec = spec
        self.operator_by_name[op.name].spec = spec

    def update_operator_function(self, operator: Operator, serialized_function: bytes) -> None:
        if operator in self.operators.values():
            operator.update_serialized_function(serialized_function)

    def remove_operator(
        self,
        operator_id: uuid.UUID,
        must_be_type: Optional[OperatorType] = None,
    ) -> None:
        """Deletes the given operator from the DAG, along with any direct output artifacts.

        Args:
            operator_id:
                The operator to delete (and to start deletion at)
            must_be_type:
                If set, will only delete the given operator if it is of the same operator type.
        """
        self.remove_operators([operator_id], must_be_type)

    def remove_operators(
        self,
        operator_ids: List[uuid.UUID],
        must_be_type: Optional[OperatorType] = None,
    ) -> None:
        """Batch version of `remove_operator()`."""
        for operator_id in operator_ids:
            op_to_remove = self.operators[str(operator_id)]
            if must_be_type is not None and get_operator_type(op_to_remove) != must_be_type:
                raise InternalAqueductError(
                    "Cannot remove operator of type %s, must be of type %s."
                    % (get_operator_type(op_to_remove), must_be_type)
                )

            for artifact_id in op_to_remove.outputs:
                del self.artifacts[str(artifact_id)]
            del self.operators[str(op_to_remove.id)]
            del self.operator_by_name[op_to_remove.name]


# Initialize a module-level dag object, to be accessed and modified when the user construct the flow.
__GLOBAL_DAG__ = DAG(metadata=Metadata())
