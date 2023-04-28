from typing import Dict

from aqueduct.error import InvalidUserArgumentException
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

        Args:
            image_name: The name of the image to retrieve. Should be in the form of `image:tag`.
            No need to include the endpoint URL such as 123456789012.dkr.ecr.us-east-1.amazonaws.com.
        """
        if len(image_name.split("/")) == 2:
            image_name = image_name.split("/")[1]

        if len(image_name.split(":")) != 2:
            raise InvalidUserArgumentException("Image name must be of the form `image:tag`.")

        response = globals.__GLOBAL_API_CLIENT__.get_image_url(
            integration_id=str(self._metadata.id),
            service=self._metadata.service,
            image_name=image_name,
        )

        return {
            "registry_name": self._metadata.name,
            "url": response.url,
        }
