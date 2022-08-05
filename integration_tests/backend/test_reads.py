import json
import os
import subprocess
import sys
import uuid
from pathlib import Path

import pytest
import requests
import utils
from setup.changing_saves_workflow import setup_changing_saves

import aqueduct


class TestBackend:
    LIST_WORKFLOW_SAVED_OBJECTS_TEMPLATE = "/api/workflow/%s/objects"
    GET_TEST_INTEGRATION_TEMPLATE = "/api/integration/%s/test"
    LIST_INTEGRATIONS_TEMPLATE = "/api/integrations"
    CONNECT_INTEGRATION_TEMPLATE = "/api/integration/connect"
    DELETE_INTEGRATION_TEMPLATE = "/api/integration/%s/delete"
    WORKFLOW_PATH = Path(__file__).parent / "setup"
    DEMO_DB_PATH = os.path.join(os.environ["HOME"], ".aqueduct/server/db/demo.db")

    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)
        cls.integration = cls.client.integration(name=pytest.integration)
        cls.flows = {"changing_saves": setup_changing_saves(cls.client, pytest.integration)}
        for flow, n_runs in cls.flows.values():
            utils.wait_for_flow_runs(cls.client, flow, n_runs)

    @classmethod
    def teardown_class(cls):
        for flow, _ in cls.flows.values():
            utils.delete_flow(cls.client, flow)

    @classmethod
    def response(cls, endpoint, additional_headers):
        headers = {"api-key": pytest.api_key}
        headers.update(additional_headers)
        url = aqueduct.api_client.__GLOBAL_API_CLIENT__.construct_full_url(endpoint)
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
        assert set([(item["object_name"], item["update_mode"]) for item in data]) == data_set

        # Check all in same integration
        assert len(set([item["integration_id"] for item in data])) == 1
        assert len(set([item["service"] for item in data])) == 1

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
        resp = self.get_response(self.GET_TEST_INTEGRATION_TEMPLATE % self.integration._metadata.id)
        assert resp.ok
