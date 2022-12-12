import uuid

from aqueduct.constants.enums import ArtifactType, OperatorType
from aqueduct.error import (
    InvalidIntegrationException,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from aqueduct.globals import __GLOBAL_API_CLIENT__ as global_api_client
from aqueduct.models.dag import DAG
from aqueduct.models.integration import IntegrationInfo
from aqueduct.models.operators import (
    LoadSpec,
    Operator,
    OperatorSpec,
    S3LoadParams,
    UnionLoadParams,
)
from aqueduct.utils.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.utils import generate_uuid


def save_artifact(
    artifact_id: uuid.UUID,
    artifact_type: ArtifactType,
    dag: DAG,
    integration_info: IntegrationInfo,
    save_params: UnionLoadParams,
) -> None:
    """Configures the given artifact to be written to a specific integration after it's computed in a published flow.

    TODO(ENG-2035): Move this method into the base integration object.

    Args:
        artifact_id:
            The artifact who's contents will be saved.
        artifact_type:
            The type of the given artifact.
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
        raise InvalidIntegrationException("Not connected to db %s!" % integration_info.name)

    # Non-tabular data cannot be saved into relational data stores.
    if (
        artifact_type not in [ArtifactType.UNTYPED, ArtifactType.TABLE]
        and integration_info.is_relational()
    ):
        raise InvalidUserActionException(
            "Unable to save non-relational data into relational data store `%s`."
            % integration_info.name
        )

    # Tabular data written into S3 must include a S3FileFormat hint.
    if artifact_type == ArtifactType.TABLE and isinstance(save_params, S3LoadParams):
        if save_params.format is None:
            raise InvalidUserArgumentException(
                "You must supply a file format when saving tabular data into S3 integration `%s`."
                % integration_info.name
            )

    # We currently do not allow multiple load operators on the same artifact to the same integration.
    # We do allow multiple artifacts to write to the same integration, as well as a single artifact
    # to write to multiple integrations.
    # Multiple load operations to the same integration, of different artifacts, are guaranteed to
    # have unique names.
    load_op_name = None

    # Replace any existing save operator on this artifact that goes to the same integration.
    existing_load_ops = dag.list_operators(
        filter_to=[OperatorType.LOAD],
        on_artifact_id=artifact_id,
    )
    for op in existing_load_ops:
        assert op.spec.load is not None
        if op.spec.load.integration_id == integration_info.id:
            load_op_name = op.name

    # If the name is not set yet, we know we have to make a new load operator, so bump the
    # suffix until a unique name is found.
    if load_op_name is None:
        load_op_name = dag.get_unclaimed_op_name(prefix="save to %s" % integration_info.name)

    apply_deltas_to_dag(
        dag,
        deltas=[
            AddOrReplaceOperatorDelta(
                op=Operator(
                    id=generate_uuid(),
                    name=load_op_name,
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
            ),
        ],
    )
