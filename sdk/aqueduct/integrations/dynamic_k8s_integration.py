from typing import Dict, Union

from aqueduct.constants.enums import K8sClusterActionType, K8sClusterStatusType
from aqueduct.error import InvalidIntegrationException, InvalidUserArgumentException
from aqueduct.integrations.connect_config import DynamicK8sConfig
from aqueduct.integrations.validation import validate_is_connected
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.response_models import DynamicEngineStatusResponse

from aqueduct import globals


def parse_dynamic_k8s_config(
    config_delta: Union[Dict[str, Union[int, str]], DynamicK8sConfig]
) -> DynamicK8sConfig:
    if not isinstance(config_delta, dict) and not isinstance(config_delta, DynamicK8sConfig):
        raise InvalidUserArgumentException(
            "`config_delta` argument must be either a dict or DynamicK8sConfig."
        )

    if isinstance(config_delta, dict):
        config_delta = DynamicK8sConfig(**config_delta)
    assert isinstance(config_delta, DynamicK8sConfig)
    return config_delta


def validate_engine_record(
    name: str, engine_statuses: Dict[str, DynamicEngineStatusResponse]
) -> None:
    if len(engine_statuses) == 0:
        raise InvalidIntegrationException("Dynamic engine %s does not exist!" % name)

    if len(engine_statuses) > 1:
        raise InvalidIntegrationException("Duplicate dynamic engine with name %s!" % name)


class DynamicK8sIntegration(Integration):
    """
    Class for Dynamic K8s integration.
    """

    def __init__(self, metadata: IntegrationInfo):
        self._metadata = metadata

    @validate_is_connected()
    def status(self) -> str:
        """Get the current status of the dynamic Kubernetes cluster."""
        engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status(
            engine_integration_ids=[str(self._metadata.id)]
        )

        validate_engine_record(self._metadata.name, engine_statuses)

        return engine_statuses[self._metadata.name].status.value

    @validate_is_connected()
    def create(
        self, config_delta: Union[Dict[str, Union[int, str]], DynamicK8sConfig] = {}
    ) -> None:
        """Creates the dynamic Kubernetes cluster, if it is not currently running.

        Args:
            config_delta (optional):
                This field contains new config values to be used in creating the cluster.
                These new values will overwrite existing ones from that point on. Any config values
                that are identical to the current ones do not need to be included in config_delta.

        Raises:
            InvalidIntegrationException:
                An error occurred when the dynamic engine doesn't exist.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        config_delta = parse_dynamic_k8s_config(config_delta)

        engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status(
            engine_integration_ids=[str(self._metadata.id)]
        )

        validate_engine_record(self._metadata.name, engine_statuses)

        status = engine_statuses[self._metadata.name].status
        if status == K8sClusterStatusType.ACTIVE and all(
            value is None for value in config_delta.dict().values()
        ):
            print("Cluster is already in %s status." % status.value)
            return

        print(
            "Cluster is currently in %s status. It could take 12-15 minutes for the cluster to be created or updated..."
            % status.value
        )
        globals.__GLOBAL_API_CLIENT__.edit_dynamic_engine(
            action=K8sClusterActionType.CREATE,
            integration_id=str(self._metadata.id),
            config_delta=config_delta,
        )

    @validate_is_connected()
    def update(self, config_delta: Union[Dict[str, Union[int, str]], DynamicK8sConfig]) -> None:
        """Update the dynamic Kubernetes cluster. This can only be done when the cluster is in
            Active status.

        Args:
            config_delta:
                This field contains new config values to be used in creating the cluster.
                These new values will overwrite existing ones from that point on. Any config values
                that are identical to the current ones do not need to be included in config_delta.

        Raises:
            InvalidIntegrationException:
                An error occurred when the dynamic engine doesn't exist.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        config_delta = parse_dynamic_k8s_config(config_delta)

        engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status(
            engine_integration_ids=[str(self._metadata.id)]
        )

        validate_engine_record(self._metadata.name, engine_statuses)

        status = engine_statuses[self._metadata.name].status
        if status != K8sClusterStatusType.ACTIVE:
            print(
                "Update is only support when the cluster is in %s status, found %s."
                % (K8sClusterStatusType.ACTIVE.value, status.value)
            )
            return

        print(
            "Cluster is currently in %s status. It could take 12-15 minutes for the cluster to be updated..."
            % status.value
        )
        globals.__GLOBAL_API_CLIENT__.edit_dynamic_engine(
            action=K8sClusterActionType.UPDATE,
            integration_id=str(self._metadata.id),
            config_delta=config_delta,
        )

    @validate_is_connected()
    def delete(self, force: bool = False) -> None:
        """Deletes the dynamic Kubernetes cluster if it is running, ignoring the keepalive period.

        Args:
            force:
                By default, if there are any pods in the "Running" or "ContainerCreating" status,
                the deletion process will fail. However, if the flag is set to "True", this check
                will be skipped, allowing the cluster to be deleted despite the presence of such pods.

        Raises:
            InvalidIntegrationException:
                An error occurred when the dynamic engine doesn't exist.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status(
            engine_integration_ids=[str(self._metadata.id)]
        )

        validate_engine_record(self._metadata.name, engine_statuses)

        status = engine_statuses[self._metadata.name].status
        if status == K8sClusterStatusType.TERMINATED:
            print("Cluster is already in %s status." % status.value)
            return

        print(
            "Cluster is currently in %s status. It could take 6-8 minutes for the cluster to be terminated..."
            % status.value
        )

        action = K8sClusterActionType.DELETE
        if force:
            action = K8sClusterActionType.FORCE_DELETE

        globals.__GLOBAL_API_CLIENT__.edit_dynamic_engine(
            action=action, integration_id=str(self._metadata.id)
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the K8s integration."""
        print("==================== Dynamic K8s Integration  =============================")
        self._metadata.describe()
