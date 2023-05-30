import uuid
from typing import List, Union

from aqueduct.constants.enums import OperatorType
from aqueduct.error import InvalidResourceException
from aqueduct.globals import __GLOBAL_API_CLIENT__ as global_api_client
from aqueduct.models.dag import DAG
from aqueduct.models.operators import LoadSpec, Operator, OperatorSpec, UnionLoadParams
from aqueduct.models.resource import ResourceInfo
from aqueduct.utils.dag_deltas import (
    AddOperatorDelta,
    DAGDelta,
    RemoveOperatorDelta,
    apply_deltas_to_dag,
)
from aqueduct.utils.utils import generate_uuid


def _save_artifact(
    artifact_ids: Union[uuid.UUID, List[uuid.UUID]],
    dag: DAG,
    resource_info: ResourceInfo,
    save_params: UnionLoadParams,
) -> None:
    """Configures the given artifact to be written to a specific resource after it's computed in a published flow.

    Args:
        artifact_ids:
            Can either be a single ID, or any number of IDs. In the latter case, that means that the first n-1 artifacts
            are parameters to the save operators. The nth artifact is the one that will be saved.
        dag:
            The dag object that we will attach the load operator to.
        resource_info:
            Config info for the destination resource.
        save_params:
            Save configuration info (eg. table name, update mode).

    Raises:
        InvalidResourceException:
            An error occurred because the requested resource could not be
            found.
        InvalidUserActionException:
            An error occurred because you are trying to load non-relational data into a relational resource.
        InvalidUserArgumentException:
            An error occurred because some necessary fields are missing in the SaveParams.
    """
    if not isinstance(artifact_ids, list):
        artifact_ids = [artifact_ids]

    resources_map = global_api_client.list_resources()
    if resource_info.name not in resources_map:
        raise InvalidResourceException("Not connected to resource %s!" % resource_info.name)

    # We currently do not allow multiple save operators on the same artifact to the same resource.
    # We do allow multiple artifacts to write to the same resource, as well as a single artifact
    # to write to multiple resources.
    save_op_name = "save to %s" % resource_info.name

    # Replace any existing save operator on this artifact that goes to the same resource.
    artifact_id_to_save = artifact_ids[-1]

    deltas: List[DAGDelta] = []
    existing_save_ops = dag.list_operators(
        filter_to=[OperatorType.LOAD],
        on_artifact_id=artifact_id_to_save,
    )
    for op in existing_save_ops:
        assert op.spec.load is not None
        if op.spec.load.resource_id == resource_info.id:
            deltas.append(RemoveOperatorDelta(op.id))

    deltas.append(
        AddOperatorDelta(
            op=Operator(
                id=generate_uuid(),
                name=save_op_name,
                description="",
                spec=OperatorSpec(
                    load=LoadSpec(
                        service=resource_info.service,
                        resource_id=resource_info.id,
                        parameters=save_params,
                    )
                ),
                inputs=artifact_ids,
            ),
            output_artifacts=[],
        )
    )

    apply_deltas_to_dag(
        dag,
        deltas=deltas,
    )
