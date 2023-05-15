from aqueduct.models.integration import BaseResource, ResourceInfo


class AirflowResource(BaseResource):
    """
    Class for Airflow integration.
    """

    def __init__(self, metadata: ResourceInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Airflow integration."""
        print("==================== Airflow Resource =============================")
        self._metadata.describe()
