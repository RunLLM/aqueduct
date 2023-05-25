from aqueduct.models.resource import BaseResource, ResourceInfo


class K8sResource(BaseResource):
    """
    Class for K8s resource.
    """

    def __init__(self, metadata: ResourceInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the K8s resource."""
        print("==================== K8s Resource =============================")
        self._metadata.describe()
