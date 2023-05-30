import logging
import os
import platform
import re
import uuid
import warnings
from collections import defaultdict
from typing import Any, DefaultDict, Dict, List, Optional, Tuple, Union

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
    InvalidResourceException,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from aqueduct.flow import Flow
from aqueduct.github import Github
from aqueduct.logger import logger
from aqueduct.models.dag import Metadata, RetentionPolicy
from aqueduct.models.operators import ParamSpec, S3LoadParams
from aqueduct.models.resource import BaseResource, ResourceInfo
from aqueduct.models.response_models import SavedObjectUpdate
from aqueduct.resources.airflow import AirflowResource
from aqueduct.resources.aws import AWSResource
from aqueduct.resources.aws_lambda import LambdaResource
from aqueduct.resources.connect_config import (
    BaseConnectionConfig,
    ResourceConfig,
    convert_dict_to_resource_connect_config,
    prepare_resource_config,
)
from aqueduct.resources.databricks import DatabricksResource
from aqueduct.resources.dynamic_k8s import DynamicK8sResource
from aqueduct.resources.ecr import ECRResource
from aqueduct.resources.gar import GARResource
from aqueduct.resources.google_sheets import GoogleSheetsResource
from aqueduct.resources.mongodb import MongoDBResource
from aqueduct.resources.parameters import USER_TAG_PATTERN
from aqueduct.resources.s3 import S3Resource
from aqueduct.resources.salesforce import SalesforceResource
from aqueduct.resources.spark import SparkResource
from aqueduct.resources.sql import RelationalDBResource
from aqueduct.utils.dag_deltas import (
    SubgraphDAGDelta,
    apply_deltas_to_dag,
    validate_overwriting_parameters,
)
from aqueduct.utils.local_data import validate_local_data
from aqueduct.utils.serialization import extract_val_from_local_data
from aqueduct.utils.type_inference import infer_artifact_type
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
            The name of the default compute resource to run all functions against.
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
                "Engine should be the string name of your compute resource."
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


