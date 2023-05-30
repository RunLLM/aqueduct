from typing import Any, Callable, Dict, List, Optional, Union

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ExecutionStatus

from aqueduct import Client, Flow
from sdk.shared.flow_helpers import publish_flow_test, trigger_flow_test


class FlowManager:
    """This is a convenience class that packages a couple of flow-specific fields together.

    It abstracts away the publishing of flows from data resource test cases, which usually
    don't care about things like flow name and engine, and just want to publish flows as the
    only mechanism for saving data. It is imported into test cases as a fixture, to simplify
    the test signature.
    """

    _client: Client
    _flow_name_fn: Callable[..., str]
    _engine: Optional[str]

    def __init__(self, client, flow_name, engine):
        self._client = client
        self._flow_name_fn = flow_name
        self._engine = engine

    def publish_flow_test(
        self,
        artifacts: Union[BaseArtifact, List[BaseArtifact]],
        expected_statuses: Union[
            ExecutionStatus, List[ExecutionStatus]
        ] = ExecutionStatus.SUCCEEDED,
        existing_flow: Optional[Flow] = None,
    ) -> Flow:
        """This is a simplified wrapper around `publish_flow_test()`, built with data resource test in mind."""
        if existing_flow is not None:
            return publish_flow_test(
                self._client,
                artifacts=artifacts,
                expected_statuses=expected_statuses,
                existing_flow=existing_flow,
                engine=self._engine,
            )
        else:
            return publish_flow_test(
                self._client,
                name=self._flow_name_fn(),
                artifacts=artifacts,
                expected_statuses=expected_statuses,
                engine=self._engine,
            )

    def trigger_flow_test(
        self,
        flow: Flow,
        expected_status: Union[ExecutionStatus, List[ExecutionStatus]] = ExecutionStatus.SUCCEEDED,
        parameters: Optional[Dict[str, Any]] = None,
    ):
        """Convenience function, mostly here for completeness. It's the same as `trigger_flow_test()`, except missing the client argument."""

        trigger_flow_test(
            self._client,
            flow,
            expected_status=expected_status,
            parameters=parameters,
        )
