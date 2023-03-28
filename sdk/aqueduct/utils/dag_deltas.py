import copy
import uuid
from abc import ABC, abstractmethod
from typing import Any, Dict, List, Optional

from aqueduct.constants.enums import OperatorType
from aqueduct.error import (
    InternalAqueductError,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, OperatorSpec, get_operator_type

from .type_inference import infer_artifact_type
from .utils import construct_param_spec


class DAGDelta(ABC):
    """Abstract class for the various types of DAG updates."""

    @abstractmethod
    def apply(self, dag: DAG) -> None:
        pass


class AddOperatorDelta(DAGDelta):
    """Adds an operator and its output artifacts to the DAG.

    Attributes:
        op:
            The new operator to add.
        output_artifacts:
            The output artifacts for this operation.
    """

    def __init__(
        self,
        op: Operator,
        output_artifacts: List[ArtifactMetadata],
    ):
        # Check that the operator's outputs correspond to the given output artifacts.
        assert len(op.outputs) == len(
            output_artifacts
        ), "Number of operator outputs does not match number of given artifacts."

        for i, artifact_id in enumerate(op.outputs):
            assert output_artifacts[i].id == artifact_id, (
                "The %dth output artifact on the operator does not match." % i
            )

        self.op = op
        self.output_artifacts = output_artifacts

    def apply(self, dag: DAG) -> None:
        dag.add_operator(self.op)
        dag.add_artifacts(self.output_artifacts)


class RemoveOperatorDelta(DAGDelta):
    """Removes a given operator, along with all it's downstream dependencies."""

    def __init__(
        self,
        op_id: uuid.UUID,
    ):
        self.op_id = op_id

    def apply(self, dag: DAG) -> None:
        dag.remove_operators(
            operator_ids=dag.list_downstream_operators(op_id=self.op_id),
        )


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
    - every parameter corresponds to a single parameter artifact in the dag.
    - every parameter name is a string.
    - any parameter feeding into a sql query must have a string value (to resolve tags within the query).

    Raises:
        InvalidUserArgumentException:
            If any of the above checks are violated.
    """
    if any(not isinstance(name, str) for name in parameters):
        raise InvalidUserArgumentException("Parameters must be keyed by strings.")

    for param_name, param_val in parameters.items():
        param_op = dag.get_param_op_by_name(param_name)
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
                    "Parameter `%s` is used by a sql query, so it must be a string type, not type %s."
                    % (param_name, type(param_val).__name__)
                )


class UpdateParametersDelta(DAGDelta):
    """Updates the values of the given parameters in the DAG to the given values. No-ops if no parameters provided.

    The parameters are expected to have already been serialized into strings.
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

            dag.update_param_spec(
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
