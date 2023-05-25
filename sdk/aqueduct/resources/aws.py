from aqueduct.models.resource import BaseResource, ResourceInfo
from aqueduct.resources.dynamic_k8s import DynamicK8sResource


class AWSResource(BaseResource):
    """
    Class for AWS resource.
    """

    def __init__(self, metadata: ResourceInfo, k8s_resource_metadata: ResourceInfo):
        self._metadata = metadata
        self.k8s = DynamicK8sResource(k8s_resource_metadata)

    def describe(self) -> None:
        """Prints out a human-readable description of the K8s resource."""
        print("==================== AWS Resource =============================")
        self._metadata.describe()
