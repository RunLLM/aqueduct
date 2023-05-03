from aqueduct.models.integration import BaseResource, ResourceInfo


class SparkResource(BaseResource):
    """
    Class for Spark integration.
    """

    def __init__(self, metadata: ResourceInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Spark integration."""
        print("==================== Spark Resource =============================")
        self._metadata.describe()
