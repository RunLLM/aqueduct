from typing import Dict

from aqueduct.models.integration import Integration, IntegrationInfo

from aqueduct import globals


class ECRIntegration(Integration):
    """
    Class for ECR integration.
    """

    def __init__(self, metadata: IntegrationInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the ECR integration."""
        print("==================== ECR Integration  =============================")
        self._metadata.describe()

    def image(self, image_name: str) -> Dict[str, str]:
        """
        Returns a dictionary with the name of the ECR resource and the image url, which can be
        used as input to the `image` field of an operator's decorator. This method also verifies
        that the image exists in the ECR repository.
        """
        response = globals.__GLOBAL_API_CLIENT__.get_image_url(
            integration_id=str(self._metadata.id),
            service=self._metadata.service,
            image_name=image_name,
        )

        return {
            "registry_name": self._metadata.name,
            "url": response.url,
        }
