import uuid

from aqueduct.api_client import APIClient
from aqueduct.dag import DAG


class FlowRun:
    """This class is a read-only handle corresponding to a single workflow run in the system."""

    def __init__(
        self,
        api_client: APIClient,
        run_id: str,
        in_notebook_or_console_context: bool,
        dag: DAG,
    ):
        assert run_id is not None
        self._api_client = api_client
        self._id = run_id
        self._in_notebook_or_console_context = in_notebook_or_console_context
        self._dag = dag

    def id(self) -> uuid.UUID:
        """Returns the id for this flow run."""
        return uuid.UUID(self._id)

    def describe(self):
        pass

