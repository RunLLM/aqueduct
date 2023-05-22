from aqueduct.models.resource import BaseResource, ResourceInfo


class SparkResource(BaseResource):
    """
    Class for Spark resource.
    """

    def __init__(self, metadata: ResourceInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Spark resource."""
        print("==================== Spark Resource =============================")
        self._metadata.describe()
