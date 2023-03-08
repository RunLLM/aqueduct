from aqueduct.constants.enums import K8sClusterStatusType
from aqueduct.error import InvalidIntegrationException
from aqueduct.models.integration import Integration, IntegrationInfo

from aqueduct import globals


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

    def create(self) -> None:
        engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status(
            engine_integration_ids=[str(self._metadata.id)]
        )
        if len(engine_statuses) != 1:
            raise InvalidIntegrationException(
                "Dynamic engine %s does not exist!" % self._metadata.name
            )

        status = engine_statuses[self._metadata.name].status
        if status == K8sClusterStatusType.ACTIVE:
            print("Cluster is already in %s status." % status.value)
            return

        print(
            "Cluster is currently in %s status. It could take 12-15 minutes for the cluster to be ready..."
            % status.value
        )
        globals.__GLOBAL_API_CLIENT__.edit_engine(
            action="create", integration_id=str(self._metadata.id)
        )

    def delete(self) -> None:
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
        globals.__GLOBAL_API_CLIENT__.edit_engine(
            action="delete", integration_id=str(self._metadata.id)
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the K8s integration."""
        print("==================== Dynamic K8s Integration  =============================")
        self._metadata.describe()
