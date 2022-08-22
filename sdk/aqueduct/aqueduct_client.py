import json
import logging
import os
import uuid
import warnings
from collections import defaultdict
from typing import Any, DefaultDict, Dict, List, Optional, Union

import __main__ as main
import yaml
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.artifacts.param_artifact import ParamArtifact

from aqueduct import api_client, dag

from .dag import (
    DAG,
    AddOrReplaceOperatorDelta,
    AirflowEngineConfig,
    EngineConfig,
    Metadata,
    SubgraphDAGDelta,
    apply_deltas_to_dag,
    validate_overwriting_parameters,
)
from .enums import ArtifactType, ExecutionStatus, RelationalDBServices, RuntimeType, ServiceType
from .error import (
    IncompleteFlowException,
    InvalidIntegrationException,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from .flow import Flow, FlowConfig
from .github import Github
from .integrations.airflow_integration import AirflowIntegration
from .integrations.google_sheets_integration import GoogleSheetsIntegration
from .integrations.integration import Integration, IntegrationInfo
from .integrations.s3_integration import S3Integration
from .integrations.salesforce_integration import SalesforceIntegration
from .integrations.sql_integration import RelationalDBIntegration
from .logger import logger
from .operators import Operator, OperatorSpec, ParamSpec, serialize_parameter_value
from .responses import Error, SavedObjectDelete, SavedObjectUpdate
from .utils import (
    _infer_requirements,
    generate_ui_url,
    generate_uuid,
    infer_artifact_type,
    parse_user_supplied_id,
    retention_policy_from_latest_runs,
    schedule_from_cron_string,
)


def get_apikey() -> str:
    """
    Get the API key if the server is running locally.

    Returns:
        The API key.
    """
    server_directory = os.path.join(os.environ["HOME"], ".aqueduct", "server")
    config_file = os.path.join(server_directory, "config", "config.yml")
    with open(config_file, "r") as f:
        try:
            return str(yaml.safe_load(f)["apiKey"])
        except yaml.YAMLError as exc:
            logger().error(
                "This API works only when you are running the server and the SDK on the same machine."
            )
            exit(1)


def infer_requirements() -> List[str]:
    """Obtains the list of pip requirements specifiers from the current python environment using `pip freeze`.

    Returns:
        A list, for example, ["transformers==4.21.0", "numpy==1.22.4"].
    """
    return _infer_requirements()


class Client:
    """This class allows users to interact with flows on their Aqueduct cluster."""

    def __init__(
        self,
        api_key: str = "",
        aqueduct_address: str = "http://localhost:8080",
        logging_level: int = logging.WARNING,
    ):
        """Creates an instance of Client.

        Args:
            api_key:
                The user unique API key provided by Aqueduct. If no key is
                provided, the client attempts to read the key stored on the
                local server and errors if non exists.
            aqueduct_address:
                The address of the Aqueduct Server service. If no address is
                provided, the client attempts to connect to
                http://localhost:8080.
            logging_level:
                A indication of what level and above to print logs from the sdk.
                Defaults to printing warning and above only. Types defined in: https://docs.python.org/3/howto/logging.html

        Returns:
            A Client instance.
        """
        logger().setLevel(level=logging_level)

        if api_key == "":
            api_key = get_apikey()

        api_client.__GLOBAL_API_CLIENT__.configure(api_key, aqueduct_address)
        self._connected_integrations: Dict[
            str, IntegrationInfo
        ] = api_client.__GLOBAL_API_CLIENT__.list_integrations()
        self._dag = dag.__GLOBAL_DAG__

        # Will show graph if in an ipynb or Python console, but not if running a Python script.
        self._in_notebook_or_console_context = (not hasattr(main, "__file__")) and (
            not "PYTEST_CURRENT_TEST" in os.environ
        )

        # Check if "@ file" in pip freeze requirements and warn user.
        if not "localhost" in aqueduct_address:
            skipped_packages = []
            for requirement in infer_requirements():
                if "@ file" in requirement:
                    skipped_packages.append(requirement.split(" ")[0])
            if len(skipped_packages) > 0:
                warnings.warn(
                    "Your local Python environment contains packages installed from the local file system. The following packages won't be installed when running your workflow: "
                    + ", ".join(skipped_packages)
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
        return Github(repo_url=repo, branch=branch)

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

        artifact_type = infer_artifact_type(default)

        val = serialize_parameter_value(name, default)

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
                        spec=OperatorSpec(param=ParamSpec(val=val)),
                        inputs=[],
                        outputs=[output_artifact_id],
                    ),
                    output_artifacts=[
                        ArtifactMetadata(
                            id=output_artifact_id,
                            name=name,
                            type=artifact_type,
                        ),
                    ],
                )
            ],
        )
        return ParamArtifact(
            self._dag,
            output_artifact_id,
        )

    def list_integrations(self) -> Dict[str, IntegrationInfo]:
        """Retrieves a dictionary of integrations the client can use.

        Returns:
            A dictionary mapping from integration name to additional info.
        """
        self._connected_integrations = api_client.__GLOBAL_API_CLIENT__.list_integrations()
        return self._connected_integrations

    def integration(
        self, name: str
    ) -> Union[
        SalesforceIntegration,
        S3Integration,
        GoogleSheetsIntegration,
        RelationalDBIntegration,
        AirflowIntegration,
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
        self._connected_integrations = api_client.__GLOBAL_API_CLIENT__.list_integrations()

        if name not in self._connected_integrations.keys():
            raise InvalidIntegrationException("Not connected to integration %s!" % name)

        integration_info = self._connected_integrations[name]
        if integration_info.service in RelationalDBServices:
            return RelationalDBIntegration(
                dag=self._dag,
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.SALESFORCE:
            return SalesforceIntegration(
                dag=self._dag,
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.GOOGLE_SHEETS:
            return GoogleSheetsIntegration(
                dag=self._dag,
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.S3:
            return S3Integration(
                dag=self._dag,
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.AIRFLOW:
            return AirflowIntegration(
                metadata=integration_info,
            )
        else:
            raise InvalidIntegrationException(
                "This method does not support loading integration of type %s"
                % integration_info.service
            )

    def list_flows(self) -> List[Dict[str, str]]:
        """Lists the flows that are accessible by this client.

        Returns:
            A list of flows, each represented as a dictionary providing essential
            information (eg. name, id, etc.), which the user can use to inspect
            the flow further in the UI or SDK.
        """
        return [
            workflow_resp.to_readable_dict()
            for workflow_resp in api_client.__GLOBAL_API_CLIENT__.list_workflows()
        ]

    def flow(self, flow_id: Union[str, uuid.UUID]) -> Flow:
        """Fetches a flow corresponding to the given flow id.

        Args:
             flow_id:
                Used to identify the flow to fetch from the system.

        Raises:
            InvalidUserArgumentException:
                If the provided flow id does not correspond to a flow the client knows about.
        """
        flow_id = parse_user_supplied_id(flow_id)

        if all(
            uuid.UUID(flow_id) != workflow.id
            for workflow in api_client.__GLOBAL_API_CLIENT__.list_workflows()
        ):
            raise InvalidUserArgumentException("Unable to find a flow with id %s" % flow_id)

        return Flow(
            flow_id,
            self._in_notebook_or_console_context,
        )

    def publish_flow(
        self,
        name: str,
        description: str = "",
        schedule: str = "",
        k_latest_runs: int = -1,
        artifacts: Optional[List[BaseArtifact]] = None,
        config: Optional[FlowConfig] = None,
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
            config:
                An optional set of config fields for this flow.
                - engine: Specify where this flow should run with one of your connected integrations.
                We currently support Airflow.

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

        if config and config.engine:
            # This is an Airflow workflow
            dag.engine_config = EngineConfig(
                type=RuntimeType.AIRFLOW,
                airflow_config=AirflowEngineConfig(
                    integration_id=config.engine._metadata.id,
                ),
            )
            resp = api_client.__GLOBAL_API_CLIENT__.register_airflow_workflow(dag)
            flow_id, airflow_file = resp.id, resp.file

            file = "{}_airflow.py".format(name)
            with open(file, "w") as f:
                f.write(airflow_file)
            print(
                """The Airflow DAG file has been downloaded to: {}. 
                Please copy it to your Airflow server to begin execution.""".format(
                    file
                )
            )
        else:
            flow_id = api_client.__GLOBAL_API_CLIENT__.register_workflow(dag).id

        url = generate_ui_url(
            api_client.__GLOBAL_API_CLIENT__.construct_base_url(),
            str(flow_id),
        )
        print("Url: ", url)

        return Flow(
            str(flow_id),
            self._in_notebook_or_console_context,
        )

    def trigger(
        self,
        flow_id: Union[str, uuid.UUID],
        parameters: Optional[Dict[str, Any]] = None,
    ) -> None:
        """Immediately triggers another run of the provided flow.

        Args:
            flow_id:
                The id of the workflow to delete (not the name)
            parameters:
                A map containing custom values to use for the designated parameters. The mapping
                is expected to be from parameter name to the custom value. These custom values
                are not persisted to the workflow. To actually change the default parameter values
                edit the workflow itself through `client.publish_flow()`.

        Raises:
            InvalidRequestError:
                An error occurred when attempting to fetch the workflow to
                delete. The provided `flow_id` may be malformed.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        if parameters is not None:
            flow = self.flow(flow_id)
            runs = flow.list_runs(limit=1)

            # NOTE: this is a defense check against triggering runs that haven't run yet.
            # We may want to revisit this in the future if more nuanced constraints are necessary.
            if len(runs) == 0:
                raise InvalidUserActionException(
                    "Cannot trigger a workflow that hasn't already run at least once."
                )
            validate_overwriting_parameters(flow.latest()._dag, parameters)

        serialized_params = None
        if parameters is not None:
            if any(not isinstance(name, str) for name in parameters):
                raise InvalidUserArgumentException("Parameters must be keyed by strings.")

            serialized_params = json.dumps(
                {name: serialize_parameter_value(name, val) for name, val in parameters.items()}
            )

        flow_id = parse_user_supplied_id(flow_id)
        api_client.__GLOBAL_API_CLIENT__.refresh_workflow(flow_id, serialized_params)

    def delete_flow(
        self,
        flow_id: Union[str, uuid.UUID],
        saved_objects_to_delete: Optional[
            DefaultDict[Union[str, Integration], List[SavedObjectUpdate]]
        ] = None,
        force: bool = False,
    ) -> None:
        """Deletes a flow object.

        Args:
            flow_id:
                The id of the workflow to delete (not the name)
            saved_objects_to_delete:
                The tables or storage paths to delete grouped by integration name.
            force:
                Force the deletion even though some workflow-written objects in the writes_to_delete argument had UpdateMode=append

        Raises:
            InvalidRequestError:
                An error occurred when attempting to fetch the workflow to
                delete. The provided `flow_id` may be malformed.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        if saved_objects_to_delete is None:
            saved_objects_to_delete = defaultdict()
        flow_id = parse_user_supplied_id(flow_id)

        # TODO(ENG-410): This method gives no indication as to whether the flow
        #  was successfully deleted.
        resp = api_client.__GLOBAL_API_CLIENT__.delete_workflow(
            flow_id, saved_objects_to_delete, force
        )

        failures = []
        for integration in resp.saved_object_deletion_results:
            for obj in resp.saved_object_deletion_results[integration]:
                if obj.exec_state.status == ExecutionStatus.FAILED:
                    trace = ""
                    if obj.exec_state.error:
                        context = obj.exec_state.error.context.strip().replace("\n", "\n>\t")
                        trace = f">\t{context}\n{obj.exec_state.error.tip}"
                    failure_string = f"[{integration}] {obj.name}\n{trace}"
                    failures.append(failure_string)
        if len(failures) > 0:
            failures_string = "\n".join(failures)
            raise Exception(
                f"Failed to delete {len(failures)} saved objects.\nFailures\n{failures_string}"
            )

    def describe(self) -> None:
        """Prints out info about this client in a human-readable format."""
        print("============================= Aqueduct Client =============================")
        print("Connected endpoint: %s" % api_client.__GLOBAL_API_CLIENT__.aqueduct_address)
        print("Log Level: %s" % logging.getLevelName(logging.root.level))
        self._connected_integrations = api_client.__GLOBAL_API_CLIENT__.list_integrations()
        print("Current Integrations:")
        for integrations in self._connected_integrations:
            print("\t -" + integrations)
