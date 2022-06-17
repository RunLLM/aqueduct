import io
import json
from typing import Any, Dict, List, Tuple, IO, Optional

import requests

from aqueduct.dag import DAG
from aqueduct.enums import ExecutionStatus
from aqueduct.error import (
    NoConnectedIntegrationsException,
    AqueductError,
    InternalAqueductError,
    ClientValidationError,
)

from aqueduct import utils
from aqueduct.logger import Logger
from aqueduct.operators import Operator
from aqueduct.integrations.integration import IntegrationInfo
from aqueduct.responses import PreviewResponse, RegisterWorkflowResponse


def _print_preview_logs(preview_resp: PreviewResponse, dag: DAG) -> None:
    """Prints all the logs generated during preview, in BFS order."""
    q: List[Operator] = dag.list_root_operators()
    seen_op_ids = set(op.id for op in q)
    while len(q) > 0:
        curr_op = q.pop(0)

        if curr_op.id in preview_resp.operator_results:
            curr_op_result = preview_resp.operator_results[curr_op.id]

            if curr_op_result.logs is not None and len(curr_op_result.logs) > 0:
                print('Operator "%s" Logs:\n' % curr_op.name)
                print(json.dumps(curr_op_result.logs, sort_keys=False, indent=4))

            if curr_op_result.err_msg and len(curr_op_result.err_msg) > 0:
                print(
                    "Operator %s Failed! Error message: \n %s"
                    % (curr_op.name, curr_op_result.err_msg)
                )
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


