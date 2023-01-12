from aqueduct.models.integration import Integration, IntegrationInfo


class DatabricksIntegration(Integration):
    """
    Class for Databricks integration.
    """

    def __init__(self, metadata: IntegrationInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Databricks integration."""
        print("==================== Databricks Integration  =============================")
        self._metadata.describe()
