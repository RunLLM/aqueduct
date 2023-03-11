import uuid
from typing import List

from aqueduct.constants.enums import OperatorType
from aqueduct.error import InvalidIntegrationException
from aqueduct.globals import __GLOBAL_API_CLIENT__ as global_api_client
from aqueduct.models.dag import DAG
from aqueduct.models.integration import IntegrationInfo
from aqueduct.models.operators import LoadSpec, Operator, OperatorSpec, UnionLoadParams
from aqueduct.utils.dag_deltas import (
    AddOperatorDelta,
    DAGDelta,
    RemoveOperatorDelta,
    apply_deltas_to_dag,
)
from aqueduct.utils.utils import generate_uuid


def _save_artifact(
    artifact_id: uuid.UUID,
    dag: DAG,
    integration_info: IntegrationInfo,
    save_params: UnionLoadParams,
) -> None:
    """Configures the given artifact to be written to a specific integration after it's computed in a published flow.

    Args:
        artifact_id:
            The artifact who's contents will be saved.
        dag:
            The dag object that we will attach the load operator to.
        integration_info:
            Config info for the destination integration.
        save_params:
            Save configuration info (eg. table name, update mode).

    Raises:
        InvalidIntegrationException:
            An error occurred because the requested integration could not be
            found.
        InvalidUserActionException:
            An error occurred because you are trying to load non-relational data into a relational integration.
        InvalidUserArgumentException:
            An error occurred because some necessary fields are missing in the SaveParams.
    """

    integrations_map = global_api_client.list_integrations()
    if integration_info.name not in integrations_map:
        raise InvalidIntegrationException(
            "Not connected to integration %s!" % integration_info.name
        )

    # We currently do not allow multiple save operators on the same artifact to the same integration.
    # We do allow multiple artifacts to write to the same integration, as well as a single artifact
    # to write to multiple integrations.
    save_op_name = "save to %s" % integration_info.name

    # Replace any existing save operator on this artifact that goes to the same integration.
    deltas: List[DAGDelta] = []
    existing_save_ops = dag.list_operators(
        filter_to=[OperatorType.LOAD],
        on_artifact_id=artifact_id,
    )
    for op in existing_save_ops:
        assert op.spec.load is not None
        if op.spec.load.integration_id == integration_info.id:
            deltas.append(RemoveOperatorDelta(op.id))

    deltas.append(
        AddOperatorDelta(
            op=Operator(
                id=generate_uuid(),
                name=save_op_name,
                description="",
                spec=OperatorSpec(
                    load=LoadSpec(
                        service=integration_info.service,
                        integration_id=integration_info.id,
                        parameters=save_params,
                    )
                ),
                inputs=[artifact_id],
            ),
            output_artifacts=[],
        )
    )

    apply_deltas_to_dag(
        dag,
        deltas=deltas,
    )
