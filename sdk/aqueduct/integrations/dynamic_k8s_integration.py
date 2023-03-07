from typing import Dict, Union

from aqueduct.constants.enums import K8sClusterStatusType
from aqueduct.error import InvalidIntegrationException, InvalidUserArgumentException
from aqueduct.integrations.connect_config import DynamicK8sConfig
from aqueduct.models.integration import Integration, IntegrationInfo
from pydantic import Extra

from aqueduct import globals


def parse_dynamic_k8s_config(
    config_delta: Union[Dict[str, str], DynamicK8sConfig]
) -> DynamicK8sConfig:
    if not isinstance(config_delta, dict) and not isinstance(config_delta, DynamicK8sConfig):
        raise InvalidUserArgumentException(
            "`config_delta` argument must be either a dict or DynamicK8sConfig."
        )

    if isinstance(config_delta, dict):
        config_delta = DynamicK8sConfig(**config_delta)
    assert isinstance(config_delta, DynamicK8sConfig)
    return config_delta


class DynamicK8sIntegration(Integration):
    """
    Class for Dynamic K8s integration.
    """

    def __init__(self, metadata: IntegrationInfo):
        self._metadata = metadata

    def status(self) -> str:
        engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status(
            engine_integration_ids=[str(self._metadata.id)]
        )
        if len(engine_statuses) != 1:
            raise InvalidIntegrationException(
                "Dynamic engine %s does not exist!" % self._metadata.name
            )

        return engine_statuses[self._metadata.name].status.value

    def create(self, config_delta: Union[Dict[str, str], DynamicK8sConfig] = {}) -> None:
        config_delta = parse_dynamic_k8s_config(config_delta)

        engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status(
            engine_integration_ids=[str(self._metadata.id)]
        )
        if len(engine_statuses) != 1:
            raise InvalidIntegrationException(
                "Dynamic engine %s does not exist!" % self._metadata.name
            )

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
            action="create",
            integration_id=str(self._metadata.id),
            config_delta=config_delta,
        )

    def update(self, config_delta: Union[Dict[str, str], DynamicK8sConfig] = {}) -> None:
        config_delta = parse_dynamic_k8s_config(config_delta)

        engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status(
            engine_integration_ids=[str(self._metadata.id)]
        )
        if len(engine_statuses) != 1:
            raise InvalidIntegrationException(
                "Dynamic engine %s does not exist!" % self._metadata.name
            )

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
            action="update",
            integration_id=str(self._metadata.id),
            config_delta=config_delta,
        )

    def delete(self, force: bool = False) -> None:
        engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status(
            engine_integration_ids=[str(self._metadata.id)]
        )
        if len(engine_statuses) != 1:
            raise InvalidIntegrationException(
                "Dynamic engine %s does not exist!" % self._metadata.name
            )

        status = engine_statuses[self._metadata.name].status
        if status == K8sClusterStatusType.TERMINATED:
            print("Cluster is already in %s status." % status.value)
            return

        print(
            "Cluster is currently in %s status. It could take 6-8 minutes for the cluster to be terminated..."
            % status.value
        )

        action = "delete"
        if force:
            action = "force-delete"

        globals.__GLOBAL_API_CLIENT__.edit_dynamic_engine(
            action=action, integration_id=str(self._metadata.id)
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the K8s integration."""
        print("==================== Dynamic K8s Integration  =============================")
        self._metadata.describe()
