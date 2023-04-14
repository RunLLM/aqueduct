import logging
import os
import platform
import uuid
import warnings
from collections import defaultdict
from typing import Any, DefaultDict, Dict, List, Optional, Union

import __main__ as main
import yaml
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.bool_artifact import BoolArtifact
from aqueduct.artifacts.create import create_param_artifact
from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.constants.enums import (
    ArtifactType,
    ExecutionStatus,
    RelationalDBServices,
    RuntimeType,
    ServiceType,
)
from aqueduct.error import (
    InvalidIntegrationException,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from aqueduct.flow import Flow
from aqueduct.github import Github
from aqueduct.integrations.airflow_integration import AirflowIntegration
from aqueduct.integrations.aws_integration import AWSIntegration
from aqueduct.integrations.connect_config import (
    BaseConnectionConfig,
    IntegrationConfig,
    convert_dict_to_integration_connect_config,
    prepare_integration_config,
)
from aqueduct.integrations.databricks_integration import DatabricksIntegration
from aqueduct.integrations.google_sheets_integration import GoogleSheetsIntegration
from aqueduct.integrations.k8s_integration import K8sIntegration
from aqueduct.integrations.lambda_integration import LambdaIntegration
from aqueduct.integrations.mongodb_integration import MongoDBIntegration
from aqueduct.integrations.s3_integration import S3Integration
from aqueduct.integrations.salesforce_integration import SalesforceIntegration
from aqueduct.integrations.spark_integration import SparkIntegration
from aqueduct.integrations.sql_integration import RelationalDBIntegration
from aqueduct.logger import logger
from aqueduct.models.dag import Metadata, RetentionPolicy
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.operators import ParamSpec
from aqueduct.models.response_models import SavedObjectUpdate
from aqueduct.utils.dag_deltas import (
    SubgraphDAGDelta,
    apply_deltas_to_dag,
    validate_overwriting_parameters,
)
from aqueduct.utils.local_data import validate_local_data
from aqueduct.utils.serialization import deserialize, extract_val_from_local_data
from aqueduct.utils.type_inference import _base64_string_to_bytes, infer_artifact_type
from aqueduct.utils.utils import (
    construct_param_spec,
    find_flow_with_user_supplied_id_and_name,
    generate_engine_config,
    generate_flow_schedule,
    generate_ui_url,
)

from aqueduct import globals


def global_config(config_dict: Dict[str, Any]) -> None:
    """Sets any global configuration variables in the current Aqueduct context.

    Args:
        config_dict:
            A dict from the configuration key to its new value.

    Available configuration keys:
        "lazy":
            A boolean indicating whether any new functions will be constructed lazily (True) or eagerly (False).
        "engine":
            The name of the default compute integration to run all functions against.
            This can still be overriden by the `engine` argument in `client.publish_flow()` or
            on the @op spec. To set this to run against the Aqueduct engine, use "aqueduct" (case-insensitive).
    """

    if globals.GLOBAL_LAZY_KEY in config_dict:
        lazy_val = config_dict[globals.GLOBAL_LAZY_KEY]
        if not isinstance(lazy_val, bool):
            raise InvalidUserArgumentException("Must supply a boolean for the lazy key.")
        globals.__GLOBAL_CONFIG__.lazy = lazy_val

    if globals.GLOBAL_ENGINE_KEY in config_dict:
        engine_name = config_dict[globals.GLOBAL_ENGINE_KEY]
        if not isinstance(engine_name, str):
            raise InvalidUserArgumentException(
                "Engine should be the string name of your compute integration."
            )
        globals.__GLOBAL_CONFIG__.engine = engine_name


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
        # We must call basicConfig() here so messages show up in Jupyter notebooks.
        logging.basicConfig(format="%(levelname)s:%(message)s", level=logging_level)

        if api_key == "":
            api_key = get_apikey()

        globals.__GLOBAL_API_CLIENT__.configure(api_key, aqueduct_address)
        self._connected_integrations: Dict[
            str, IntegrationInfo
        ] = globals.__GLOBAL_API_CLIENT__.list_integrations()
        self._dag = globals.__GLOBAL_DAG__

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
        return Github(repo_url=repo, branch=branch)

    def create_param(
        self,
        name: str,
        default: Any,
        description: str = "",
        use_local: bool = False,
        as_type: Optional[ArtifactType] = None,
        format: Optional[str] = None,
    ) -> BaseArtifact:
        """Creates a parameter artifact that can be fed into other operators.

        Parameter values are configurable at runtime.

        Args:
            name:
                The name to assign this parameter.
            default:
                The default value to give this parameter, if no value is provided.
                Every parameter must have a default. If we decide to use local data,
                a path to the local data file must be specified.
            description:
                A description of what this parameter represents.
            use_local:
                Whether this parameter uses local data source or not.
            as_type:
                The expected type of the local data. Only supported types are ArtifactType.TABLE and ArtifactType.IMAGE.
            format:
                If local data type is ArtifactType.TABLE, the user has to specify the table format.
                We currently support "json", "csv", and "parquet".
        Returns:
            A parameter artifact.
        """
        if use_local:
            if not isinstance(default, str):
                raise InvalidUserArgumentException(
                    "The default value must be a path to local data."
                )
            validate_local_data(default, as_type, format)
            default = extract_val_from_local_data(default, as_type, format)
        return create_param_artifact(
            self._dag,
            name,
            default,
            description,
            explicitly_named=True,
            is_local_data=use_local,
        )

    def connect_integration(
        self,
        name: str,
        service: Union[str, ServiceType],
        config: Union[Dict[str, str], IntegrationConfig],
    ) -> None:
        """Connects the Aqueduct server to an integration.

        Args:
            name:
                The name to assign this integration. Will error if an integration with that name
                already exists.
            service:
                The type of integration to connect to.
            config:
                Either a dictionary or an IntegrationConnectConfig object that contains the
                configuration credentials needed to connect.
        """
        if service not in ServiceType:
            raise InvalidUserArgumentException(
                "Service argument must match exactly one of the enum values in ServiceType (case-sensitive)."
            )

        self._connected_integrations = globals.__GLOBAL_API_CLIENT__.list_integrations()
        if name in self._connected_integrations.keys():
            raise InvalidUserActionException(
                "Cannot connect a new integration with name `%s`. An integration with this name already exists."
                % name
            )

        if not isinstance(config, dict) and not isinstance(config, BaseConnectionConfig):
            raise InvalidUserArgumentException(
                "`config` argument must be either a dict or IntegrationConnectConfig."
            )

        if isinstance(config, dict):
            config = convert_dict_to_integration_connect_config(service, config)
        assert isinstance(config, BaseConnectionConfig)

        config = prepare_integration_config(service, config)

        globals.__GLOBAL_API_CLIENT__.connect_integration(name, service, config)
        logger().info("Successfully connected to new %s integration `%s`." % (service, name))

    def delete_integration(
        self,
        name: str,
    ) -> None:
        """Deletes the integration from Aqueduct.

        Args:
            name:
                The name of the integration to delete.
        """
        existing_integrations = globals.__GLOBAL_API_CLIENT__.list_integrations()
        if name not in existing_integrations.keys():
            raise InvalidIntegrationException("Not connected to integration %s!" % name)

        globals.__GLOBAL_API_CLIENT__.delete_integration(existing_integrations[name].id)

        # Update the connected integrations cached on this object.
        self._connected_integrations = globals.__GLOBAL_API_CLIENT__.list_integrations()

    def list_integrations(self) -> Dict[str, IntegrationInfo]:
        """Retrieves a dictionary of integrations the client can use.

        Returns:
            A dictionary mapping from integration name to additional info.
        """
        self._connected_integrations = globals.__GLOBAL_API_CLIENT__.list_integrations()
        return self._connected_integrations

    def integration(
        self, name: str
    ) -> Union[
        SalesforceIntegration,
        S3Integration,
        GoogleSheetsIntegration,
        RelationalDBIntegration,
        AirflowIntegration,
        K8sIntegration,
        LambdaIntegration,
        MongoDBIntegration,
        DatabricksIntegration,
        SparkIntegration,
        AWSIntegration,
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
        self._connected_integrations = globals.__GLOBAL_API_CLIENT__.list_integrations()

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
        elif integration_info.service == ServiceType.K8S:
            return K8sIntegration(
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.LAMBDA:
            return LambdaIntegration(
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.MONGO_DB:
            return MongoDBIntegration(
                dag=self._dag,
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.DATABRICKS:
            return DatabricksIntegration(
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.SPARK:
            return SparkIntegration(
                metadata=integration_info,
            )
        elif integration_info.service == ServiceType.AWS:
            dynamic_k8s_integration_name = "%s:aqueduct_ondemand_k8s" % name
            dynamic_k8s_integration_info = self._connected_integrations[
                dynamic_k8s_integration_name
            ]
            return AWSIntegration(
                metadata=integration_info,
                k8s_integration_metadata=dynamic_k8s_integration_info,
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
            for workflow_resp in globals.__GLOBAL_API_CLIENT__.list_workflows()
        ]

    def flow(
        self,
        flow_id: Optional[Union[str, uuid.UUID]] = None,
        flow_name: Optional[str] = None,
    ) -> Flow:
        """Fetches a flow corresponding to the given flow id.

        Args:
            flow_id:
                Used to identify the flow to fetch from the system.
                Between `flow_id` and `flow_name`, at least one must be provided.
                If both are specified, they must correspond to the same flow.
            flow_name:
                Used to identify the flow to fetch from the system.

        Raises:
            InvalidUserArgumentException:
                If the provided flow id or name does not correspond to a flow the client knows about.
        """
        flows = [(flow.id, flow.name) for flow in globals.__GLOBAL_API_CLIENT__.list_workflows()]
        flow_id = find_flow_with_user_supplied_id_and_name(
            flows,
            flow_id,
            flow_name,
        )

        return Flow(
            flow_id,
            self._in_notebook_or_console_context,
        )

    def publish_flow(
        self,
        name: str,
        description: str = "",
        schedule: str = "",
        engine: Optional[str] = None,
        artifacts: Optional[Union[BaseArtifact, List[BaseArtifact]]] = None,
        metrics: Optional[List[NumericArtifact]] = None,
        checks: Optional[List[BoolArtifact]] = None,
        k_latest_runs: Optional[int] = None,
        source_flow: Optional[Union[Flow, str, uuid.UUID]] = None,
        run_now: Optional[bool] = None,
        use_local: Optional[bool] = False,
    ) -> Flow:
        """Uploads and kicks off the given flow in the system.

        If a flow already exists with the same name, the existing flow will be updated
        to this new state.

        The default execution engine of the flow is Aqueduct. In order to specify which
        execution engine the flow will be running on, use "config" parameter. Eg:
        >>> flow = client.publish_flow(
        >>>     name="k8s_example",
        >>>     artifacts=[output],
        >>>     engine="k8s_integration",
        >>> )

        Args:
            name:
                The name of the newly created flow.
            description:
                A description for the new flow.
            schedule:
                A cron expression specifying the cadence that this flow
                will run on. If empty, the flow will only execute manually.
                For example, to run at the top of every hour:

                >> schedule = aqueduct.hourly(minute: 0)
            engine:
                The name of the compute integration (eg. "my_lambda_integration") this the flow will
                be computed on.
            artifacts:
                All the artifacts that you care about computing. These artifacts are guaranteed
                to be computed. Additional artifacts may also be computed if they are upstream
                dependencies.
            metrics:
                All the metrics that you would like to compute. If not supplied, we include by default
                all metrics computed on artifacts in the flow.
            checks:
                All the checks that you would like to compute. If not supplied, we will by default
                all checks computed on artifacts in the flow.
            k_latest_runs:
                Number of most-recent runs of this flow that Aqueduct should keep. Runs outside of
                this bound are garbage collected. Defaults to persisting all runs.
            source_flow:
                Used to identify the source flow for this flow. This can be identified
                via an object (Flow), name (str), or id (str or uuid).
            run_now:
                Used to specify if the flow should run immediately at publish time. The default
                behavior is 'True'.
            use_local:
                Must be set if any artifact in the flow is derived from local data.

        Raises:
            InvalidUserArgumentException:
                An invalid combination of parameters was provided.
            InvalidCronStringException:
                An error occurred because the supplied schedule is invalid.
            IncompleteFlowException:
                An error occurred because you are missing some required fields or operators.

        Returns:
            A flow object handle to be used to fetch information about this productionized flow.
        """
        if not isinstance(name, str) or name == "":
            raise InvalidUserArgumentException(
                "A non-empty string must be supplied for the flow's name."
            )

        if engine is not None and not isinstance(engine, str):
            raise InvalidUserArgumentException(
                "`engine` parameter must be a string, got %s." % type(engine)
            )

        if artifacts is None or artifacts == []:
            raise InvalidUserArgumentException(
                "Must supply at least one artifact to compute when creating a flow."
            )

        if isinstance(artifacts, BaseArtifact):
            artifacts = [artifacts]

        if not isinstance(artifacts, list) or any(
            not isinstance(artifact, BaseArtifact) for artifact in artifacts
        ):
            raise InvalidUserArgumentException(
                "`artifacts` argument must either be an artifact or a list of artifacts."
            )

        if source_flow and schedule != "":
            raise InvalidUserArgumentException(
                "Cannot create a flow with both a schedule and a source flow, pick one."
            )

        if (
            source_flow
            and not isinstance(source_flow, Flow)
            and not isinstance(source_flow, str)
            and not isinstance(source_flow, uuid.UUID)
        ):
            raise InvalidUserArgumentException(
                "`source_flow` argument must either be a flow, str, or uuid."
            )

        source_flow_id = None
        if isinstance(source_flow, Flow):
            source_flow_id = source_flow.id()
        elif isinstance(source_flow, str):
            # Check if there is a flow with the name `source_flow`
            for workflow in globals.__GLOBAL_API_CLIENT__.list_workflows():
                if workflow.name == source_flow:
                    source_flow_id = workflow.id
                    break

            if not source_flow_id:
                # No flow with name `source_flow` was found so try to convert
                # the str to a uuid
                source_flow_id = uuid.UUID(source_flow)
        elif isinstance(source_flow, uuid.UUID):
            source_flow_id = source_flow

        # If metrics and/or checks are explicitly included, add them to the artifacts list,
        # but don't include them implicitly.
        implicitly_include_metrics = True
        if metrics is not None:
            if not isinstance(metrics, list):
                raise InvalidUserArgumentException("`metrics` argument must be a list.")
            artifacts += metrics
            implicitly_include_metrics = False

        implicitly_include_checks = True
        if checks is not None:
            if not isinstance(checks, list):
                raise InvalidUserArgumentException("`checks` argument must be a list.")
            artifacts += checks
            implicitly_include_checks = False

        flow_schedule = generate_flow_schedule(schedule, source_flow_id)

        if k_latest_runs is None:
            retention_policy = RetentionPolicy(k_latest_runs=-1)
        else:
            if not isinstance(k_latest_runs, int):
                raise InvalidUserArgumentException(
                    "`k_latest_runs` parameter must be an int, got %s" % type(k_latest_runs)
                )
            retention_policy = RetentionPolicy(k_latest_runs=k_latest_runs)

        dag = apply_deltas_to_dag(
            self._dag,
            deltas=[
                SubgraphDAGDelta(
                    artifact_ids=[artifact.id() for artifact in artifacts],
                    include_saves=True,
                    include_metrics=implicitly_include_metrics,
                    include_checks=implicitly_include_checks,
                ),
            ],
            make_copy=True,
        )
        if not use_local and any(
            artifact_metadata.from_local_data for artifact_metadata in list(dag.artifacts.values())
        ):
            raise InvalidUserActionException(
                "Cannot create a flow with local data. Consider setting `use_local` to True to publish a workflow with local data parameters."
            )
        dag.metadata = Metadata(
            name=name,
            description=description,
            schedule=flow_schedule,
            retention_policy=retention_policy,
        )
        dag.set_engine_config(
            global_engine_config=generate_engine_config(
                self._connected_integrations,
                globals.__GLOBAL_CONFIG__.engine,
            ),
            publish_flow_engine_config=generate_engine_config(self._connected_integrations, engine),
        )

        dag.validate_and_resolve_artifact_names()

        if dag.engine_config.type == RuntimeType.AIRFLOW:
            if run_now is not None:
                raise InvalidUserArgumentException(
                    "run_now parameter is not supported for Airflow engine."
                )
            # This is an Airflow workflow
            resp = globals.__GLOBAL_API_CLIENT__.register_airflow_workflow(dag)
            flow_id, airflow_file = resp.id, resp.file

            file = "{}_airflow.py".format(name)
            with open(file, "w") as f:
                f.write(airflow_file)

            if resp.is_update:
                print(
                    """The updated Airflow DAG file has been downloaded to: {}.
                    Please copy it to your Airflow server to begin execution.
                    New Airflow DAG runs will not be synced properly with Aqueduct
                    until you have copied the file.""".format(
                        file
                    )
                )
            else:
                print(
                    """The Airflow DAG file has been downloaded to: {}.
                    Please copy it to your Airflow server to begin execution.""".format(
                        file
                    )
                )
        else:
            if run_now is None:
                run_now = True
            registered_metadata = globals.__GLOBAL_API_CLIENT__.register_workflow(dag, run_now)
            flow_id = registered_metadata.id
            server_python_version = (
                registered_metadata.python_version.strip()
            )  # Remove newline at the end
            client_python_version = f"Python {platform.python_version()}"
            if (
                dag.engine_config.type == RuntimeType.AQUEDUCT
                and client_python_version != server_python_version
            ):
                warnings.warn(
                    "There is a mismatch between the Python version on the engine (%s) and the client (%s)."
                    % (server_python_version, client_python_version)
                )

        url = generate_ui_url(
            globals.__GLOBAL_API_CLIENT__.construct_base_url(),
            str(flow_id),
        )
        print("Url: ", url)

        return Flow(
            str(flow_id),
            self._in_notebook_or_console_context,
        )

    def trigger(
        self,
        flow_id: Optional[Union[str, uuid.UUID]] = None,
        flow_name: Optional[str] = None,
        parameters: Optional[Dict[str, Any]] = None,
    ) -> None:
        """Immediately triggers another run of the provided flow.

        Args:
            flow_id:
                The id of the flow to delete.
                Between `flow_id` and `flow_name`, at least one must be provided.
                If both are specified, they must correspond to the same flow.
            flow_name:
                The name of the flow to delete.
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
        param_specs: Dict[str, ParamSpec] = {}
        if parameters is not None:
            flow = self.flow(flow_id)
            latest_run = flow.latest()

            # NOTE: this is a defense check against triggering runs that haven't run yet.
            # We may want to revisit this in the future if more nuanced constraints are necessary.
            if not latest_run:
                raise InvalidUserActionException(
                    "Cannot trigger a workflow that hasn't already run at least once."
                )
            validate_overwriting_parameters(latest_run._dag, parameters)

            for name, new_val in parameters.items():
                artifact_type = infer_artifact_type(new_val)
                param_specs[name] = construct_param_spec(new_val, artifact_type)

        flows = [(flow.id, flow.name) for flow in globals.__GLOBAL_API_CLIENT__.list_workflows()]
        flow_id = find_flow_with_user_supplied_id_and_name(
            flows,
            flow_id,
            flow_name,
        )
        globals.__GLOBAL_API_CLIENT__.refresh_workflow(flow_id, param_specs)

    def delete_flow(
        self,
        flow_id: Optional[Union[str, uuid.UUID]] = None,
        flow_name: Optional[str] = None,
        saved_objects_to_delete: Optional[
            DefaultDict[Union[str, Integration], List[SavedObjectUpdate]]
        ] = None,
        force: bool = False,
    ) -> None:
        """Deletes a flow object.

        Args:
            flow_id:
                The id of the flow to delete.
                Between `flow_id` and `flow_name`, at least one must be provided.
                If both are specified, they must correspond to the same flow.
            flow_name:
                The name of the flow to delete.
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

        flows = [(flow.id, flow.name) for flow in globals.__GLOBAL_API_CLIENT__.list_workflows()]
        flow_id = find_flow_with_user_supplied_id_and_name(
            flows,
            flow_id,
            flow_name,
        )

        # TODO(ENG-410): This method gives no indication as to whether the flow
        #  was successfully deleted.
        resp = globals.__GLOBAL_API_CLIENT__.delete_workflow(
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
        print("Connected endpoint: %s" % globals.__GLOBAL_API_CLIENT__.aqueduct_address)
        print("Log Level: %s" % logging.getLevelName(logging.root.level))
        self._connected_integrations = globals.__GLOBAL_API_CLIENT__.list_integrations()
        print("Current Integrations:")
        for integrations in self._connected_integrations:
            print("\t -" + integrations)