DEPRECATED_AQUEDUCT_DEMO_DB_NAME = "aqueduct_demo"
AQUEDUCT_DEMO_DB_NAME = "Demo"


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
        self._connected_resources: Dict[
            str, ResourceInfo
        ] = globals.__GLOBAL_API_CLIENT__.list_resources()
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
            A github resource object linked to the repo and branch.

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
        config: Union[Dict[str, str], ResourceConfig],
    ) -> None:
        """Deprecated. Use `client.connect_resource()` instead."""
        logger().warning(
            "client.connect_resource() will be deprecated soon. Use `client.connect_resource() instead."
        )
        return self.connect_resource(name, service, config)

    def connect_resource(
        self,
        name: str,
        service: Union[str, ServiceType],
        config: Union[Dict[str, str], ResourceConfig],
    ) -> None:
        """Connects the Aqueduct server to an resource.

        Args:
            name:
                The name to assign this resource. Will error if an resource with that name
                already exists.
            service:
                The type of resource to connect to.
            config:
                Either a dictionary or an ResourceConnectConfig object that contains the
                configuration credentials needed to connect.
        """
        if service not in ServiceType:
            raise InvalidUserArgumentException(
                "Service argument must match exactly one of the enum values in ServiceType (case-sensitive)."
            )

        self._connected_resources = globals.__GLOBAL_API_CLIENT__.list_resources()
        if name in self._connected_resources.keys():
            raise InvalidUserActionException(
                "Cannot connect a new resource with name `%s`. An resource with this name already exists."
                % name
            )

        if not isinstance(config, dict) and not isinstance(config, BaseConnectionConfig):
            raise InvalidUserArgumentException(
                "`config` argument must be either a dict or ResourceConnectConfig."
            )

        if isinstance(config, dict):
            config = convert_dict_to_resource_connect_config(service, config)
        assert isinstance(config, BaseConnectionConfig)

        config = prepare_resource_config(service, config)

        logger().info("Connecting to new %s resource `%s`..." % (service, name))
        globals.__GLOBAL_API_CLIENT__.connect_resource(name, service, config)

    def delete_integration(
        self,
        name: str,
    ) -> None:
        """Deprecated. Use `client.delete_resource()` instead."""
        logger().warning(
            "client.delete_integration() will be deprecated soon. Use `client.delete_resource() instead."
        )
        return self.delete_resource(name)

    def delete_resource(
        self,
        name: str,
    ) -> None:
        """Deletes the resource from Aqueduct.

        Args:
            name:
                The name of the resource to delete.
        """
        existing_resources = globals.__GLOBAL_API_CLIENT__.list_resources()

        # If the user uses the deprecated demo name, and there isn't a resource for this, that means they actually
        # want to use the new demo name.
        if (
            name == DEPRECATED_AQUEDUCT_DEMO_DB_NAME
            and DEPRECATED_AQUEDUCT_DEMO_DB_NAME not in existing_resources.keys()
        ) or name == AQUEDUCT_DEMO_DB_NAME:
            raise InvalidUserActionException("Cannot delete the Aqueduct demo database: %s" % name)
        if name not in existing_resources.keys():
            raise InvalidResourceException("Not connected to resource %s!" % name)

        # Update the connected resources cached on this object.
        globals.__GLOBAL_API_CLIENT__.delete_resource(existing_resources[name].id)
        self._connected_resources = globals.__GLOBAL_API_CLIENT__.list_resources()

    def list_integrations(self) -> Dict[str, ResourceInfo]:
        """Deprecated. Use `client.list_resources()` instead."""
        logger().warning(
            "client.list_resources() will be deprecated soon. Use `client.list_resources() instead."
        )
        return self.list_resources()

    def list_resources(self) -> Dict[str, ResourceInfo]:
        """Retrieves a dictionary of resources the client can use.

        Returns:
            A dictionary mapping from resource name to additional info.
        """
        self._connected_resources = globals.__GLOBAL_API_CLIENT__.list_resources()
        return self._connected_resources

    def integration(
        self,
        name: str,
    ) -> Union[
        SalesforceResource,
        S3Resource,
        GoogleSheetsResource,
        RelationalDBResource,
        AirflowResource,
        LambdaResource,
        MongoDBResource,
        DatabricksResource,
        SparkResource,
        AWSResource,
        ECRResource,
        DynamicK8sResource,
        GARResource,
    ]:
        """Deprecated. Use `client.resource()` instead."""
        logger().warning(
            "client.resource() will be deprecated soon. Use `client.resource() instead."
        )
        return self.resource(name)

    def resource(
        self, name: str
    ) -> Union[
        SalesforceResource,
        S3Resource,
        GoogleSheetsResource,
        RelationalDBResource,
        AirflowResource,
        LambdaResource,
        MongoDBResource,
        DatabricksResource,
        SparkResource,
        AWSResource,
        ECRResource,
        DynamicK8sResource,
        GARResource,
    ]:
        """Retrieves a connected resource object.

        Args:
            name:
                The name of the resource

        Returns:
            The resource object with the given name.

        Raises:
            InvalidResourceException:
                An error occurred because the cluster is not connected to the
                provided resource or the provided resource is of an
                incompatible type.
        """
        self._connected_resources = globals.__GLOBAL_API_CLIENT__.list_resources()
        connected_names = self._connected_resources.keys()

        if name == DEPRECATED_AQUEDUCT_DEMO_DB_NAME:
            # If the user uses the deprecated demo name, and there isn't a resource for this, that means they actually
            # want to use the new demo name. We implicitly convert this for them, with a warning.
            if DEPRECATED_AQUEDUCT_DEMO_DB_NAME not in connected_names:
                logger().warning(
                    "`%s` is the deprecated name for the aqueduct demo db. Please use `%s` instead."
                    % (DEPRECATED_AQUEDUCT_DEMO_DB_NAME, AQUEDUCT_DEMO_DB_NAME)
                )
                name = AQUEDUCT_DEMO_DB_NAME

        if name not in connected_names:
            raise InvalidResourceException("Not connected to resource %s!" % name)

        resource_info = self._connected_resources[name]
        if resource_info.service in RelationalDBServices:
            return RelationalDBResource(
                dag=self._dag,
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.SALESFORCE:
            return SalesforceResource(
                dag=self._dag,
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.GOOGLE_SHEETS:
            return GoogleSheetsResource(
                dag=self._dag,
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.S3:
            return S3Resource(
                dag=self._dag,
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.AIRFLOW:
            return AirflowResource(
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.K8S:
            return DynamicK8sResource(
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.LAMBDA:
            return LambdaResource(
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.MONGO_DB:
            return MongoDBResource(
                dag=self._dag,
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.DATABRICKS:
            return DatabricksResource(
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.SPARK:
            return SparkResource(
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.AWS:
            dynamic_k8s_resource_name = "%s:aqueduct_ondemand_k8s" % name
            dynamic_k8s_resource_info = self._connected_resources[dynamic_k8s_resource_name]
            return AWSResource(
                metadata=resource_info,
                k8s_resource_metadata=dynamic_k8s_resource_info,
            )
        elif resource_info.service == ServiceType.ECR:
            return ECRResource(
                metadata=resource_info,
            )
        elif resource_info.service == ServiceType.GAR:
            return GARResource(
                metadata=resource_info,
            )
        else:
            raise InvalidResourceException(
                "This method does not support loading resource of type %s" % resource_info.service
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
        flow_identifier: Optional[Union[str, uuid.UUID]] = None,
        flow_id: Optional[Union[str, uuid.UUID]] = None,
        flow_name: Optional[str] = None,
    ) -> Flow:
        """Fetches a flow corresponding to the given flow id.

        Args:
            flow_identifier:
                Used to identify the flow to fetch from the system.
                Use either the flow name or id as identifier to fetch
                from the system.
            flow_id:
                Used to identify the flow to fetch from the system.
                Between `flow_id` and `flow_name`, at least one must be provided.
                If both are specified, they must correspond to the same flow.
            flow_name:
                Used to identify the flow to fetch from the system.

            flow_identifier takes precedence over flow_id or flow_name arguments

        Raises:
            InvalidUserArgumentException:
                If the provided flow id or name does not correspond to a flow the client knows about.
        """
        # TODO(ENG-3013): Completely remove these optional parameters.
        if flow_id or flow_name:
            warnings.warn(
                "flow_id and flow_name arguments will be deprecated. Please use flow_identifier.",
                DeprecationWarning,
                stacklevel=2,
            )
            if not flow_identifier:
                if flow_id:
                    flow_identifier = flow_id
                elif flow_name:
                    flow_identifier = flow_name

        flows = [(flow.id, flow.name) for flow in globals.__GLOBAL_API_CLIENT__.list_workflows()]
        flow_id_key = find_flow_with_user_supplied_id_and_name(flows, flow_identifier)
        return Flow(
            flow_id_key,
            self._in_notebook_or_console_context,
        )

    def publish_flow(
        self,
        name: str,
        description: str = "",
        schedule: str = "",
        engine: Optional[Union[str, DynamicK8sResource]] = None,
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
        >>>     engine="k8s_resource",
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
                The name of the compute resource (eg. "my_lambda_resource") this the flow will
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

        if engine is not None and not (
            isinstance(engine, str) or isinstance(engine, DynamicK8sResource)
        ):
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
                self._connected_resources,
                globals.__GLOBAL_CONFIG__.engine,
            ),
            publish_flow_engine_config=generate_engine_config(self._connected_resources, engine),
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
        flow_identifier: Optional[Union[str, uuid.UUID]] = None,
        parameters: Optional[Dict[str, Any]] = None,
        flow_id: Optional[Union[str, uuid.UUID]] = None,
        flow_name: Optional[str] = None,
    ) -> None:
        """Immediately triggers another run of the provided flow.

        Args:
            flow_identifier:
                The uuid or name of the flow to delete.
            parameters:
                A map containing custom values to use for the designated parameters. The mapping
                is expected to be from parameter name to the custom value. These custom values
                are not persisted to the workflow. To actually change the default parameter values
                edit the workflow itself through `client.publish_flow()`.
            flow_id:
                Used to identify the flow to fetch from the system.
                Between `flow_id` and `flow_name`, at least one must be provided.
                If both are specified, they must correspond to the same flow.
            flow_name:
                Used to identify the flow to fetch from the system.

            flow_identifier takes precedence over flow_id or flow_name arguments
        Raises:
            InvalidRequestError:
                An error occurred when attempting to fetch the workflow to
                delete. The provided `flow_id` may be malformed.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        # TODO(ENG-3013): Completely remove these optional parameters.
        if flow_id or flow_name:
            warnings.warn(
                "flow_id and flow_name arguments will be deprecated. Please use flow_identifier.",
                DeprecationWarning,
                stacklevel=2,
            )
            if not flow_identifier:
                if flow_id:
                    flow_identifier = flow_id
                elif flow_name:
                    flow_identifier = flow_name
        param_specs: Dict[str, ParamSpec] = {}
        if parameters is not None:
            flow = self.flow(flow_identifier)
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
        flow_id_key = find_flow_with_user_supplied_id_and_name(flows, flow_identifier)
        globals.__GLOBAL_API_CLIENT__.refresh_workflow(flow_id_key, param_specs)

    def delete_flow(
        self,
        flow_identifier: Optional[Union[str, uuid.UUID]] = None,
        flow_id: Optional[Union[str, uuid.UUID]] = None,
        flow_name: Optional[str] = None,
        saved_objects_to_delete: Optional[
            DefaultDict[Union[str, BaseResource], List[SavedObjectUpdate]]
        ] = None,
        force: bool = False,
    ) -> None:
        """Deletes a flow object.

        Args:
            flow_identifier:
                The id of the flow to delete. Must be name or uuid
            flow_id:
                Used to identify the flow to fetch from the system.
                Between `flow_id` and `flow_name`, at least one must be provided.
                If both are specified, they must correspond to the same flow.
            flow_name:
                Used to identify the flow to fetch from the system.

            flow_identifier takes precedence over flow_id or flow_name arguments

            saved_objects_to_delete:
                The tables or storage paths to delete grouped by resource name.
            force:
                Force the deletion even though some workflow-written objects in the writes_to_delete argument had UpdateMode=append

        Raises:
            InvalidRequestError:
                An error occurred when attempting to fetch the workflow to
                delete. The provided `flow_id` may be malformed.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        # TODO(ENG-3013): Completely remove these optional parameters.
        if flow_id or flow_name:
            warnings.warn(
                "flow_id and flow_name arguments will be deprecated. Please use flow_identifier.",
                DeprecationWarning,
                stacklevel=2,
            )
            if not flow_identifier:
                if flow_id:
                    flow_identifier = flow_id
                elif flow_name:
                    flow_identifier = flow_name

        if saved_objects_to_delete is None:
            saved_objects_to_delete = defaultdict()

        # TODO(ENG-3015): Until parameterized S3 filepath deletion is fixed, we prevent users from
        #  deleting those objects.
        s3_parameterized_filepaths: List[Tuple[str, str]] = []
        for saved_obj_list in saved_objects_to_delete.values():
            for saved_obj_to_delete in saved_obj_list:
                if isinstance(saved_obj_to_delete.spec.parameters, S3LoadParams):
                    filepath = saved_obj_to_delete.spec.parameters.filepath
                    if len(re.findall(USER_TAG_PATTERN, filepath)) > 0:
                        s3_parameterized_filepaths.append(
                            (saved_obj_to_delete.resource_name, filepath)
                        )

        if len(s3_parameterized_filepaths) > 0:
            raise InvalidUserArgumentException(
                "Deleting objects at parameterized filepaths in S3 is currently unsupported. The following resource-filepath "
                "combinations in `saved_objects_to_delete` are parameterized: \n"
                + ", ".join(
                    [
                        f"{resource_name}: {filepath}"
                        for resource_name, filepath in s3_parameterized_filepaths
                    ]
                )
            )

        flows = [(flow.id, flow.name) for flow in globals.__GLOBAL_API_CLIENT__.list_workflows()]
        flow_id_key = find_flow_with_user_supplied_id_and_name(flows, flow_identifier)

        # TODO(ENG-410): This method gives no indication as to whether the flow
        #  was successfully deleted.
        resp = globals.__GLOBAL_API_CLIENT__.delete_workflow(
            flow_id_key, saved_objects_to_delete, force
        )

        failures = []
        for resource in resp.saved_object_deletion_results:
            for obj in resp.saved_object_deletion_results[resource]:
                if obj.exec_state.status == ExecutionStatus.FAILED:
                    trace = ""
                    if obj.exec_state.error:
                        context = obj.exec_state.error.context.strip().replace("\n", "\n>\t")
                        trace = f">\t{context}\n{obj.exec_state.error.tip}"
                    failure_string = f"[{resource}] {obj.name}\n{trace}"
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
        self._connected_resources = globals.__GLOBAL_API_CLIENT__.list_resources()
        print("Current Resources:")
        for resources in self._connected_resources:
            print("\t -" + resources)
