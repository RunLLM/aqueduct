import json
import os
import uuid
from pathlib import Path

import pytest
import requests
import utils
from aqueduct.models.response_models import (
    GetArtifactResultResponse,
    GetDagResultResponse,
    GetNodeArtifactResponse,
    GetNodeOperatorResponse,
    GetOperatorResultResponse,
)
from aqueduct_executor.operators.utils.enums import JobType
from exec_state import assert_exec_state
from setup.changing_saves_workflow import setup_changing_saves
from setup.flow_with_failure import setup_flow_with_failure
from setup.flow_with_metrics_and_checks import setup_flow_with_metrics_and_checks
from setup.flow_with_multiple_operators import setup_flow_with_multiple_operators
from setup.flow_with_sleep import setup_flow_with_sleep

import aqueduct
from aqueduct import globals


class TestBackend:
    # V2
    GET_WORKFLOWS_TEMPLATE = "/api/v2/workflows"

    GET_DAGS_TEMPLATE = "/api/v2/workflow/%s/dags"
    GET_DAG_RESULTS_TEMPLATE = "/api/v2/workflow/%s/results"
    GET_NODES_RESULTS_TEMPLATE = "/api/v2/workflow/%s/result/%s/nodes/results"

    GET_NODES_TEMPLATE = "/api/v2/workflow/%s/dag/%s/nodes"

    GET_NODE_ARTIFACT_TEMPLATE = "/api/v2/workflow/%s/dag/%s/node/artifact/%s"
    GET_NODE_ARTIFACT_RESULT_CONTENT_TEMPLATE = (
        "/api/v2/workflow/%s/dag/%s/node/artifact/%s/result/%s/content"
    )
    GET_NODE_ARTIFACT_RESULTS_TEMPLATE = "/api/v2/workflow/%s/dag/%s/node/artifact/%s/results"

    GET_NODE_OPERATOR_TEMPLATE = "/api/v2/workflow/%s/dag/%s/node/operator/%s"
    GET_NODE_OPERATOR_CONTENT_TEMPLATE = "/api/v2/workflow/%s/dag/%s/node/operator/%s/content"

    # V1
    LIST_WORKFLOW_SAVED_OBJECTS_TEMPLATE = "/api/workflow/%s/objects"
    GET_TEST_INTEGRATION_TEMPLATE = "/api/integration/%s/test"
    LIST_INTEGRATIONS_TEMPLATE = "/api/integrations"
    CONNECT_INTEGRATION_TEMPLATE = "/api/integration/connect"
    DELETE_INTEGRATION_TEMPLATE = "/api/integration/%s/delete"
    GET_WORKFLOW_RESULT_TEMPLATE = "/api/workflow/%s/result/%s"
    LIST_ARTIFACT_RESULTS_TEMPLATE = "/api/workflow/%s/artifact/%s/results"

    WORKFLOW_PATH = Path(__file__).parent / "setup"
    DEMO_DB_PATH = os.path.join(os.environ["HOME"], ".aqueduct/server/db/demo.db")

    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client()
        cls.integration = cls.client.resource(name=pytest.integration)
        cls.flows = {
            "changing_saves": setup_changing_saves(cls.client, pytest.integration),
            "flow_with_multiple_operators": setup_flow_with_multiple_operators(
                cls.client, pytest.integration
            ),
            "flow_with_failure": setup_flow_with_failure(cls.client, pytest.integration),
            "flow_with_metrics_and_checks": setup_flow_with_metrics_and_checks(
                cls.client,
                pytest.integration,
            ),
            # this flow is intended to provide 'noise' of op / artf with the same name,
            # but under different flow.
            "another_flow_with_metrics_and_checks": setup_flow_with_metrics_and_checks(
                cls.client,
                pytest.integration,
                workflow_name="another_flow_with_metrics_and_checks",
            ),
        }

        # we do not call `wait_for_flow_runs` on these flows
        cls.running_flows = {
            "flow_with_sleep": setup_flow_with_sleep(cls.client, pytest.integration),
        }
        for flow_id, n_runs in cls.flows.values():
            utils.wait_for_flow_runs(cls.client, flow_id, n_runs)

    @classmethod
    def teardown_class(cls):
        for flow_id, _ in cls.flows.values():
            utils.delete_flow(cls.client, flow_id)

        for flow_id, _ in cls.running_flows.values():
            utils.delete_flow(cls.client, flow_id)

    @classmethod
    def response(cls, endpoint, additional_headers):
        headers = {"api-key": pytest.api_key}
        headers.update(additional_headers)
        url = globals.__GLOBAL_API_CLIENT__.construct_full_url(endpoint)
        return url, headers

    @classmethod
    def get_response(cls, endpoint, additional_headers={}):
        url, headers = cls.response(endpoint, additional_headers)
        r = requests.get(url, headers=headers)
        return r

    @classmethod
    def post_response(cls, endpoint, additional_headers={}):
        url, headers = cls.response(endpoint, additional_headers)
        r = requests.post(url, headers=headers)
        return r

    def test_endpoint_list_workflow_tables(self):
        endpoint = self.LIST_WORKFLOW_SAVED_OBJECTS_TEMPLATE % self.flows["changing_saves"][0]
        data = self.get_response(endpoint).json()["object_details"]

        assert len(data) == 3

        # table_name, update_mode
        data_set = set(
            [
                ("table_1", "append"),
                ("table_1", "replace"),
                ("table_2", "replace"),
            ]
        )

        print(data)
        assert (
            set(
                [
                    (item["spec"]["parameters"]["table"], item["spec"]["parameters"]["update_mode"])
                    for item in data
                ]
            )
            == data_set
        )

        # Check all in same integration
        assert len(set([item["integration_name"] for item in data])) == 1
        assert len(set([item["spec"]["service"] for item in data])) == 1

    def test_endpoint_delete_integration(self):
        integration_name = f"test_delete_integration_{uuid.uuid4().hex[:8]}"

        # Check integration did not exist
        data = self.get_response(self.LIST_INTEGRATIONS_TEMPLATE).json()
        assert integration_name not in set([integration["name"] for integration in data])

        # Create integration
        status = self.post_response(
            self.CONNECT_INTEGRATION_TEMPLATE,
            additional_headers={
                "integration-name": integration_name,
                "integration-service": "SQLite",
                "integration-config": json.dumps({"database": self.DEMO_DB_PATH}),
            },
        ).status_code
        assert status == 200

        # Check integration created
        data = self.get_response(self.LIST_INTEGRATIONS_TEMPLATE).json()
        integration_data = {integration["name"]: integration["id"] for integration in data}
        assert integration_name in set(integration_data.keys())

        # Delete integration
        status = self.post_response(
            self.DELETE_INTEGRATION_TEMPLATE % integration_data[integration_name]
        ).status_code
        assert status == 200

        # Check integration does not exist
        data = self.get_response(self.LIST_INTEGRATIONS_TEMPLATE).json()
        assert integration_name not in set([integration["name"] for integration in data])

    def test_endpoint_test_integration(self):
        resp = self.get_response(self.GET_TEST_INTEGRATION_TEMPLATE % self.integration.id())
        assert resp.ok

    def test_endpoint_get_workflow_dag_result_with_failure(self):
        flow_id = self.flows["flow_with_failure"][0]
        flow = self.client.flow(flow_id)
        runs = flow.list_runs()
        resp = self.get_response(
            self.GET_WORKFLOW_RESULT_TEMPLATE % (flow_id, runs[0]["run_id"])
        ).json()
        assert_exec_state(resp["result"]["exec_state"], "failed")
        # operators
        operators = resp["operators"]
        assert len(operators) == 3
        for op in operators.values():
            name = op["name"]
            exec_state = op["result"]["exec_state"]

            if "query" in name:  # extract
                assert_exec_state(exec_state, "succeeded")
            elif name == "bad_op":
                assert_exec_state(exec_state, "failed")
            elif name == "bad_op_downstream":
                assert_exec_state(exec_state, "canceled")
            else:
                raise Exception(f"unexpected operator name {name}")

        # artifacts
        artifacts = resp["artifacts"]
        assert len(artifacts) == 3
        for artf in artifacts.values():
            name = artf["name"]
            exec_state = artf["result"]["exec_state"]

            if "query" in name:
                assert_exec_state(exec_state, "succeeded")
            elif name == "bad_op artifact":
                assert_exec_state(exec_state, "canceled")
            elif name == "bad_op_downstream artifact":
                assert_exec_state(exec_state, "canceled")
            else:
                raise Exception(f"unexpected operator name {name}")

    def test_endpoint_get_workflow_dag_result_with_metrics_and_checks(self):
        flow_id = self.flows["flow_with_metrics_and_checks"][0]
        flow = self.client.flow(flow_id)
        runs = flow.list_runs()
        resp = self.get_response(
            self.GET_WORKFLOW_RESULT_TEMPLATE % (flow_id, runs[0]["run_id"])
        ).json()
        assert_exec_state(resp["result"]["exec_state"], "succeeded")

        # operators
        operators = resp["operators"]
        assert len(operators) == 3
        for op in operators.values():
            name = op["name"]
            exec_state = op["result"]["exec_state"]
            if "query" in name or name == "size" or name == "check":  # extract
                assert_exec_state(exec_state, "succeeded")
            else:
                raise Exception(f"unexpected operator name {name}")

        # artifacts
        artifacts = resp["artifacts"]
        assert len(artifacts) == 3
        for artf in artifacts.values():
            name = artf["name"]
            exec_state = artf["result"]["exec_state"]
            value = artf["result"]["content_serialized"]

            if "query" in name:
                assert_exec_state(exec_state, "succeeded")
            elif name == "size artifact":
                assert_exec_state(exec_state, "succeeded")
                assert int(value) > 0
            elif name == "check artifact":
                assert_exec_state(exec_state, "succeeded")
                assert value == "true"
            else:
                raise Exception(f"unexpected operator name {name}")

    def test_endpoint_get_workflow_dag_result_on_flow_with_sleep(self):
        flow_id = self.running_flows["flow_with_sleep"][0]
        flow = self.client.flow(flow_id)
        runs = flow.list_runs()
        resp = self.get_response(
            self.GET_WORKFLOW_RESULT_TEMPLATE % (flow_id, runs[0]["run_id"])
        ).json()
        assert_exec_state(resp["result"]["exec_state"], "pending")

        # operators
        operators = resp["operators"]
        assert len(operators) == 2
        for op in operators.values():
            name = op["name"]
            exec_state = op["result"]["exec_state"]
            if "query" in name:  # extract
                assert_exec_state(exec_state, "succeeded")
            elif name == "sleeping_op":
                assert_exec_state(exec_state, "pending")
            else:
                raise Exception(f"unexpected operator name {name}")

        # artifacts
        artifacts = resp["artifacts"]
        assert len(artifacts) == 2
        for artf in artifacts.values():
            name = artf["name"]
            exec_state = artf["result"]["exec_state"]

            if "query" in name:
                assert_exec_state(exec_state, "succeeded")
            elif name == "sleeping_op artifact":
                assert_exec_state(exec_state, "pending")
            else:
                raise Exception(f"unexpected operator name {name}")

    def test_endpoint_list_artifact_results_with_metrics_and_checks(self):
        flow_id, num_runs = self.flows["flow_with_metrics_and_checks"]
        flow = self.client.flow(flow_id)
        runs = flow.list_runs()
        resp = self.get_response(
            self.GET_WORKFLOW_RESULT_TEMPLATE % (flow_id, runs[0]["run_id"])
        ).json()

        # artifacts
        artifacts = resp["artifacts"]
        assert len(artifacts) == 3
        for artf in artifacts.values():
            name = artf["name"]
            id = artf["id"]
            resp = self.get_response(self.LIST_ARTIFACT_RESULTS_TEMPLATE % (flow_id, id)).json()
            results = resp["results"]
            assert len(results) == num_runs

            for result in results:
                exec_state = result["exec_state"]
                value = result["content_serialized"]
                assert_exec_state(exec_state, "succeeded")

                if "query" in name:
                    assert value is None
                elif name == "size artifact":
                    assert int(value) > 0
                elif name == "check artifact":
                    assert value == "true"

    def test_endpoint_workflows_get(self):
        resp = self.get_response(self.GET_WORKFLOWS_TEMPLATE)
        resp = resp.json()

        if len(resp) > 0:
            keys = [
                "id",
                "user_id",
                "name",
                "description",
                "schedule",
                "created_at",
                "retention_policy",
                "notification_settings",
            ]

            user_id = resp[0]["user_id"]

            for v2_workflow in resp:
                for key in keys:
                    assert key in v2_workflow
                assert v2_workflow["user_id"] == user_id

    def test_endpoint_workflow_dags_get(self):
        flow_id, _ = self.flows["flow_with_metrics_and_checks"]
        resp = self.get_response(self.GET_DAGS_TEMPLATE % flow_id)
        resp = resp.json()

        assert len(resp) == 2
        assert "id" in resp[0]
        assert resp[0]["workflow_id"] == str(flow_id)
        assert "created_at" in resp[0]

    def test_endpoint_dag_results_get(self):
        flow_id, n_runs = self.flows["flow_with_metrics_and_checks"]
        resp = self.get_response(self.GET_DAG_RESULTS_TEMPLATE % flow_id).json()

        assert len(resp) == n_runs

        def check_structure(resp, all_succeeded=False):
            for result in resp:
                result = GetDagResultResponse(**result)
                if all_succeeded:
                    assert result.exec_state.status == "succeeded"
                    assert result.exec_state.failure_type == None
                    assert result.exec_state.error == None

        check_structure(resp, all_succeeded=True)

        # Using the order parameter
        flow_id, n_runs = self.flows["flow_with_failure"]
        resp = self.get_response(
            self.GET_DAG_RESULTS_TEMPLATE % flow_id + "?order_by=status",
        ).json()

        check_structure(resp)
        statuses = [result["exec_state"]["status"] for result in resp]
        sorted_statuses = sorted(statuses, reverse=True)  # Descending order
        assert statuses == sorted_statuses

        # Default is descending
        flow_id, n_runs = self.flows["flow_with_failure"]
        resp = self.get_response(
            self.GET_DAG_RESULTS_TEMPLATE % flow_id + "?order_by=status&order_descending=true",
        ).json()

        check_structure(resp)
        descending_statuses = [result["exec_state"]["status"] for result in resp]
        assert statuses == descending_statuses

        # Ascending works
        flow_id, n_runs = self.flows["flow_with_failure"]
        resp = self.get_response(
            self.GET_DAG_RESULTS_TEMPLATE % flow_id + "?order_by=status&order_descending=false",
        ).json()

        check_structure(resp)
        ascending_statuses = [result["exec_state"]["status"] for result in resp]
        assert descending_statuses[::-1] == ascending_statuses

        # Using the limit parameter
        resp = self.get_response(
            self.GET_DAG_RESULTS_TEMPLATE % flow_id + "?limit=1",
        ).json()

        check_structure(resp)
        assert len(resp) == 1

        # Using both the order and limit parameters
        resp = self.get_response(
            self.GET_DAG_RESULTS_TEMPLATE % flow_id + "?order_by=status&limit=1",
        ).json()

        check_structure(resp)
        workflow_status = [result["exec_state"]["status"] for result in resp]
        assert len(workflow_status) == 1
        workflow_status = workflow_status[0]
        assert workflow_status == sorted_statuses[0]

    def test_endpoint_nodes_get(self):
        for flow_id, _ in [
            self.flows["flow_with_metrics_and_checks"],
            self.flows["flow_with_multiple_operators"],
        ]:
            flow = self.client.flow(flow_id)
            workflow_resp = flow._get_workflow_resp()
            dag_id = list(workflow_resp.workflow_dags.keys())[0]
            resp = self.get_response(self.GET_NODES_TEMPLATE % (flow_id, dag_id)).json()

            all_output_counts = []
            for artifact in resp["operators"]:
                result = GetNodeOperatorResponse(**artifact)
                all_output_counts.append(len(result.outputs))
            assert sum(all_output_counts) == len(all_output_counts)
            assert set(all_output_counts) == set([1])

            all_output_counts = []
            for artifact in resp["artifacts"]:
                result = GetNodeArtifactResponse(**artifact)
                all_output_counts.append(len(result.outputs))
            assert sum(all_output_counts) == len(all_output_counts) - 1
            assert set(all_output_counts) == set([0, 1])

    def test_endpoint_nodes_results_get(self):
        for flow_id, _ in [
            self.flows["flow_with_metrics_and_checks"],
            self.flows["flow_with_multiple_operators"],
        ]:
            flow = self.client.flow(flow_id)
            workflow_resp = flow._get_workflow_resp()
            dag_result_id = workflow_resp.workflow_dag_results[0].id
            resp = self.get_response(
                self.GET_NODES_RESULTS_TEMPLATE % (flow_id, dag_result_id)
            ).json()
            assert "operators" in resp.keys()
            assert "artifacts" in resp.keys()
            assert len(resp["operators"]) == len(resp["artifacts"])
            for op in resp["operators"]:
                result = GetOperatorResultResponse(**op)
                result.exec_state.status == "succeeded"
            for artf in resp["artifacts"]:
                result = GetArtifactResultResponse(**artf)
                result.exec_state.status == "succeeded"

    def test_endpoint_node_artifact_get(self):
        for flow_id, _ in [
            self.flows["flow_with_metrics_and_checks"],
            self.flows["flow_with_multiple_operators"],
        ]:
            flow = self.client.flow(flow_id)
            workflow_resp = flow._get_workflow_resp()
            dag_id = workflow_resp.workflow_dag_results[0].workflow_dag_id
            dag_result_id = workflow_resp.workflow_dag_results[0].id

            dag_result_resp = globals.__GLOBAL_API_CLIENT__.get_workflow_dag_result(
                flow_id,
                dag_result_id,
            )
            artifact_ids = list(dag_result_resp.artifacts.keys())
            artifact_id = str(artifact_ids[0])
            all_output_counts = []
            for artifact_id in artifact_ids:
                artifact_id = str(artifact_id)
                resp = self.get_response(
                    self.GET_NODE_ARTIFACT_TEMPLATE % (flow_id, dag_id, artifact_id)
                ).json()
                result = GetNodeArtifactResponse(**resp)
                all_output_counts.append(len(result.outputs))
            assert sum(all_output_counts) == len(all_output_counts) - 1
            assert set(all_output_counts) == set([0, 1])

    # def test_endpoint_node_artifact_result_content_get(self):
    #     flow_id, n_runs = self.flows["flow_with_multiple_operators"]
    #     flow = self.client.flow(flow_id)
    #     workflow_resp = flow._get_workflow_resp()
    #     dag_id = workflow_resp.workflow_dag_results[0].workflow_dag_id
    #     dag_result_id = workflow_resp.workflow_dag_results[0].id

    #     dag_result_resp = globals.__GLOBAL_API_CLIENT__.get_workflow_dag_result(
    #         flow_id,
    #         dag_result_id,
    #     )
    #     artifact_ids = list(dag_result_resp.artifacts.keys())
    #     artifact_id = str(artifact_ids[0])

    #     resp = self.get_response(self.GET_NODE_ARTIFACT_RESULTS_TEMPLATE % (flow_id, dag_id, artifact_id)).json()
    #     downstream_ids = [GetArtifactResultResponse(**result).id for result in resp]
    #     for downstream_id in downstream_ids:
    #         artifact_result_id = str(downstream_id)
    #         resp = self.get_response(self.GET_NODE_ARTIFACT_RESULT_CONTENT_TEMPLATE % (flow_id, dag_id, artifact_id, artifact_result_id)).json()
    #         # One of these should be successful (direct descendent of operator)
    #         print(resp)
    #     # TODO: Investigate output
    #     # >> {"error":"Unexpected error reading DAG.\nQuery returned no rows."}

    def test_endpoint_node_artifact_results_get(self):
        for flow_id, _ in [
            self.flows["flow_with_metrics_and_checks"],
            self.flows["flow_with_multiple_operators"],
        ]:
            flow = self.client.flow(flow_id)
            workflow_resp = flow._get_workflow_resp()
            dag_id = workflow_resp.workflow_dag_results[0].workflow_dag_id
            dag_result_id = workflow_resp.workflow_dag_results[0].id

            dag_result_resp = globals.__GLOBAL_API_CLIENT__.get_workflow_dag_result(
                flow_id,
                dag_result_id,
            )
            artifact_ids = list(dag_result_resp.artifacts.keys())
            artifact_id = str(artifact_ids[0])

            resp = self.get_response(
                self.GET_NODE_ARTIFACT_RESULTS_TEMPLATE % (flow_id, dag_id, artifact_id)
            ).json()
            for result in resp:
                result = GetArtifactResultResponse(**result)

    def test_endpoint_node_operator_get(self):
        for flow_id, _ in [
            self.flows["flow_with_metrics_and_checks"],
            self.flows["flow_with_multiple_operators"],
        ]:
            flow = self.client.flow(flow_id)
            workflow_resp = flow._get_workflow_resp()
            dag_id = workflow_resp.workflow_dag_results[0].workflow_dag_id
            dag_result_id = workflow_resp.workflow_dag_results[0].id

            dag_result_resp = globals.__GLOBAL_API_CLIENT__.get_workflow_dag_result(
                flow_id,
                dag_result_id,
            )
            operator_ids = list(dag_result_resp.operators.keys())
            operator_id = str(operator_ids[0])

            resp = self.get_response(
                self.GET_NODE_OPERATOR_TEMPLATE % (flow_id, dag_id, operator_id)
            ).json()
            result = GetNodeOperatorResponse(**resp)
            assert str(result.id) == operator_id
            assert result.dag_id == dag_id

    # def test_endpoint_node_operator_content_get(self):
    #     flow_id, n_runs = self.flows["flow_with_multiple_operators"]
    #     flow = self.client.flow(flow_id)
    #     workflow_resp = flow._get_workflow_resp()
    #     dag_id = workflow_resp.workflow_dag_results[0].workflow_dag_id
    #     dag_result_id = workflow_resp.workflow_dag_results[0].id

    #     dag_result_resp = globals.__GLOBAL_API_CLIENT__.get_workflow_dag_result(
    #         flow_id,
    #         dag_result_id,
    #     )
    #     operator_ids = list(dag_result_resp.operators.keys())
    #     operator_id = str(operator_ids[0])

    #     resp = self.get_response(self.GET_NODE_OPERATOR_CONTENT_TEMPLATE % (flow_id, dag_id, operator_id))
    #     print(resp.text)

    #     # TODO: Investigate output
    #     # >> {"error":"Unexpected error reading DAG.\nQuery returned no rows."}
