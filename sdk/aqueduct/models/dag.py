import copy
import uuid
from typing import Any, Callable, Dict, List, Optional, Set

from aqueduct.constants.enums import (
    ArtifactType,
    OperatorType,
    RuntimeType,
    SparkRuntimeType,
    TriggerType,
)
from aqueduct.error import (
    ArtifactNotFoundException,
    InternalAqueductError,
    InvalidUserActionException,
)
from pydantic import BaseModel

from ..logger import logger
from ..utils.naming import bump_artifact_suffix
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

    # The field must be set when publishing the workflow.
    metadata: Metadata

    # Represents the default engine the DAG will be executed on. Can be overwritten
    # by individual operators.
    engine_config: EngineConfig = EngineConfig()

    def validate_and_resolve_artifact_names(self) -> None:
        """To be called from publish_flow() only.

        Checks that all explicitly named artifacts are unique.

        For implicitly named artifacts that collide, bumps their names using the `(num)` suffix.
        The index is grouped by the input operator, so an operator with multiple output artifacts
        that use our default naming scheme will always have consecutive numbers. Eg. (1), (2), (3)
        """

        # In the first pass, check that there aren't any explicitly named artifacts that collide with each other.
        # Add any explicitly named artifacts to `seen_artifact_names`. Those names are now completely claimed.
        seen_artifact_names: Set[str] = set()
        for artifact in self.artifacts.values():
            if artifact.explicitly_named:
                if artifact.name in seen_artifact_names:
                    raise InvalidUserActionException(
                        "Unable to publish flow. You are attempting to publish multiple artifacts explicitly named `%s`. "
                        "Please use `artifact.set_name(<new name>)` to resolve this naming collision. Or rerun the operators "
                        "with different output artifact names." % artifact.name
                    )
                seen_artifact_names.add(artifact.name)

        # In the second pass, resolve the names of any implicitly named artifacts that collide. Loop through
        # the operators in topological order so that numbers are assigned in a reasonable fashion. Output artifacts
        # of the same operator are always numbered consecutively.
        # For some determinism around naming, we sort the starting operators by name and then id.
        q: List[Operator] = sorted(
            [op for op in self.operators.values() if len(op.inputs) == 0],
            key=lambda op: (op.name, op.id),
        )
        seen_op_ids: List[uuid.UUID] = []

        any_op_was_renamed = False
        while len(q) > 0:
            curr_op = q.pop(0)

            # Only traverse operators you haven't seen before.
            if curr_op.id in seen_op_ids:
                continue
            seen_op_ids.append(curr_op.id)

            output_artifacts = self.must_get_artifacts(curr_op.outputs)
            for artifact in output_artifacts:
                # Skip name resolution for explicitly named artifacts.
                # We've already checked in the first pass.
                if not artifact.explicitly_named:
                    # Find an unallocated name for each artifact.
                    original_name = artifact.name
                    while artifact.name in seen_artifact_names:
                        artifact.name = bump_artifact_suffix(artifact.name)

                    if original_name != artifact.name:
                        logger().warning(
                            "Multiple artifacts were named `%s`. Since artifact names must be unique, "
                            "we renamed one of them to `%s`." % (original_name, artifact.name)
                        )
                        any_op_was_renamed = True

                    seen_artifact_names.add(artifact.name)

                q.extend(self.list_operators(on_artifact_id=artifact.id))

        if any_op_was_renamed:
            logger().warning(
                "Note that any artifacts you explicitly gave a name to were not renamed."
            )

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
                # DAG's expected to run on Spark cannot have different Operator specs.
                if (
                    dag_engine_config.type in SparkRuntimeType
                    and op.spec.engine_config.type not in SparkRuntimeType
                ):
                    raise InvalidUserActionException(
                        "All operators must run on Spark Type Engine. Operator %s is designated to run on custom engine `%s`."
                        % (op.name, op.spec.engine_config.name),
                    )
                # We don't allow individual operators to set Databricks as an engine spec without setting it globally.
                if (
                    dag_engine_config.type not in SparkRuntimeType
                    and op.spec.engine_config.type in SparkRuntimeType
                ):
                    raise InvalidUserActionException(
                        """In order to use a Spark type engine
                        while previewing operators, please set 
                        aqueduct.global_config({'engine': '<spark_type_engine_name>'})""",
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
        with_output_artifact_id: Optional[uuid.UUID] = None,
    ) -> Operator:
        op = self.get_operator(with_id, with_output_artifact_id)
        if op is None:
            raise InternalAqueductError(
                "Unable to find operator: with_id %s, with_output_artifact_id %s"
                % (str(with_id), str(with_output_artifact_id)),
            )
        return op

    def get_operator(
        self,
        with_id: Optional[uuid.UUID] = None,
        with_output_artifact_id: Optional[uuid.UUID] = None,
        with_input_artifact_ids: Optional[List[uuid.UUID]] = None,
    ) -> Optional[Operator]:
        if (int(with_id is not None) + int(with_output_artifact_id is not None)) + int(
            with_input_artifact_ids is not None
        ) != 1:
            raise InternalAqueductError(
                "Cannot fetch operator with zero or multiple search parameters set."
            )

        if with_id is not None:
            return self.operators.get(str(with_id))
        elif with_output_artifact_id:
            # Search with output artifact id
            for _, op in self.operators.items():
                if with_output_artifact_id in op.outputs:
                    return op
        elif with_input_artifact_ids:
            for _, op in self.operators.items():
                if set(with_input_artifact_ids) == set(op.inputs):
                    return op

        return None

    def get_colliding_metric_or_check(self, candidate_op: Operator) -> Optional[Operator]:
        """A metric or check is considered to be colliding only if it has the same name and input artifacts
        as another metric or check. Metrics can only collide with other metrics, and checks can only collide
        with checks.

        Assumes that only one of these collisions exists.
        """
        assert get_operator_type(candidate_op) in [
            OperatorType.CHECK,
            OperatorType.METRIC,
            OperatorType.SYSTEM_METRIC,
        ]

        match = [
            op
            for op in self.operators.values()
            if (
                get_operator_type(op) == get_operator_type(candidate_op)
                and set(op.inputs) == set(candidate_op.inputs)
                and op.name == candidate_op.name
            )
        ]
        assert len(match) < 2, (
            "We should not be having multiple %s's with the same name and input artifacts."
            % get_operator_type(candidate_op)
        )
        return match[0] if len(match) == 1 else None

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
            raise ArtifactNotFoundException(
                "Artifact has been overwritten and no longer exists. This happens with "
                "metrics and checks when a metric/check with the same name is run on the "
                "same input artifacts. If this is not the case, please file a bug on us!"
            )
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

    def get_param_op_by_name(self, name: str) -> Optional[Operator]:
        for op in self.operators.values():
            if op.name == name and get_operator_type(op) == OperatorType.PARAM:
                return op
        return None

    ######################## DAG WRITES #############################

    def add_operator(self, op: Operator) -> None:
        self.add_operators([op])

    def add_operators(self, ops: List[Operator]) -> None:
        for op in ops:
            self.operators[str(op.id)] = op

    def add_artifacts(self, artifacts: List[ArtifactMetadata]) -> None:
        for artifact in artifacts:
            self.artifacts[str(artifact.id)] = artifact

    def update_artifact_type(self, artifact_id: uuid.UUID, artifact_type: ArtifactType) -> None:
        self.must_get_artifact(artifact_id).type = artifact_type

    def update_artifact_name(self, artifact_id: uuid.UUID, new_name: str) -> None:
        """Updates an artifact to have a user-specified name. This means the artifact is
        now explicitly named.
        """
        artifact = self.must_get_artifact(artifact_id)
        artifact.name = new_name
        artifact.explicitly_named = True

    def update_param_spec(self, name: str, new_spec: OperatorSpec) -> None:
        """Checks that:
        1) The parameter already exists, and there is not more than one with the same name.
        2) The spec is a parameter spec.

        The new parameter's value must also be of the same type, but we enforce during execution.
        """
        assert get_operator_type_from_spec(new_spec) == OperatorType.PARAM
        param_op = self.get_param_op_by_name(name)
        assert param_op is not None
        assert get_operator_type(param_op) == OperatorType.PARAM

        self.operators[str(param_op.id)].spec = new_spec

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
