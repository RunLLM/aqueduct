import uuid
from typing import Dict, List, Optional

from aqueduct.constants.enums import ArtifactType, OperatorType, RuntimeType, TriggerType
from aqueduct.error import (
    ArtifactNotFoundException,
    InternalAqueductError,
    InvalidUserActionException,
)
from pydantic import BaseModel

from .artifact import ArtifactMetadata
from .config import EngineConfig
from .dag_rules import check_customized_resources_are_supported
from .operators import Operator, OperatorSpec, get_operator_type, get_operator_type_from_spec


class Schedule(BaseModel):
    trigger: Optional[TriggerType] = None
    cron_schedule: str = ""
    disable_manual_trigger: bool = False
    # source_id refers to the source flow for this schedule when
    # the trigger is TriggerType.CASCADE
    source_id: Optional[uuid.UUID] = None


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

    # The field must be set when publishing the workflow.
    metadata: Metadata

    # Represents the default engine the DAG will be executed on. Can we overwritten
    # by individual operators.
    engine_config: EngineConfig = EngineConfig()

    class Config:
        fields = {
            "operator_by_name": {"exclude": ...},
        }

    def set_engine_config(
        self,
        global_engine_config: Optional[EngineConfig],
        publish_flow_engine_config: Optional[EngineConfig] = None,
    ) -> None:
        """Sets the engine config on the dag.

        The hierarchy of engine selection is:
        1) @op(engine=...), which is not set on the DAG, but instead is found on the operator spec.
        2) client.publish_flow(.., engine=...)
        3) aq.global_config(engine=...)

        Before setting the config, we need to perform the following checks on each operator:
        - Check if the operator's compute engine can handle any specified resource requests.
        """
        dag_engine_config = EngineConfig()
        if global_engine_config is not None:
            dag_engine_config = global_engine_config
        if publish_flow_engine_config is not None:
            dag_engine_config = publish_flow_engine_config

        for op in self.operators.values():
            op_engine_config = dag_engine_config
            if op.spec.engine_config is not None:
                # DAG's that are expected to execute on Airflow cannot have any custom Operator specs.
                if dag_engine_config.type == RuntimeType.AIRFLOW:
                    raise InvalidUserActionException(
                        "All operators must run on Airflow. Operator %s is designated to run on custom engine `%s`."
                        % (op.name, op.spec.engine_config.name),
                    )
                # DAG's expected to run on Databricks cannot have different Operator specs.
                if (
                    dag_engine_config.type == RuntimeType.DATABRICKS
                    and op.spec.engine_config.type != RuntimeType.DATABRICKS
                ):
                    raise InvalidUserActionException(
                        "All operators must run on Databricks. Operator %s is designated to run on custom engine `%s`."
                        % (op.name, op.spec.engine_config.name),
                    )
                # We don't allow individual operators to set Databricks as an engine spec without setting it globally.
                if (
                    dag_engine_config.type != RuntimeType.DATABRICKS
                    and op.spec.engine_config.type == RuntimeType.DATABRICKS
                ):
                    raise InvalidUserActionException(
                        """In order to use 
                        Databricks while previewing operators, please set 
                        aqueduct.global_config({'engine': '<databricks_integration>'})""",
                    )

                op_engine_config = op.spec.engine_config

            # Since we know exactly what engine the operator will run with, check whether
            # the custom resource constraints are valid.
            if op.spec.resources is not None:
                check_customized_resources_are_supported(
                    op.spec.resources, op_engine_config, op.name
                )

        self.engine_config = dag_engine_config

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
        artifact = self.get_artifact(artifact_id)
        if artifact is None:
            raise ArtifactNotFoundException("Unable to find artifact.")
        return artifact

    def get_artifact(self, artifact_id: uuid.UUID) -> Optional[ArtifactMetadata]:
        return self.artifacts.get(str(artifact_id))

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

    def validate_artifact_name(self, name: str) -> None:
        """Checks that the artifact name is unique."""
        existing = self.get_artifact_by_name(name)
        if existing is not None:
            raise InvalidUserActionException(
                "Artifact with name `%s` has already been created locally. Artifact names must be unique."
                % name,
            )

    def list_metrics_for_operator(self, op: Operator) -> List[Operator]:
        """Returns all the metric operators on the given operator's outputs."""
        metric_operators = []
        for artf in op.outputs:
            metric_operators.extend(
                self.list_operators(
                    filter_to=[OperatorType.METRIC],
                    on_artifact_id=artf,
                )
            )
        return metric_operators

    def list_checks_for_operator(self, op: Operator) -> List[Operator]:
        """Returns all the check operators on the given operator's outputs."""
        check_operators = []
        for artf in op.outputs:
            check_operators.extend(
                self.list_operators(
                    filter_to=[OperatorType.CHECK],
                    on_artifact_id=artf,
                )
            )
        return check_operators

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

    def update_artifact_name(self, artifact_id: uuid.UUID, new_name: str) -> None:
        self.must_get_artifact(artifact_id).name = new_name

    def update_operator_name(self, op_id: uuid.UUID, new_name: str) -> None:
        # Update the name -> operator map.
        old_name = self.must_get_operator(op_id).name
        self.operator_by_name[new_name] = self.operator_by_name[old_name]
        del self.operator_by_name[old_name]

        # Update the name on the operator spec.
        self.must_get_operator(op_id).name = new_name

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
