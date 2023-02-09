from __future__ import annotations

import warnings
from typing import Any

from aqueduct.constants.enums import OperatorType

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.transform import to_artifact_class
from aqueduct.error import InvalidUserArgumentException, InvalidUserActionException
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, OperatorSpec, get_operator_type
from aqueduct.utils.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct.utils.utils import construct_param_spec, generate_uuid


def _operator_is_implicitly_created_param(op: Operator) -> bool:
    return get_operator_type(op) != OperatorType.PARAM and op.spec.param.implicitly_created


def create_param_artifact(
    dag: DAG,
    candidate_name: str,
    default: Any,
    description: str = "",
    is_implicit: bool = False,
) -> BaseArtifact:
    """Creates a parameter operator and return an artifact that can be fed into other operators.

    For implicitly created parameters, the naming collision policy is as follows:
    - We will bump the parameter name until it is unique, unless the colliding operator is
    also a parameter operator that was created implicitly. In that case, we will overwrite it,
    which is more natural:

        ```
        @op
        def foo(bar: int):
            ...

        foo(123) # Creates implicit param named `bar`.
        foo(234) # Overwrites the previously created `bar` parameter with new default value 234.
        ```

    For explicitly created parameters (those made with client.create_param()), we need to check
    that it's not overwriting an implicitly created parameter:

        ```
        @op
        def foo(bar: int):
            ...

        foo(123) # Creates implicit param named `bar`.
        client.create_param("bar", default=555) # Throws an error to avoid unexpected overwriting behavior.
        ```
    """
    if default is None:
        raise InvalidUserArgumentException("Parameter default value cannot be None.")

    param_name = candidate_name
    if is_implicit:
        if not dag.is_name_unique(candidate_name):
            # We only overwrite an existing operator in a very particular case, otherwise we'll attempt to find
            # a unique name.
            colliding_op = dag.get_operator(with_name=candidate_name)
            if (
                colliding_op is None or
                not _operator_is_implicitly_created_param(colliding_op)
            ):
                param_name = dag.get_unclaimed_name(prefix=candidate_name)

        warnings.warn(
            """Input to function argument "%s" is not an artifact type. We have implicitly \
created a parameter named "%s" and your input will be used as its default value. This parameter \
will be used when running the function."""
            % (candidate_name, param_name)
        )
    else:
        colliding_op = dag.get_operator(with_name=param_name)
        if colliding_op is not None and _operator_is_implicitly_created_param(colliding_op):
            raise InvalidUserActionException(
                """Unable to create parameter `%s`, since there is an implicitly created parameter with the
                same name. If the old parameter is not longer relevant, you can remove it with `client.delete_param()`
                and rerun this operation. Otherwise, you'll need to rename one of the two.
                """ % param_name,
            )

    artifact_type = infer_artifact_type(default)
    param_spec = construct_param_spec(default, artifact_type, is_implicit=is_implicit)

    operator_id = generate_uuid()
    output_artifact_id = generate_uuid()
    apply_deltas_to_dag(
        dag,
        deltas=[
            AddOrReplaceOperatorDelta(
                op=Operator(
                    id=operator_id,
                    name=param_name,
                    description=description,
                    spec=OperatorSpec(param=param_spec),
                    inputs=[],
                    outputs=[output_artifact_id],
                ),
                output_artifacts=[
                    ArtifactMetadata(
                        id=output_artifact_id,
                        name=param_name,
                        type=artifact_type,
                    ),
                ],
            )
        ],
    )

    return to_artifact_class(dag, output_artifact_id, artifact_type, default)
