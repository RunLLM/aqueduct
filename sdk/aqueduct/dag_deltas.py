import copy
import uuid
from abc import ABC, abstractmethod
from typing import Any, Callable, Dict, List, Optional

from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.dag import DAG
from aqueduct.enums import OperatorType
from aqueduct.error import (
    InternalAqueductError,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from aqueduct.logger import logger
from aqueduct.operators import Operator, OperatorSpec, get_operator_type
from aqueduct.utils import construct_param_spec, infer_artifact_type


class DAGDelta(ABC):
    """Abstract class for the various types of DAG updates."""

    @abstractmethod
    def apply(self, dag: DAG) -> None:
        pass


# These helpers are meant to be fed into `AddOrReplaceOperatorDelta` as different ways
# of resolving whether an operator already exists in the DAG or not.


def find_duplicate_operator_by_name(dag: DAG, op: Operator) -> Optional[Operator]:
    return dag.get_operator(with_name=op.name)


def find_duplicate_load_operator(dag: DAG, op: Operator) -> Optional[Operator]:
    """Load operators are only duplicates if they are loading the same artifact into the same integration."""
    assert get_operator_type(op) == OperatorType.LOAD
    assert len(op.inputs) == 1

    artifact_to_load = dag.must_get_artifact(op.inputs[0])
    existing_load_ops = dag.list_operators(
        filter_to=[OperatorType.LOAD], on_artifact_id=artifact_to_load.id
    )
    for existing_load_op in existing_load_ops:
        if existing_load_op.name == op.name:
            return existing_load_op
    return None


class AddOrReplaceOperatorDelta(DAGDelta):
    """Adds an operator and its output artifacts to the DAG.

    If the operator name already exists in the dag, we will remove the old, colliding
    operator, along with all its downstream dependencies before adding the operator.

    Attributes:
        op:
            The new operator to add.
        output_artifacts:
            The output artifacts for this operation.
        find_duplicate_fn:
            A caller-supplied function that defines when we want the new operator to replace
            and old one. Returns the operator to replace, or None if no collision is found.
            Defaults to replacing an operator with the same name.
    """

    def __init__(
        self,
        op: Operator,
        output_artifacts: List[ArtifactMetadata],
        find_duplicate_fn: Callable[
            [DAG, Operator], Optional[Operator]
        ] = find_duplicate_operator_by_name,
    ):
        # Check that the operator's outputs correspond to the given output artifacts.
        if len(op.outputs) != len(output_artifacts):
            raise InternalAqueductError(
                "Number of operator outputs does not match number of given artifacts."
            )

        for i, artifact_id in enumerate(op.outputs):
            if output_artifacts[i].id != artifact_id:
                raise InternalAqueductError(
                    "The %dth output artifact on the operator does not match." % i
                )

        self.op = op
        self.output_artifacts = output_artifacts
        self.find_duplicate_fn = find_duplicate_fn

    def apply(self, dag: DAG) -> None:
        # Find any colliding operator, and remove it and its dependencies first!
        colliding_op = self.find_duplicate_fn(dag, self.op)
        if colliding_op is not None:
            if get_operator_type(self.op) != get_operator_type(colliding_op):
                raise InvalidUserActionException(
                    "Attempting to replace operator `%s` with a new operator `%s` of type %s, "
                    "but the existing operator has type %s."
                    % (
                        colliding_op.name,
                        self.op.name,
                        get_operator_type(self.op),
                        get_operator_type(colliding_op),
                    ),
                )

            # The colliding operator cannot be a dependency of the new operator. Otherwise, we would
            # not be able to remove the colliding operator.
            downstream_op_ids = dag.list_downstream_operators(colliding_op.id)
            for op_id in downstream_op_ids:
                downstream_op = dag.must_get_operator(op_id)
                if len(set(downstream_op.outputs).intersection(set(self.op.inputs))) > 0:
                    raise InvalidUserActionException(
                        "Attempting to replace operator `%s`, but it cannot be overwritten "
                        "because it is an upstream dependency of the new operator `%s`."
                        % (colliding_op.name, self.op.name)
                    )

            logger().info(
                "The previously defined operator `%s` is being overwritten. Any downstream "
                "artifacts of that operator will need to be recomputed and re-saved." % self.op.name
            )
            for op_id in downstream_op_ids:
                dag.remove_operator(op_id)

        dag.add_operator(self.op)
        dag.add_artifacts(self.output_artifacts)


class SubgraphDAGDelta(DAGDelta):
    """
    Computes a valid subgraph of the given DAG, where every terminal artifact or
    operator must have been explicitly requested.

    Attributes:
        artifact_ids:
            These artifacts describe what our returned subgraph will look like:
            only these artifacts can be terminal nodes.
        include_saves:
            Whether to implicitly include all saves on all artifacts in the subgraph.
        include_metrics:
            Whether to implicitly include all metrics on all artifacts in the subgraph.
            This means all dependencies of such metrics will be included, even if they
            were not part of the original subgraph. If false, all metric operators not
            explicitly defined in `artifact_ids` will be excluded.
        include_checks:
            The checks version of `include_metrics`.
    """

    def __init__(
        self,
        artifact_ids: Optional[List[uuid.UUID]] = None,
        include_saves: bool = False,
        include_metrics: bool = False,
        include_checks: bool = False,
    ):
        if artifact_ids is None or len(artifact_ids) == 0:
            raise InternalAqueductError("Must set artifact ids when pruning dag.")

        self.artifact_ids: List[uuid.UUID] = [] if artifact_ids is None else artifact_ids
        self.include_saves = include_saves
        self.include_metrics = include_metrics
        self.include_checks = include_checks

    def apply(self, dag: DAG) -> None:
        # Check that all the artifact ids exist in the dag.
        for artifact_id in self.artifact_ids:
            _ = dag.must_get_artifact(artifact_id)

        # Starting at the terminal artifacts, perform a DFS in the reverse direction to find
        # all upstream artifacts.
        upstream_artifact_ids = set()
        load_operator_ids = []

        q: List[uuid.UUID] = copy.copy(self.artifact_ids)
        seen_artifact_ids = set(q)
        while len(q) > 0:
            curr_artifact_id = q.pop(0)

            # Keep the extract/function operators along the path.
            upstream_artifact_ids.add(curr_artifact_id)

            # If requested, keep load operators on all artifacts along the way.
            if self.include_saves:
                load_ops = dag.list_operators(
                    filter_to=[OperatorType.LOAD], on_artifact_id=curr_artifact_id
                )
                load_operator_ids.extend([op.id for op in load_ops])

            # The operator who's output is the current artifact.
            curr_op = dag.must_get_operator(with_output_artifact_id=curr_artifact_id)
            candidate_next_artifact_ids = copy.copy(curr_op.inputs)

            implicit_types_to_include = []
            if self.include_checks:
                implicit_types_to_include.append(OperatorType.CHECK)
            if self.include_metrics:
                implicit_types_to_include.append(OperatorType.METRIC)

            if len(implicit_types_to_include) > 0:
                check_or_metric_ops = dag.list_operators(
                    on_artifact_id=curr_artifact_id,
                    filter_to=implicit_types_to_include,
                )
                check_or_metric_artifacts = dag.list_artifacts(
                    on_op_ids=[op.id for op in check_or_metric_ops]
                )
                candidate_next_artifact_ids.extend(
                    [artifact.id for artifact in check_or_metric_artifacts]
                )

            # Prune the upstream candidates against our "already seen" group.
            next_artifact_ids = set(candidate_next_artifact_ids).difference(seen_artifact_ids)

            # Update the queue and seen groups.
            q.extend(next_artifact_ids)
            seen_artifact_ids.update(next_artifact_ids)

        # Remove all operators and artifacts not in the upstream DAG we computed above.
        all_op_ids = set(op.id for op in dag.list_operators())
        upstream_ops = [
            dag.must_get_operator(with_output_artifact_id=artifact_id)
            for artifact_id in upstream_artifact_ids
        ]
        upstream_op_ids = [op.id for op in upstream_ops] + load_operator_ids
        dag.remove_operators(list(all_op_ids.difference(upstream_op_ids)))


class RemoveCheckOperatorDelta(DAGDelta):
    """Removes the check operator on the given artifact that has the given name.

    Raises:
        InvalidUserActionException: if a matching check operator could not be found.
    """

    def __init__(
        self,
        check_name: str,
        artifact_id: uuid.UUID,
    ):
        self.check_name = check_name
        self.artifact_id = artifact_id

    def apply(self, dag: DAG) -> None:
        check_ops = dag.list_operators(
            filter_to=[OperatorType.CHECK], on_artifact_id=self.artifact_id
        )
        check_names_for_op = [op.name for op in check_ops]

        assert len(set(check_names_for_op)) == len(
            check_names_for_op
        ), "Check operator names must be unique."

        found: bool = False
        for i, name in enumerate(check_names_for_op):
            if name == self.check_name:
                found = True
                dag.remove_operator(operator_id=check_ops[i].id, must_be_type=OperatorType.CHECK)

        if not found:
            raise InvalidUserActionException(
                "No check with name %s exists on artifact!" % self.check_name
            )


def validate_overwriting_parameters(dag: DAG, parameters: Dict[str, Any]) -> None:
    """Validates any parameters the user supplies that override the default value.

    The following checks are performed:
    - every parameter corresponds to a parameter artifact in the dag.
    - every parameter name is a string.
    - any parameter feeding into a sql query must have a string value (to resolve tags within the query).

    Raises:
        InvalidUserArgumentException:
            If any of the above checks are violated.
    """
    if any(not isinstance(name, str) for name in parameters):
        raise InvalidUserArgumentException("Parameters must be keyed by strings.")

    for param_name, param_val in parameters.items():
        param_op = dag.get_operator(with_name=param_name)
        if param_op is None:
            raise InvalidUserArgumentException(
                "Parameter %s cannot be found, or is not utilized in the current computation."
                % param_name
            )
        if get_operator_type(param_op) != OperatorType.PARAM:
            raise InvalidUserArgumentException(
                "Parameter %s must refer to a parameter, but instead refers to a: %s"
                % (param_name, get_operator_type(param_op))
            )

        # Any parameter that is consumed by a SQL operator must be a string type!
        assert len(param_op.outputs) == 1
        param_artifact_id = param_op.outputs[0]
        ops_on_param = dag.list_operators(on_artifact_id=param_artifact_id)
        if any(get_operator_type(op) == OperatorType.EXTRACT for op in ops_on_param):
            if not isinstance(param_val, str):
                raise InvalidUserArgumentException(
                    "Parameter %s is used by a sql query, so it must be a string type, not type %s."
                    % (param_name, type(param_val).__name__)
                )


class UpdateParametersDelta(DAGDelta):
    """Updates the values of the given parameters in the DAG to the given values. No-ops if no parameters provided.

    The parameters are expected to have already been serialized into strings, and validated.
    """

    def __init__(
        self,
        parameters: Optional[Dict[str, Any]],
    ):
        self.parameters = parameters

    def apply(self, dag: DAG) -> None:
        if self.parameters is None:
            return
        validate_overwriting_parameters(dag, self.parameters)

        for param_name, new_val in self.parameters.items():
            artifact_type = infer_artifact_type(new_val)
            param_spec = construct_param_spec(new_val, artifact_type)

            dag.update_operator_spec(
                param_name,
                OperatorSpec(
                    param=param_spec,
                ),
            )


def apply_deltas_to_dag(dag: DAG, deltas: List[DAGDelta], make_copy: bool = False) -> DAG:
    if make_copy:
        dag = copy.deepcopy(dag)

    for delta in deltas:
        delta.apply(dag)

    return dag
