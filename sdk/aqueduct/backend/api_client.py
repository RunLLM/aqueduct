import datetime
import io
import json
import uuid
from typing import IO, Any, DefaultDict, Dict, List, Optional, Tuple, Union

import requests
from aqueduct.constants.enums import ExecutionStatus, K8sClusterActionType, RuntimeType, ServiceType
from aqueduct.error import (
    AqueductError,
    ClientValidationError,
    InternalServerError,
    InvalidRequestError,
    InvalidUserActionException,
    NoConnectedResourcesException,
    ResourceNotFoundError,
    UnprocessableEntityError,
)
from aqueduct.logger import logger
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG, Metadata
from aqueduct.models.operators import Operator, ParamSpec
from aqueduct.models.resource import BaseResource, ResourceInfo
from aqueduct.models.response_models import (
    DeleteWorkflowResponse,
    DynamicEngineStatusResponse,
    GetDagResponse,
    GetDagResultResponse,
    GetImageURLResponse,
    GetNodeArtifactResponse,
    GetNodeOperatorResponse,
    GetVersionResponse,
    GetWorkflowDagResultResponse,
    GetWorkflowResponse,
    GetWorkflowV1Response,
    ListWorkflowResponseEntry,
    ListWorkflowSavedObjectsResponse,
    PreviewResponse,
    RegisterAirflowWorkflowResponse,
    RegisterWorkflowResponse,
    SavedObjectUpdate,
    WorkflowDagResponse,
    WorkflowDagResultResponse,
)
from aqueduct.utils.serialization import deserialize
from pkg_resources import get_distribution, parse_version

from ..resources.connect_config import DynamicK8sConfig, ResourceConfig
from .response_helpers import (
    _construct_preview_response,
    _handle_preview_resp,
    _parse_artifact_result_response,
)

# The maximum http request size is capped at 32 MB. DAG containing
# local data parameter(s) should not go beyond this value.
MAX_REQUEST_BODY_SIZE = 32 << 20