class APIClient:
    """
    Internal client class used to send requests to the aqueduct cluster.
    """

    HTTP_PREFIX = "http://"
    HTTPS_PREFIX = "https://"

    PREVIEW_ROUTE = "/api/preview"
    REGISTER_WORKFLOW_ROUTE = "/api/workflow/register"
    LIST_INTEGRATIONS_ROUTE = "/api/integrations"
    LIST_TABLES_ROUTE = "/api/tables"
    GET_WORKFLOW_ROUTE_TEMPLATE = "/api/workflow/%s"
    REFRESH_WORKFLOW_ROUTE_TEMPLATE = "/api/workflow/%s/refresh"
    DELETE_WORKFLOW_ROUTE_TEMPLATE = "/api/workflow/%s/delete"
    LIST_GITHUB_REPO_ROUTE = "/api/integrations/github/repos"
    LIST_GITHUB_BRANCH_ROUTE = "/api/integrations/github/branches"
    NODE_POSITION_ROUTE = "/api/positioning"

    def __init__(self, api_key: str, aqueduct_address: str):
        self.api_key = api_key
        self.aqueduct_address = aqueduct_address

        # If a dummy client is initialized, don't perform validation.
        if self.api_key == "" and self.aqueduct_address == "":
            Logger.logger.info(
                "Neither api key or server address were specified. This is a dummy client."
            )
            return

        # Check that the connection with the backend is working.
        if self.aqueduct_address.startswith(self.HTTP_PREFIX):
            self.aqueduct_address = self.aqueduct_address[len(self.HTTP_PREFIX) :]
            self.use_https = self._test_connection_protocol(try_http=True, try_https=False)
        elif self.aqueduct_address.startswith(self.HTTPS_PREFIX):
            self.aqueduct_address = self.aqueduct_address[len(self.HTTPS_PREFIX) :]
            self.use_https = self._test_connection_protocol(try_http=False, try_https=True)
        else:
            self.use_https = self._test_connection_protocol(try_http=True, try_https=True)

    def _construct_full_url(self, route_suffix: str, use_https: bool) -> str:
        protocol_prefix = self.HTTPS_PREFIX if use_https else self.HTTP_PREFIX
        return "%s%s%s" % (protocol_prefix, self.aqueduct_address, route_suffix)

    def _test_connection_protocol(self, try_http: bool, try_https: bool) -> bool:
        """Returns whether the connection uses https. Raises an exception if unable to connect at all.

        First tries https, then falls back to http.
        """
        assert try_http or try_https, "Must test at least one of http or https protocols."

        if try_https:
            try:
                url = self._construct_full_url(self.LIST_INTEGRATIONS_ROUTE, use_https=True)
                self._test_url(url)
                return True
            except Exception as e:
                Logger.logger.info(
                    "Testing if connection is HTTPS fails with:\n{}: {}".format(type(e).__name__, e)
                )

        if try_http:
            try:
                url = self._construct_full_url(self.LIST_INTEGRATIONS_ROUTE, use_https=False)
                self._test_url(url)
                return False
            except Exception as e:
                Logger.logger.info(
                    "Testing if connection is HTTP fails with:\n{}: {}".format(type(e).__name__, e)
                )

        raise ClientValidationError(
            "Unable to connect to server. Double check that both your API key `%s` and your specified address `%s` are correct. "
            % (self.api_key, self.aqueduct_address),
        )

    def _test_url(self, url: str) -> None:
        """Perform a get on the url with default headers, raising an error if anything goes wrong.

        We don't are about the value of the response, as long as the request succeeds.
        """
        headers = utils.generate_auth_headers(self.api_key)
        resp = requests.get(url, headers=headers)
        utils.raise_errors(resp)

    def list_integrations(self) -> Dict[str, IntegrationInfo]:
        url = self._construct_full_url(self.LIST_INTEGRATIONS_ROUTE, self.use_https)
        headers = utils.generate_auth_headers(self.api_key)
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
        url = self._construct_full_url(self.LIST_GITHUB_REPO_ROUTE, self.use_https)
        headers = utils.generate_auth_headers(self.api_key)

        resp = requests.get(url, headers=headers)
        return [x for x in resp.json()["repos"]]

    def list_github_branches(self, repo_url: str) -> List[str]:
        url = self._construct_full_url(self.LIST_GITHUB_BRANCH_ROUTE, self.use_https)
        headers = utils.generate_auth_headers(self.api_key)
        headers["github-repo"] = repo_url

        resp = requests.get(url, headers=headers)
        return [x for x in resp.json()["branches"]]

    def list_tables(self, limit: int) -> List[Tuple[str, str]]:
        url = self._construct_full_url(self.LIST_TABLES_ROUTE, self.use_https)
        headers = utils.generate_auth_headers(self.api_key)
        headers["limit"] = str(limit)
        resp = requests.get(url, headers=headers)
        utils.raise_errors(resp)

        return [(table["name"], table["owner"]) for table in resp.json()["tables"]]

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
        assert dag.workflow_id is None, "Unexpected internal field set when previewing a workflow."
        headers = {
            **utils.generate_auth_headers(self.api_key),
        }
        body = {
            "dag": dag.json(exclude_none=True),
        }

        files: Dict[str, IO[Any]] = {}
        for op in dag.list_operators():
            file = op.file()
            if file:
                files[str(op.id)] = io.BytesIO(file)

        url = self._construct_full_url(self.PREVIEW_ROUTE, self.use_https)
        resp = requests.post(url, headers=headers, data=body, files=files)
        utils.raise_errors(resp)

        preview_resp = PreviewResponse(**resp.json())
        _print_preview_logs(preview_resp, dag)
        if preview_resp.status == ExecutionStatus.PENDING:
            raise InternalAqueductError("Preview route should not be returning PENDING status.")
        if preview_resp.status == ExecutionStatus.FAILED:
            raise AqueductError(
                "Preview execution failed. See console logs for error message and trace."
            )
        return preview_resp

    def register_workflow(
        self,
        dag: DAG,
    ) -> RegisterWorkflowResponse:
        assert dag.workflow_id is None, "Unexpected internal field set when registering a workflow."

        headers = utils.generate_auth_headers(self.api_key)
        body = {
            "dag": dag.json(exclude_none=True),
        }

        files: Dict[str, IO[Any]] = {}
        for op in dag.list_operators():
            file = op.file()
            if file:
                files[str(op.id)] = io.BytesIO(file)

        url = self._construct_full_url(self.REGISTER_WORKFLOW_ROUTE, self.use_https)
        resp = requests.post(url, headers=headers, data=body, files=files)
        utils.raise_errors(resp)

        return RegisterWorkflowResponse(**resp.json())

    def refresh_workflow(
        self,
        flow_id: str,
        serialized_params: Optional[str] = None,
    ) -> None:
        headers = utils.generate_auth_headers(self.api_key)
        url = self._construct_full_url(
            self.REFRESH_WORKFLOW_ROUTE_TEMPLATE % flow_id, self.use_https
        )

        body = {}
        if serialized_params is not None:
            body["parameters"] = serialized_params

        response = requests.post(url, headers=headers, data=body)
        utils.raise_errors(response)

    def delete_workflow(self, flow_id: str) -> None:
        headers = utils.generate_auth_headers(self.api_key)
        url = self._construct_full_url(
            self.DELETE_WORKFLOW_ROUTE_TEMPLATE % flow_id, self.use_https
        )
        response = requests.post(url, headers=headers)
        utils.raise_errors(response)

    def get_workflow(self, flow_id: str) -> Any:
        headers = utils.generate_auth_headers(self.api_key)
        url = self._construct_full_url(self.GET_WORKFLOW_ROUTE_TEMPLATE % flow_id, self.use_https)
        response = requests.get(url, headers=headers)
        utils.raise_errors(response)
        return response.json()

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
        url = self._construct_full_url(self.NODE_POSITION_ROUTE, self.use_https)
        headers = {
            **utils.generate_auth_headers(self.api_key),
        }
        data = json.dumps(operator_mapping, sort_keys=False)
        resp = requests.post(url, headers=headers, data=data)
        utils.raise_errors(resp)

        resp_json = resp.json()

        return resp_json["operator_positions"], resp_json["artifact_positions"]
