import json
import logging
import uuid
from typing import Any, Dict, List, Union, Optional

from aqueduct.generic_artifact import Artifact as GenericArtifact

from .api_client import APIClient
from .artifact import ArtifactSpec, Artifact
from .dag import DAG, apply_deltas_to_dag, SubgraphDAGDelta, Metadata, AddOrReplaceOperatorDelta
from .enums import ServiceType, RelationalDBServices, OperatorType
from .error import (
    InvalidIntegrationException,
    IncompleteFlowException,
    InvalidUserArgumentException,
)
from .flow import Flow, _show_dag
from .github import Github
from .integrations.integration import IntegrationInfo
from .integrations.sql_integration import RelationalDBIntegration
from .integrations.salesforce_integration import SalesforceIntegration
from .integrations.google_sheets_integration import GoogleSheetsIntegration
from .integrations.s3_integration import S3Integration
from .operators import Operator, ParamSpec, OperatorSpec
from .param_artifact import ParamArtifact
from .utils import (
    schedule_from_cron_string,
    retention_policy_from_latest_runs,
    generate_uuid,
    artifact_name_from_op_name,
)

import __main__ as main
import os


class Client:
    """This class allows users to interact with flows on their Aqueduct cluster."""

    def __init__(
        self,
        api_key: str,
        aqueduct_address: str,
        log_level: int = logging.ERROR,
    ):
        """Creates an instance of Client.

        Args:
            api_key:
                The user unique API key provided by Aqueduct.
            aqueduct_address:
                The address of the Aqueduct Server service.
            log_level:
                A indication of what level and above to print logs from the sdk.
                Defaults to printing error and above only. Types defined in: https://docs.python.org/3/howto/logging.html

        Returns:
            A Client instance.
        """
        logging.basicConfig(level=log_level)
        self._api_client = APIClient(api_key, aqueduct_address)
        self._connected_integrations: Dict[
            str, IntegrationInfo
        ] = self._api_client.list_integrations()
        self._dag = DAG(metadata=Metadata())

        # Will show graph if in an ipynb or Python console, but not if running a Python script.
        self._in_notebook_or_console_context = (not hasattr(main, "__file__")) and (
            not "PYTEST_CURRENT_TEST" in os.environ
        )

    def github(self, repo: str, branch: str = "") -> Github:
        """Retrieves a Github object connecting to specified repos and branch.

        You can only connect to a public repo if you didn't already connect a
        Github account to your Aqueduct account.

        Args:
            repo:
                The full github repo URL, e.g. "my_organization/my_repo"
            branch:
                Optional branch name. The default main branch will be used if not specified.

        Returns:
            A github integration object linked to the repo and branch.

        """
        return Github(client=self._api_client, repo_url=repo, branch=branch)

    def create_param(self, name: str, default: Any, description: str = "") -> ParamArtifact:
        """Creates a parameter artifact that can be fed into other operators.

        Parameter values are configurable at runtime.

        Args:
            name:
                The name to assign this parameter.
            default:
                The default value to give this parameter, if no value is provided.
                Every parameter must have a default.
            description:
                A description of what this parameter represents.

        Returns:
            A parameter artifact.
        """
        if default is None:
            raise InvalidUserArgumentException("Parameter default value cannot be None.")

        # Check that the supplied value is JSON-able.
        try:
            serialized_default = str(json.dumps(default))
        except Exception as e:
            raise InvalidUserArgumentException(
                "Provided parameter must be able to be converted into a JSON object: %s" % str(e)
            )

        operator_id = generate_uuid()
        output_artifact_id = generate_uuid()
        apply_deltas_to_dag(
            self._dag,
            deltas=[
                AddOrReplaceOperatorDelta(
                    op=Operator(
                        id=operator_id,
                        name=name,
                        description=description,
                        spec=OperatorSpec(param=ParamSpec(val=serialized_default)),
                        inputs=[],
                        outputs=[output_artifact_id],
                    ),
                    output_artifacts=[
                        Artifact(
                            id=output_artifact_id,
                            name=artifact_name_from_op_name(name),
                            spec=ArtifactSpec(jsonable={}),
                        ),
                    ],
                )
            ],
        )
        return ParamArtifact(
            self._api_client,
            self._dag,
            output_artifact_id,
        )

    def list_integrations(self) -> Dict[str, IntegrationInfo]:
        """Retrieves a dictionary of integrations the client can use.

        Returns:
            A dictionary mapping from integration name to additional info.
        """
        self._connected_integrations = self._api_client.list_integrations()
        return self._connected_integrations

    def integration(
        self, name: str
    ) -> Union[
        SalesforceIntegration,
        S3Integration,
        GoogleSheetsIntegration,
        RelationalDBIntegration,
    ]:
        """Retrieves a connected integration object.

        Args:
            name:
                The name of the integration

        Returns:
            The integration object with the given name.

        Raises:
            InvalidIntegrationException:
                An error occurred because the cluster is not connected to the
                provided integration or the provided integration is of an
                incompatible type.
        """
        if name not in self._connected_integrations.keys():
            raise InvalidIntegrationException("Not connected to integration %s!" % name)

        integration_info = self._connected_integrations[name]
        if integration_info.service in RelationalDBServices:
            return RelationalDBIntegration(
                api_client=self._api_client,
                dag=self._dag,
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.SALESFORCE:
            return SalesforceIntegration(
                api_client=self._api_client,
                dag=self._dag,
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.GOOGLE_SHEETS:
            return GoogleSheetsIntegration(
                api_client=self._api_client,
                dag=self._dag,
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.S3:
            return S3Integration(
                api_client=self._api_client,
                dag=self._dag,
                metadata=integration_info,
            )
        else:
            raise InvalidIntegrationException(
                "This method does not support loading integration of type %s"
                % integration_info.service
            )

    def publish_flow(
        self,
        name: str,
        description: str = "",
        schedule: str = "",
        k_latest_runs: int = -1,
        artifacts: Optional[List[GenericArtifact]] = None,
    ) -> Flow:
        """Uploads and kicks off the given flow in the system.

        If a flow already exists with the same name, the existing flow will be updated
        to this new state.

        Args:
            name:
                The name of the newly created flow.
            description:
                A description for the new flow.
            schedule: A cron expression specifying the cadence that this flow
                will run on. If empty, the flow will only execute manually.
                For example, to run at the top of every hour:

                >> schedule = aqueduct.hourly(minute: 0)

            k_latest_runs:
                Number of most-recent runs of this flow that Aqueduct should store.
                Runs outside of this bound are deleted. Defaults to persisting all runs.
            artifacts:
                All the artifacts that you care about computing. These artifacts are guaranteed
                to be computed. Additional artifacts may also be included as intermediate
                computation steps. All checks are on the resulting flow are also included.
        Raises:
            InvalidCronStringException:
                An error occurred because the supplied schedule is invalid.
            IncompleteFlowException:
                An error occurred because you are missing some required fields or operators.

        Returns:
            A flow object handle to be used to fetch information about this productionized flow.
        """
        if artifacts is None or len(artifacts) == 0:
            raise IncompleteFlowException(
                "Must supply at least one output artifact when creating a flow."
            )

        cron_schedule = schedule_from_cron_string(schedule)
        retention_policy = retention_policy_from_latest_runs(k_latest_runs)

        dag = apply_deltas_to_dag(
            self._dag,
            deltas=[
                SubgraphDAGDelta(
                    artifact_ids=[artifact.id() for artifact in artifacts],
                    include_load_operators=True,
                    include_check_artifacts=True,
                ),
            ],
            make_copy=True,
        )
        dag.metadata = Metadata(
            name=name,
            description=description,
            schedule=cron_schedule,
            retention_policy=retention_policy,
        )

        if self._in_notebook_or_console_context:
            _show_dag(self._api_client, dag)

        dag.workflow_id = self._api_client.register_workflow(dag).id
        return Flow(
            self._api_client,
            connected_integrations=self._connected_integrations,
            dag=dag,
            in_notebook_or_console_context=self._in_notebook_or_console_context,
        )

    def trigger(self, flow_id: Union[str, uuid.UUID]) -> None:
        """Immediately triggers another run of the provided flow.

        Args:
            flow_id:
                The id of the workflow to delete (not the name)

        Raises:
            InvalidRequestError:
                An error occurred when attempting to fetch the workflow to
                delete. The provided `flow_id` may be malformed.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        if not isinstance(flow_id, str) and not isinstance(flow_id, uuid.UUID):
            raise InvalidUserArgumentException("Provided flow id must be either str or uuid.")

        if isinstance(flow_id, uuid.UUID):
            flow_id = str(flow_id)
        self._api_client.refresh_workflow(flow_id)

    def delete_flow(self, flow_id: Union[str, uuid.UUID]) -> None:
        """Deletes a flow object.

        Args:
            flow_id:
                The id of the workflow to delete (not the name)

        Raises:
            InvalidRequestError:
                An error occurred when attempting to fetch the workflow to
                delete. The provided `flow_id` may be malformed.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        if not isinstance(flow_id, str) and not isinstance(flow_id, uuid.UUID):
            raise InvalidUserArgumentException("Provided flow id must be either str or uuid.")

        if isinstance(flow_id, uuid.UUID):
            flow_id = str(flow_id)

        # TODO(ENG-410): This method gives no indication as to whether the flow
        #  was successfully deleted.
        self._api_client.delete_workflow(flow_id)

    def show_dag(self, artifacts: Optional[List[GenericArtifact]] = None) -> None:
        """Prints out the flow as a pyplot graph.

        A user outside the notebook environment will be redirected to a page in their browser
        containing the graph.

        Args:
            artifacts:
                If specified, the subgraph terminating at these artifacts will be specified.
                Otherwise, the entire graph is printed.
        """
        dag = self._dag
        if artifacts is not None:
            dag = apply_deltas_to_dag(
                self._dag,
                deltas=[
                    SubgraphDAGDelta(
                        artifact_ids=[artifact.id() for artifact in artifacts],
                        include_load_operators=True,
                        include_check_artifacts=True,
                    ),
                ],
                make_copy=True,
            )
        _show_dag(self._api_client, dag)

    def describe(self) -> None:
        """Prints out info about this client in a human-readable format."""
        print("============================= Aqueduct Client =============================")
        print("Connected endpoint: %s" % self._api_client.aqueduct_address)
        print("Log Level: %s" % logging.getLevelName(logging.root.level))
        self._connected_integrations = self._api_client.list_integrations()
        print("Current Integrations:")
        for integrations in self._connected_integrations:
            print("\t -" + integrations)

    def _get_flow_info(self, flow_id: str) -> Any:
        """WARNING: this is only meant for our SDK integration tests to use. We do not publicly
        support fetching an existing flow through the SDK yet.
        """
        return self._api_client.get_workflow(flow_id)
