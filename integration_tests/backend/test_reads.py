import json
import os
import subprocess
import sys
import uuid
from pathlib import Path

import pytest
import requests
import utils

import aqueduct


class TestReads:
    LIST_WORKFLOW_SAVED_OBJECTS_TEMPLATE = "/api/workflow/%s/objects"
    LIST_INTEGRATIONS_TEMPLATE = "/api/integrations"
    CONNECT_INTEGRATION_TEMPLATE = "/api/integration/connect"
    DELETE_INTEGRATION_TEMPLATE = "/api/integration/%s/delete"
    WORKFLOW_PATH = Path(__file__).parent / "setup"
    DEMO_DB_PATH = os.path.join(os.environ["HOME"], ".aqueduct/server/db/demo.db")

    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)
        cls.flows = {}

        workflow_files = [
            f
            for f in os.listdir(cls.WORKFLOW_PATH)
            if os.path.isfile(os.path.join(cls.WORKFLOW_PATH, f))
        ]
        for workflow in workflow_files:
            proc = subprocess.Popen(
                [
                    "python3",
                    os.path.join(cls.WORKFLOW_PATH, workflow),
                    pytest.api_key,
                    pytest.server_address,
                ],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
            out, err = proc.communicate()
            out = out.decode("utf-8")
            err = err.decode("utf-8")
            if err:
                raise Exception(f"Could not run workflow {workflow}.\n\n{err}")
            else:
                parsed = out.strip().split()
                cls.flows[workflow] = parsed[-2]
                n_runs = int(parsed[-1])
                utils.wait_for_flow_runs(cls.client, cls.flows[workflow], n_runs)

    @classmethod
    def teardown_class(cls):
        for flow in cls.flows:
            utils.delete_flow(cls.client, cls.flows[flow])

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
        endpoint = self.LIST_WORKFLOW_SAVED_OBJECTS_TEMPLATE % self.flows["changing_saves.py"]
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
