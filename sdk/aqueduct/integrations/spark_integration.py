from aqueduct.models.integration import Integration, IntegrationInfo


class SparkIntegration(Integration):
    """
    Class for Spark integration.
    """

    def __init__(self, metadata: IntegrationInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Spark integration."""
        print("==================== Spark Integration  =============================")
        self._metadata.describe()
