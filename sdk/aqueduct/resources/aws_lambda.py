from aqueduct.models.integration import BaseResource, ResourceInfo


class LambdaResource(BaseResource):
    """
    Class for K8s integration.
    """

    def __init__(self, metadata: ResourceInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the K8s integration."""
        print("==================== Lambda Resource =============================")
        self._metadata.describe()
