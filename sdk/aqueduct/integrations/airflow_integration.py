from aqueduct.integrations.integration import Integration
from aqueduct.models.integration_info import IntegrationInfo


class AirflowIntegration(Integration):
    """
    Class for Airflow integration.
    """

    def __init__(self, metadata: IntegrationInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Airflow integration."""
        print("==================== Airflow Integration  =============================")
        self._metadata.describe()