class APIClient:
    """
    Internal client class used to send requests to the aqueduct cluster.
    """

    HTTP_PREFIX = "http://"
    HTTPS_PREFIX = "https://"

    # V2
    GET_WORKFLOW_TEMPLATE = "/api/v2/workflow/%s"
    GET_DAGS_TEMPLATE = "/api/v2/workflow/%s/dags"
    GET_DAG_RESULTS_TEMPLATE = "/api/v2/workflow/%s/results"
    GET_NODES_TEMPLATE = "/api/v2/workflow/%s/dag/%s/nodes"

    # V1
    GET_VERSION_ROUTE = "/api/version"
    CONNECT_RESOURCE_ROUTE = "/api/resource/connect"
    DELETE_RESOURCE_ROUTE_TEMPLATE = "/api/resource/%s/delete"
    PREVIEW_ROUTE = "/api/preview"
    REGISTER_WORKFLOW_ROUTE = "/api/workflow/register"
    REGISTER_AIRFLOW_WORKFLOW_ROUTE = "/api/workflow/register/airflow"
    LIST_RESOURCES_ROUTE = "/api/resources"
    LIST_RESOURCE_OBJECTS_ROUTE_TEMPLATE = "/api/resource/%s/objects"
    GET_WORKFLOW_ROUTE_TEMPLATE = "/api/workflow/%s"
    GET_WORKFLOW_DAG_RESULT_TEMPLATE = "/api/workflow/%s/result/%s"
    LIST_WORKFLOW_SAVED_OBJECTS_ROUTE = "/api/workflow/%s/objects"
    GET_ARTIFACT_RESULT_TEMPLATE = "/api/artifact/%s/%s/result"

    LIST_WORKFLOWS_ROUTE = "/api/workflows"
    REFRESH_WORKFLOW_ROUTE_TEMPLATE = "/api/workflow/%s/refresh"
    DELETE_WORKFLOW_ROUTE_TEMPLATE = "/api/workflow/%s/delete"
    LIST_GITHUB_REPO_ROUTE = "/api/resources/github/repos"
    LIST_GITHUB_BRANCH_ROUTE = "/api/resources/github/branches"
    NODE_POSITION_ROUTE = "/api/positioning"
    EXPORT_FUNCTION_ROUTE = "/api/function/%s/export"

    GET_DYNAMIC_ENGINE_STATUS_ROUTE = "/api/resource/dynamic-engine/status"
    EDIT_DYNAMIC_ENGINE_ROUTE_TEMPLATE = "/api/resource/dynamic-engine/%s/edit"

    GET_IMAGE_URL_ROUTE = "/api/resource/container-registry/url"

    # Auth header
    API_KEY_HEADER = "api-key"

    configured = False

    def raise_errors(self, response: requests.Response) -> None:
        def _extract_err_msg() -> str:
            resp_json = response.json()
            if "error" not in resp_json:
                raise Exception("No 'error' field on response: %s" % json.dumps(resp_json))
            return str(resp_json["error"])

        if response.status_code == 400:
            raise InvalidRequestError(_extract_err_msg())
        if response.status_code == 403:
            raise ClientValidationError(_extract_err_msg())
        elif response.status_code == 422:
            raise UnprocessableEntityError(_extract_err_msg())
        elif response.status_code == 500:
            raise InternalServerError(_extract_err_msg())
        elif response.status_code == 404:
            raise ResourceNotFoundError(_extract_err_msg())
        elif response.status_code != 200:
            raise AqueductError(_extract_err_msg())

    def configure(self, api_key: str, aqueduct_address: str) -> None:
        self.api_key = api_key
        self.aqueduct_address = aqueduct_address

        # Clean URL
        if self.aqueduct_address.endswith("/"):
            self.aqueduct_address = self.aqueduct_address[:-1]

        self.configured = True

        # Check that the connection with the backend is working.
        try:
            if self.aqueduct_address.startswith(self.HTTP_PREFIX):
                self.aqueduct_address = self.aqueduct_address[len(self.HTTP_PREFIX) :]
                self.use_https = False
                success = self._check_for_server(use_https=False)
            elif self.aqueduct_address.startswith(self.HTTPS_PREFIX):
                self.aqueduct_address = self.aqueduct_address[len(self.HTTPS_PREFIX) :]
                self.use_https = True
                success = self._check_for_server(use_https=True)
            else:
                # If no http(s) prefix is provided, we'll try both.
                self.use_https = True
                success = self._check_for_server(use_https=True)
                if not success:
                    self.use_https = False
                    success = self._check_for_server(use_https=False)

        except Exception as e:
            self.configured = False
            raise e

        if not success:
            raise ClientValidationError(
                "Unable to connect to server. Double check that both your API key `%s` and your specified address `%s` are correct. "
                % (self.api_key, self.aqueduct_address),
            )

    def _check_config(self) -> None:
        if not self.configured:
            raise Exception(
                "API client has not been configured, please complete the configuration "
                "by initializing an Aqueduct client via: "
                "`client = aqueduct.Client(api_key, aqueduct_address)`"
            )

    def _generate_auth_headers(self) -> Dict[str, str]:
        self._check_config()
        return {self.API_KEY_HEADER: self.api_key}

    def construct_base_url(self, use_https: Optional[bool] = None) -> str:
        self._check_config()
        if use_https is None:
            use_https = self.use_https
        protocol_prefix = self.HTTPS_PREFIX if use_https else self.HTTP_PREFIX
        return "%s%s" % (protocol_prefix, self.aqueduct_address)

    def construct_full_url(self, route_suffix: str, use_https: Optional[bool] = None) -> str:
        self._check_config()
        if use_https is None:
            use_https = self.use_https
        return "%s%s" % (self.construct_base_url(use_https), route_suffix)

    def _validate_server_version(self, server_version: str) -> None:
        """Checks that the SDK and the server versions match."""
        sdk_version = get_distribution("aqueduct-sdk").version
        if parse_version(server_version) > parse_version(sdk_version):
            raise ClientValidationError(
                "The SDK is outdated, it is using version %s, while the server is of version %s. "
                "Please update your `aqueduct-sdk` package to the appropriate version by running "
                "`pip3 install aqueduct-sdk==<version>`. If running within a Jupyter notebook, "
                "remember to restart the kernel." % (sdk_version, server_version)
            )
        elif parse_version(server_version) < parse_version(sdk_version):
            raise ClientValidationError(
                "The server is outdated, it is using version %s, while the sdk is of version %s. "
                "Please update your server, or downgrade your SDK so that the versions match. "
                "The guide for updating the server is here: https://docs.aqueducthq.com/guides/updating-aqueduct"
                % (server_version, sdk_version)
            )

    def _check_for_server(self, use_https: bool) -> bool:
        """Check's if the server exists and can be connected to.

        Raises:
             ClientValidationError:
                If the server cannot be found, or if there is a versioning mismatch between server and sdk.

        """
        try:
            server_version = self._get_server_version(use_https=use_https)
        except Exception as e:
            logger().info(
                "Testing connection with {} fails with:\n\t{}: {}".format(
                    "HTTPS" if use_https else "HTTP",
                    type(e).__name__,
                    e,
                )
            )
            return False
        else:
            self._validate_server_version(server_version)
            return True

    def _get_server_version(self, use_https: bool) -> str:
        """Fetches the server's version number as a string."""
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.GET_VERSION_ROUTE, use_https=use_https)
        resp = requests.get(url, headers=headers)
        self.raise_errors(resp)
        return GetVersionResponse(**resp.json()).version

    def url_prefix(self) -> str:
        self._check_config()
        return self.HTTPS_PREFIX if self.use_https else self.HTTP_PREFIX

    def list_resources(self) -> Dict[str, ResourceInfo]:
        url = self.construct_full_url(self.LIST_RESOURCES_ROUTE)
        headers = self._generate_auth_headers()
        resp = requests.get(url, headers=headers)
        self.raise_errors(resp)
        if len(resp.json()) == 0:
            raise NoConnectedResourcesException(
                "Unable to create flow. Must be connected to at least one resource!"
            )

        return {
            resource_info["name"]: ResourceInfo(**resource_info) for resource_info in resp.json()
        }

    def list_github_repos(self) -> List[str]:
        url = self.construct_full_url(self.LIST_GITHUB_REPO_ROUTE)
        headers = self._generate_auth_headers()

        resp = requests.get(url, headers=headers)
        return [x for x in resp.json()["repos"]]

    def list_github_branches(self, repo_url: str) -> List[str]:
        url = self.construct_full_url(self.LIST_GITHUB_BRANCH_ROUTE)
        headers = self._generate_auth_headers()
        headers["github-repo"] = repo_url

        resp = requests.get(url, headers=headers)
        return [x for x in resp.json()["branches"]]

    def list_tables(self, resource_id: str) -> List[str]:
        """Returns a list of the tables in the specified resource.
        If the resource is not a relational database, it will throw an error.
        """
        url = self.construct_full_url(self.LIST_RESOURCE_OBJECTS_ROUTE_TEMPLATE % resource_id)
        headers = self._generate_auth_headers()
        resp = requests.get(url, headers=headers)
        self.raise_errors(resp)

        return [x for x in resp.json()["object_names"]]

    def connect_resource(
        self, name: str, service: Union[str, ServiceType], config: ResourceConfig
    ) -> None:
        resource_service = service
        if isinstance(resource_service, ServiceType):
            # The enum value needs to be used
            resource_service = resource_service.value

        headers = self._generate_auth_headers()
        headers.update(
            {
                "resource-name": name,
                "resource-service": resource_service,
                # `by_alias` is necessary to get this to use `schema` as a key for SnowflakeConfig.
                # `exclude_none` is necessary to exclude `role` when None as SnowflakeConfig.
                "resource-config": config.json(exclude_none=True, by_alias=True),
            }
        )
        url = self.construct_full_url(
            self.CONNECT_RESOURCE_ROUTE,
        )
        resp = requests.post(url, url, headers=headers)
        self.raise_errors(resp)

    def get_dynamic_engine_status_by_dag(
        self,
        dag: DAG,
    ) -> Dict[str, DynamicEngineStatusResponse]:
        """Makes a request against the /api/resource/dynamic-engine/status endpoint.
           If an resource id does not correspond to a dynamic resource, the response won't
           have an entry for that resource.

        Args:
            dag:
                The DAG object. We will extract the engine resource IDs and send them
                to the backend to retrieve their status. Currently, we are only interested in
                the status of dynamic engines.

        Returns:
            A DynamicEngineStatusResponse object, parsed from the backend endpoint's response.
        """
        engine_resource_ids = set()

        dag_engine_config = dag.engine_config
        if dag_engine_config.type == RuntimeType.K8S:
            assert dag_engine_config.k8s_config is not None
            engine_resource_ids.add(str(dag_engine_config.k8s_config.resource_id))
        for op in dag.operators.values():
            if op.spec.engine_config and op.spec.engine_config.type == RuntimeType.K8S:
                assert op.spec.engine_config.k8s_config is not None
                engine_resource_ids.add(str(op.spec.engine_config.k8s_config.resource_id))

        return self.get_dynamic_engine_status(list(engine_resource_ids))

    def get_dynamic_engine_status(
        self,
        engine_resource_ids: List[str],
    ) -> Dict[str, DynamicEngineStatusResponse]:
        """Makes a request against the /api/resource/dynamic-engine/status endpoint.
           If an resource id does not correspond to a dynamic resource, the response won't
           have an entry for that resource.

        Args:
            engine_resource_ids:
                A list of engine resource IDs. Currently, we are only interested in
                the status of dynamic engines.

        Returns:
            A DynamicEngineStatusResponse object, parsed from the backend endpoint's response.
        """
        headers = self._generate_auth_headers()
        headers["resource-ids"] = json.dumps(engine_resource_ids)

        url = self.construct_full_url(self.GET_DYNAMIC_ENGINE_STATUS_ROUTE)
        resp = requests.get(url, headers=headers)
        self.raise_errors(resp)

        return {
            dynamic_engine_status["name"]: DynamicEngineStatusResponse(**dynamic_engine_status)
            for dynamic_engine_status in resp.json()
        }

    def edit_dynamic_engine(
        self,
        action: K8sClusterActionType,
        resource_id: str,
        config_delta: Optional[DynamicK8sConfig] = None,
    ) -> None:
        """Makes a request against the /api/resource/dynamic-engine/{resourceId}/edit endpoint.

        Args:
            resource_id:
                The engine resource ID.
        """
        if action not in [
            K8sClusterActionType.CREATE,
            K8sClusterActionType.UPDATE,
            K8sClusterActionType.DELETE,
            K8sClusterActionType.FORCE_DELETE,
        ]:
            raise InvalidRequestError(
                "Invalid action %s for interacting with dynamic engine." % action
            )

        if config_delta == None:
            config_delta = DynamicK8sConfig()

        assert isinstance(config_delta, DynamicK8sConfig)

        headers = self._generate_auth_headers()
        headers["action"] = action.value

        url = self.construct_full_url(self.EDIT_DYNAMIC_ENGINE_ROUTE_TEMPLATE % resource_id)

        body = {
            "config_delta": config_delta.json(exclude_none=True),
        }

        resp = requests.post(url, headers=headers, data=body)

        self.raise_errors(resp)

    def delete_resource(
        self,
        resource_id: uuid.UUID,
    ) -> None:
        url = self.construct_full_url(self.DELETE_RESOURCE_ROUTE_TEMPLATE % resource_id)
        headers = self._generate_auth_headers()
        resp = requests.post(url, headers=headers)
        self.raise_errors(resp)

    def preview(
        self,
        dag: DAG,
    ) -> PreviewResponse:
        """Makes a request against the /api/preview endpoint.

        Args:
            dag:
                The DAG object to be serialized into the request header. Preview will
                execute this entire DAG object.

        Returns:
            A PreviewResponse object, parsed from the preview endpoint's response.
        """
        headers = self._generate_auth_headers()

        # TODO(ENG-2994): `by_alias` is required until this naming inconsistency is resolved.
        body = {
            "dag": dag.json(exclude_none=True, by_alias=True),
        }

        if len(body["dag"]) > MAX_REQUEST_BODY_SIZE and any(
            artifact_metadata.from_local_data for artifact_metadata in list(dag.artifacts.values())
        ):
            raise InvalidUserActionException(
                "Local Data after serialization is too large. Aqueduct uses json serialization. The maximum size of workflow with local data is %s in bytes, the current size is %s in bytes."
                % (MAX_REQUEST_BODY_SIZE, len(body["dag"]))
            )

        files: Dict[str, IO[Any]] = {}
        for op in dag.list_operators():
            file = op.file()
            if file:
                files[str(op.id)] = io.BytesIO(file)

        url = self.construct_full_url(self.PREVIEW_ROUTE)
        resp = requests.post(url, headers=headers, data=body, files=files)
        self.raise_errors(resp)

        preview_resp = _construct_preview_response(resp)
        _handle_preview_resp(preview_resp, dag)
        return preview_resp

    def register_workflow(
        self,
        dag: DAG,
        run_now: bool,
    ) -> RegisterWorkflowResponse:
        headers, body, files = self._construct_register_workflow_request(dag, run_now)
        url = self.construct_full_url(self.REGISTER_WORKFLOW_ROUTE)
        resp = requests.post(url, headers=headers, data=body, files=files)
        self.raise_errors(resp)

        return RegisterWorkflowResponse(**resp.json())

    def register_airflow_workflow(
        self,
        dag: DAG,
    ) -> RegisterAirflowWorkflowResponse:
        headers, body, files = self._construct_register_workflow_request(dag, False)
        url = self.construct_full_url(self.REGISTER_AIRFLOW_WORKFLOW_ROUTE, self.use_https)
        resp = requests.post(url, headers=headers, data=body, files=files)
        self.raise_errors(resp)

        return RegisterAirflowWorkflowResponse(**resp.json())

    def _construct_register_workflow_request(
        self,
        dag: DAG,
        run_now: bool,
    ) -> Tuple[Dict[str, str], Dict[str, str], Dict[str, IO[Any]]]:
        headers = self._generate_auth_headers()
        # This header value will be string "True" or "False"
        headers.update({"run-now": str(run_now)})
        # TODO(ENG-2994): `by_alias` is required until this naming inconsistency is resolved.
        body = {
            "dag": dag.json(exclude_none=True, by_alias=True),
        }

        if len(body["dag"]) > MAX_REQUEST_BODY_SIZE and any(
            artifact_metadata.from_local_data for artifact_metadata in list(dag.artifacts.values())
        ):
            raise InvalidUserActionException(
                "Local Data after serialization is too large. Aqueduct uses json serialization. The maximum size of workflow with local data is %s in bytes, the current size is %s in bytes."
                % (MAX_REQUEST_BODY_SIZE, len(body["dag"]))
            )

        files: Dict[str, IO[Any]] = {}
        for op in dag.list_operators():
            file = op.file()
            if file:
                files[str(op.id)] = io.BytesIO(file)

        return headers, body, files

    def refresh_workflow(
        self,
        flow_id: str,
        param_specs: Dict[str, ParamSpec],
    ) -> None:
        """
        `param_specs`: a dictionary from parameter names to its corresponding new ParamSpec.
        """
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.REFRESH_WORKFLOW_ROUTE_TEMPLATE % flow_id)

        body = {
            "parameters": json.dumps(
                {param_name: param_spec.dict() for param_name, param_spec in param_specs.items()}
            )
        }

        response = requests.post(url, headers=headers, data=body)
        self.raise_errors(response)

    def delete_workflow(
        self,
        flow_id: str,
        saved_objects_to_delete: DefaultDict[Union[str, BaseResource], List[SavedObjectUpdate]],
        force: bool,
    ) -> DeleteWorkflowResponse:
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.DELETE_WORKFLOW_ROUTE_TEMPLATE % flow_id)
        body = {
            "external_delete": {
                # TODO(ENG-2994): `by_alias` is required until this naming inconsistency is resolved.
                str(resource): [
                    obj.spec.json(by_alias=True) for obj in saved_objects_to_delete[resource]
                ]
                for resource in saved_objects_to_delete
            },
            "force": force,
        }
        response = requests.post(url, headers=headers, json=body)
        self.raise_errors(response)
        return DeleteWorkflowResponse(**response.json())

    def get_workflow(self, flow_id: str) -> GetWorkflowV1Response:
        headers = self._generate_auth_headers()

        url = self.construct_full_url(self.GET_WORKFLOW_TEMPLATE % flow_id)
        flow_response = requests.get(url, headers=headers)
        self.raise_errors(flow_response)
        resp_flow = GetWorkflowResponse(**flow_response.json())
        metadata = Metadata(
            name=resp_flow.name,
            description=resp_flow.description,
            schedule=resp_flow.schedule,
            retention_policy=resp_flow.retention_policy,
        )

        url = self.construct_full_url(self.GET_DAGS_TEMPLATE % flow_id)
        dags_response = requests.get(url, headers=headers)
        self.raise_errors(dags_response)
        resp_dags = {dag["id"]: GetDagResponse(**dag) for dag in dags_response.json()}
        # Metadata from WorkflowResponse so it is the same for all DAG.
        dags = {}
        for dag in resp_dags.values():
            url = self.construct_full_url(self.GET_NODES_TEMPLATE % (flow_id, dag.id))
            nodes_resp = requests.get(url, headers=headers)
            self.raise_errors(nodes_resp)
            resp_nodes = nodes_resp.json()

            ops = {}
            for operator in resp_nodes["operators"]:
                op = GetNodeOperatorResponse(**operator)
                ops[str(op.id)] = Operator(
                    id=op.id,
                    name=op.name,
                    description=op.description,
                    spec=op.spec,
                    inputs=op.inputs,
                    outputs=op.outputs,
                )

            artfs = {}
            for artifact in resp_nodes["artifacts"]:
                artf = GetNodeArtifactResponse(**artifact)
                artfs[str(artf.id)] = ArtifactMetadata(
                    id=artf.id,
                    name=artf.name,
                    type=artf.type,
                )

            dags[dag.id] = WorkflowDagResponse(
                id=dag.id,
                workflow_id=dag.workflow_id,
                metadata=metadata,
                operators=ops,
                artifacts=artfs,
            )

        url = self.construct_full_url(self.GET_DAG_RESULTS_TEMPLATE % flow_id)
        results_resp = requests.get(url, headers=headers)
        self.raise_errors(results_resp)
        dag_results = [GetDagResultResponse(**dag_result) for dag_result in results_resp.json()]
        workflow_dag_results = [
            WorkflowDagResultResponse(
                id=dag_result.id,
                created_at=int(
                    datetime.datetime.strptime(
                        resp_dags[str(dag_result.dag_id)].created_at[:-4],
                        "%Y-%m-%dT%H:%M:%S.%f"
                        if resp_dags[str(dag_result.dag_id)].created_at[-1] == "Z"
                        else "%Y-%m-%dT%H:%M:%S.%f%z",
                    ).timestamp()
                ),
                status=dag_result.exec_state.status,
                exec_state=dag_result.exec_state,
                workflow_dag_id=dag_result.dag_id,
            )
            for dag_result in dag_results
        ]
        return GetWorkflowV1Response(workflow_dags=dags, workflow_dag_results=workflow_dag_results)

    def get_workflow_dag_result(self, flow_id: str, result_id: str) -> GetWorkflowDagResultResponse:
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.GET_WORKFLOW_DAG_RESULT_TEMPLATE % (flow_id, result_id))
        resp = requests.get(url, headers=headers)
        self.raise_errors(resp)
        return GetWorkflowDagResultResponse(**resp.json())

    def list_saved_objects(self, flow_id: str) -> ListWorkflowSavedObjectsResponse:
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.LIST_WORKFLOW_SAVED_OBJECTS_ROUTE % flow_id)
        resp = requests.get(url, headers=headers)
        self.raise_errors(resp)
        workflow_writes_response = ListWorkflowSavedObjectsResponse(**resp.json())
        return workflow_writes_response

    def list_workflows(self) -> List[ListWorkflowResponseEntry]:
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.LIST_WORKFLOWS_ROUTE)
        response = requests.get(url, headers=headers)
        self.raise_errors(response)

        return [ListWorkflowResponseEntry(**workflow) for workflow in response.json()]

    def get_artifact_result_data(
        self, dag_result_id: str, artifact_id: str
    ) -> Tuple[Optional[Any], ExecutionStatus]:
        """Returns an empty string if the operator was not successfully executed."""
        headers = self._generate_auth_headers()
        url = self.construct_full_url(
            self.GET_ARTIFACT_RESULT_TEMPLATE % (dag_result_id, artifact_id)
        )
        resp = requests.get(url, headers=headers)
        self.raise_errors(resp)

        parsed_response = _parse_artifact_result_response(resp)
        execution_status = parsed_response["metadata"]["exec_state"]["status"]

        serialization_type = parsed_response["metadata"]["serialization_type"]
        artifact_type = parsed_response["metadata"]["artifact_type"]

        return_value = None
        if "data" in parsed_response:
            return_value = deserialize(serialization_type, artifact_type, parsed_response["data"])

        if execution_status != ExecutionStatus.SUCCEEDED:
            logger().warning("Artifact result unavailable due to unsuccessful execution.")

        return (return_value, execution_status)

    def get_image_url(
        self,
        resource_id: str,
        service: ServiceType,
        image_name: str,
    ) -> GetImageURLResponse:
        """Makes a request against the /api/resource/container-registry/url endpoint.

        Args:
            resource_id:
                Container registry resource ID.
            service:
                Container registry service type.
            image_name:
                Image name to get the URL for.

        Returns:
            GetImageURLResponse that contains the URL to the image.
        """
        headers = self._generate_auth_headers()
        headers["resource-id"] = resource_id
        headers["service"] = service.value
        headers["image-name"] = image_name

        url = self.construct_full_url(self.GET_IMAGE_URL_ROUTE)
        resp = requests.get(url, headers=headers)
        self.raise_errors(resp)

        return GetImageURLResponse(**resp.json())
