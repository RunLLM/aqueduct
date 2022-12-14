from aqueduct.backend.api_client import APIClient

# Initialize an unconfigured api client. It will be configured when the user construct an Aqueduct client.
__GLOBAL_API_CLIENT__ = APIClient()
