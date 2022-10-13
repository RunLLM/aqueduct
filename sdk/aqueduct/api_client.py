import io
import json
from typing import IO, Any, DefaultDict, Dict, List, Optional, Tuple, Union

import multipart
import requests
from aqueduct.dag import DAG
from aqueduct.enums import ExecutionStatus
from aqueduct.error import (
    AqueductError,
    ClientValidationError,
    InternalAqueductError,
    NoConnectedIntegrationsException,
)
from aqueduct.integrations.integration import Integration, IntegrationInfo
from aqueduct.logger import logger
from aqueduct.operators import Operator, ParamSpec
from aqueduct.responses import (
    ArtifactResult,
    DeleteWorkflowResponse,
    GetVersionResponse,
    GetWorkflowResponse,
    ListWorkflowResponseEntry,
    ListWorkflowSavedObjectsResponse,
    Logs,
    OperatorResult,
    PreviewResponse,
    RegisterAirflowWorkflowResponse,
    RegisterWorkflowResponse,
    SavedObjectUpdate,
)
from aqueduct.serialization import deserialize
from aqueduct.utils import GITHUB_ISSUE_LINK, indent_multiline_string
from pkg_resources import parse_version, require
from requests_toolbelt.multipart import decoder

from aqueduct import utils


def _handle_preview_resp(preview_resp: PreviewResponse, dag: DAG) -> None:
    """
    Prints all the logs generated during preview, in BFS order.

    Raises:
        AqueductError:
            If the preview execution has failed. This error will have the context
            and error message of every failed operator in it.
        InternalAqueductError:
            If something unexpected happened in our system.
    """
    # There can be multiple operator failures, one for each entry.
    op_err_msgs: List[str] = []

    def _construct_failure_error_msg(op_name: str, op_result: OperatorResult) -> str:
        """This is the message is raised in the Exception message."""
        assert op_result.error is not None
        return (
            f"Operator `{op_name}` failed!\n"
            f"{op_result.error.context}\n"
            f"\n"
            f"{op_result.error.tip}\n"
            f"\n"
        )

    def _print_op_user_logs(op_name: str, logs: Logs) -> None:
        """Prints out the logs for a single operator. The format is:

        stdout:
            {logs}
            {logs}
        ----------------------------------
        stderr:
            {logs}
            {logs}

        If either stdout or stderr is empty, we do not print anything for
        the empty section, and do not draw the "--" delimiter line.
        """
        if logs.is_empty():
            return

        print(f"Operator {op_name} Logs:")
        if len(logs.stdout) > 0:
            print("stdout:")
            print(indent_multiline_string(logs.stdout).rstrip("\n"))

        if len(logs.stdout) > 0 and len(logs.stderr) > 0:
            print("----------------------------------")

        if len(logs.stderr) > 0:
            print("stderr:")
            print(indent_multiline_string(logs.stderr).rstrip("\n"))
        print("")

    q: List[Operator] = dag.list_root_operators()
    seen_op_ids = set(op.id for op in q)
    while len(q) > 0:
        curr_op = q.pop(0)

        if curr_op.id in preview_resp.operator_results:
            curr_op_result = preview_resp.operator_results[curr_op.id]

            if curr_op_result.user_logs is not None:
                _print_op_user_logs(curr_op.name, curr_op_result.user_logs)

            if curr_op_result.error is not None:
                op_err_msgs.append(_construct_failure_error_msg(curr_op.name, curr_op_result))
            else:
                # Continue traversing, marking operators added to the queue as "seen"
                for output_artifact_id in curr_op.outputs:
                    next_operators = [
                        op
                        for op in dag.list_operators(on_artifact_id=output_artifact_id)
                        if op.id not in seen_op_ids
                    ]
                    q.extend(next_operators)
                    seen_op_ids.union(set(op.id for op in next_operators))

    if preview_resp.status == ExecutionStatus.PENDING:
        raise InternalAqueductError("Preview route should not be returning PENDING status.")

    if preview_resp.status == ExecutionStatus.FAILED:
        # If non of the operators failed, this must be an issue with our
        if len(op_err_msgs) == 0:
            raise InternalAqueductError(
                f"Unexpected Server Error! If this issue persists, please file a bug report in github: "
                f"{GITHUB_ISSUE_LINK} . We will get back to you as soon as we can.",
            )

        failure_err_msg = "\n".join(op_err_msgs)
        raise AqueductError(f"Preview Execution Failed:\n\n{failure_err_msg}\n")


