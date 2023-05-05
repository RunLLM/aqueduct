from aqueduct.models.integration import BaseResource, ResourceInfo


class DatabricksResource(BaseResource):
    """
    Class for Databricks integration.
    """

    def __init__(self, metadata: ResourceInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Databricks integration."""
        print("==================== Databricks Resource =============================")
        self._metadata.describe()
