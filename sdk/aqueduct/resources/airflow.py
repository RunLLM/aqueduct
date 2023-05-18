from aqueduct.models.resource import BaseResource, ResourceInfo


class AirflowResource(BaseResource):
    """
    Class for Airflow resource.
    """

    def __init__(self, metadata: ResourceInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Airflow resource."""
        print("==================== Airflow Resource =============================")
        self._metadata.describe()