class APIClient:
    """
    Internal client class used to send requests to the aqueduct cluster.
    """

    HTTP_PREFIX = "http://"
    HTTPS_PREFIX = "https://"

    GET_VERSION_ROUTE = "/api/version"
    PREVIEW_ROUTE = "/api/preview"
    REGISTER_WORKFLOW_ROUTE = "/api/workflow/register"
    REGISTER_AIRFLOW_WORKFLOW_ROUTE = "/api/workflow/register/airflow"
    LIST_INTEGRATIONS_ROUTE = "/api/integrations"
    LIST_TABLES_ROUTE = "/api/tables"
    GET_WORKFLOW_ROUTE_TEMPLATE = "/api/workflow/%s"
    LIST_WORKFLOW_SAVED_OBJECTS_ROUTE = "/api/workflow/%s/objects"
    GET_ARTIFACT_RESULT_TEMPLATE = "/api/artifact/%s/%s/result"

    LIST_WORKFLOWS_ROUTE = "/api/workflows"
    REFRESH_WORKFLOW_ROUTE_TEMPLATE = "/api/workflow/%s/refresh"
    DELETE_WORKFLOW_ROUTE_TEMPLATE = "/api/workflow/%s/delete"
    LIST_GITHUB_REPO_ROUTE = "/api/integrations/github/repos"
    LIST_GITHUB_BRANCH_ROUTE = "/api/integrations/github/branches"
    NODE_POSITION_ROUTE = "/api/positioning"
    EXPORT_FUNCTION_ROUTE = "/api/function/%s/export"

    # Auth header
    API_KEY_HEADER = "api-key"

    configured = False

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
                "API client has not been configured, please complete the configuration \
                by initializing an Aqueduct client with the api key and the server address."
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
        sdk_version = require("aqueduct-sdk")[0].version
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
        utils.raise_errors(resp)
        return GetVersionResponse(**resp.json()).version

    def url_prefix(self) -> str:
        self._check_config()
        return self.HTTPS_PREFIX if self.use_https else self.HTTP_PREFIX

    def list_integrations(self) -> Dict[str, IntegrationInfo]:
        url = self.construct_full_url(self.LIST_INTEGRATIONS_ROUTE)
        headers = self._generate_auth_headers()
        resp = requests.get(url, headers=headers)
        utils.raise_errors(resp)
        if len(resp.json()) == 0:
            raise NoConnectedIntegrationsException(
                "Unable to create flow. Must be connected to at least one integration!"
            )

        return {
            integration_info["name"]: IntegrationInfo(**integration_info)
            for integration_info in resp.json()
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

    def list_tables(self, limit: int) -> List[Tuple[str, str]]:
        url = self.construct_full_url(self.LIST_TABLES_ROUTE)
        headers = self._generate_auth_headers()
        headers["limit"] = str(limit)
        resp = requests.get(url, headers=headers)
        utils.raise_errors(resp)

        return [(table["name"], table["owner"]) for table in resp.json()["tables"]]

    def _construct_preview_response(self, response: requests.Response) -> PreviewResponse:
        artifact_results = {}
        artifact_result_constructor = {}
        preview_response = {}
        is_metadata_received = False
        multipart_data = decoder.MultipartDecoder.from_response(response)
        parse = multipart.parse_options_header

        for part in multipart_data.parts:
            field_name = part.headers[b"Content-Disposition"].decode(multipart_data.encoding)
            field_name = parse(field_name)[1]["name"]

            if field_name == "metadata":
                is_metadata_received = True
                metadata = json.loads(part.content.decode(multipart_data.encoding))
            elif utils.is_string_valid_uuid(field_name):
                if is_metadata_received:
                    artifact_result_constructor = metadata["artifact_types_metadata"][field_name]
                    artifact_result_constructor["content"] = part.content
                    artifact_results[field_name] = ArtifactResult(**artifact_result_constructor)
                else:
                    raise AqueductError("Unable to retrieve artifacts metadata")
            else:
                raise AqueductError("Unable to get correct preview response")

        preview_response["status"] = metadata["status"]
        preview_response["operator_results"] = metadata["operator_results"]
        preview_response["artifact_results"] = artifact_results

        return PreviewResponse(**preview_response)

    def preview(
        self,
        dag: DAG,
    ) -> PreviewResponse:
        """Makes a request against the /preview endpoint.

        Args:
            dag:
                The DAG object to be serialized into the request header. Preview will
                execute this entire DAG object.

        Returns:
            A PreviewResponse object, parsed from the preview endpoint's response.
        """
        headers = self._generate_auth_headers()
        body = {
            "dag": dag.json(exclude_none=True),
        }

        files: Dict[str, IO[Any]] = {}
        for op in dag.list_operators():
            file = op.file()
            if file:
                files[str(op.id)] = io.BytesIO(file)

        url = self.construct_full_url(self.PREVIEW_ROUTE)
        resp = requests.post(url, headers=headers, data=body, files=files)
        utils.raise_errors(resp)

        preview_resp = self._construct_preview_response(resp)
        _handle_preview_resp(preview_resp, dag)
        return preview_resp

    def register_workflow(
        self,
        dag: DAG,
    ) -> RegisterWorkflowResponse:
        headers, body, files = self._construct_register_workflow_request(dag)
        url = self.construct_full_url(self.REGISTER_WORKFLOW_ROUTE)
        resp = requests.post(url, headers=headers, data=body, files=files)
        utils.raise_errors(resp)

        return RegisterWorkflowResponse(**resp.json())

    def register_airflow_workflow(
        self,
        dag: DAG,
    ) -> RegisterAirflowWorkflowResponse:
        headers, body, files = self._construct_register_workflow_request(dag)
        url = self.construct_full_url(self.REGISTER_AIRFLOW_WORKFLOW_ROUTE, self.use_https)
        resp = requests.post(url, headers=headers, data=body, files=files)
        utils.raise_errors(resp)

        return RegisterAirflowWorkflowResponse(**resp.json())

    def _construct_register_workflow_request(
        self,
        dag: DAG,
    ) -> Tuple[Dict[str, str], Dict[str, str], Dict[str, IO[Any]]]:
        headers = self._generate_auth_headers()
        body = {
            "dag": dag.json(exclude_none=True),
        }

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
        utils.raise_errors(response)

    def delete_workflow(
        self,
        flow_id: str,
        saved_objects_to_delete: DefaultDict[Union[str, Integration], List[SavedObjectUpdate]],
        force: bool,
    ) -> DeleteWorkflowResponse:
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.DELETE_WORKFLOW_ROUTE_TEMPLATE % flow_id)
        body = {
            "external_delete": {
                str(integration): [obj.object_name for obj in saved_objects_to_delete[integration]]
                for integration in saved_objects_to_delete
            },
            "force": force,
        }
        response = requests.post(url, headers=headers, json=body)
        utils.raise_errors(response)
        deleteWorkflowResponse = DeleteWorkflowResponse(**response.json())
        return deleteWorkflowResponse

    def get_workflow(self, flow_id: str) -> GetWorkflowResponse:
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.GET_WORKFLOW_ROUTE_TEMPLATE % flow_id)
        resp = requests.get(url, headers=headers)
        utils.raise_errors(resp)
        workflow_response = GetWorkflowResponse(**resp.json())
        return workflow_response

    def list_saved_objects(self, flow_id: str) -> ListWorkflowSavedObjectsResponse:
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.LIST_WORKFLOW_SAVED_OBJECTS_ROUTE % flow_id)
        resp = requests.get(url, headers=headers)
        utils.raise_errors(resp)
        workflow_writes_response = ListWorkflowSavedObjectsResponse(**resp.json())
        return workflow_writes_response

    def list_workflows(self) -> List[ListWorkflowResponseEntry]:
        headers = self._generate_auth_headers()
        url = self.construct_full_url(self.LIST_WORKFLOWS_ROUTE)
        response = requests.get(url, headers=headers)
        utils.raise_errors(response)

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
        utils.raise_errors(resp)

        parsed_response = utils.parse_artifact_result_response(resp)
        execution_status = parsed_response["metadata"]["exec_state"]["status"]

        if execution_status != ExecutionStatus.SUCCEEDED:
            print("Artifact result unavailable due to unsuccessful execution.")
            return None, execution_status

        serialization_type = parsed_response["metadata"]["serialization_type"]
        artifact_type = parsed_response["metadata"]["artifact_type"]
        return (
            deserialize(serialization_type, artifact_type, parsed_response["data"]),
            execution_status,
        )

    def get_node_positions(
        self, operator_mapping: Dict[str, Dict[str, Any]]
    ) -> Tuple[Dict[str, Dict[str, float]], Dict[str, Dict[str, float]]]:
        """Queries the `self.NODE_POSITION_ROUTE` endpoint for graph display's nodes' positions.

        Args:
            operator_mapping:
                The mapping between each operator in the graph and all required metadata.
                This is serialized into a json and passed directly to the endpoint.
                The expected keys are:
                    `inputs`: list of the input artifacts' UUIDs
                    `output`: list of the output artifacts' UUIDs

        Returns:
            Two mappings of UUIDs to positions, structured as a dictionary with the keys "x" and "y".
            The first mapping is for operators and the second is for artifacts.
        """
        url = self.construct_full_url(self.NODE_POSITION_ROUTE)
        headers = self._generate_auth_headers()
        data = json.dumps(operator_mapping, sort_keys=False)
        resp = requests.post(url, headers=headers, data=data)
        utils.raise_errors(resp)

        resp_json = resp.json()

        return resp_json["operator_positions"], resp_json["artifact_positions"]
