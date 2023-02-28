from aqueduct.integrations.dynamic_k8s_integration import DynamicK8sIntegration
from aqueduct.models.integration import Integration, IntegrationInfo


class AWSIntegration(Integration):
    """
    Class for AWS integration.
    """

    def __init__(self, metadata: IntegrationInfo, k8s_integration_metadata: IntegrationInfo):
        self._metadata = metadata
        self.k8s = DynamicK8sIntegration(k8s_integration_metadata)

    def describe(self) -> None:
        """Prints out a human-readable description of the K8s integration."""
        print("==================== AWS Integration  =============================")
        self._metadata.describe()
