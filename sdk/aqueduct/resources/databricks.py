from aqueduct.models.resource import BaseResource, ResourceInfo


class DatabricksResource(BaseResource):
    """
    Class for Databricks resource.
    """

    def __init__(self, metadata: ResourceInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Databricks resource."""
        print("==================== Databricks Resource =============================")
        self._metadata.describe()
