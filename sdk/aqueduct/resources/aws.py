from aqueduct.resources.dynamic_k8s import DynamicK8sResource
from aqueduct.models.integration import BaseResource, ResourceInfo


class AWSResource(BaseResource):
    """
    Class for AWS integration.
    """

    def __init__(self, metadata: ResourceInfo, k8s_integration_metadata: ResourceInfo):
        self._metadata = metadata
        self.k8s = DynamicK8sResource(k8s_integration_metadata)

    def describe(self) -> None:
        """Prints out a human-readable description of the K8s integration."""
        print("==================== AWS Resource =============================")
        self._metadata.describe()
