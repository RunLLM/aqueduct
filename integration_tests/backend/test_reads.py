import os
import subprocess
import sys
from pathlib import Path
from time import sleep

import pytest
import requests
import utils
from setup.changing_saves_workflow import setup_changing_saves

import aqueduct


class TestBackend:
    LIST_WORKFLOW_SAVED_OBJECTS_TEMPLATE = "/api/workflow/%s/objects"
    GET_TEST_INTEGRATION_TEMPLATE = "/api/integration/%s/test"

    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)
        cls.integration = cls.client.integration(name=pytest.integration)
        cls.flows = {"changing_saves": setup_changing_saves(cls.client, pytest.integration)}
        for flow in cls.flows.values():
            utils.wait_for_flow_runs(cls.client, flow, 4)

    @classmethod
    def teardown_class(cls):
        for flow in cls.flows:
            utils.delete_flow(cls.client, cls.flows[flow])

    @classmethod
    def get_response_class(cls, endpoint, additional_headers={}):
        headers = {"api-key": pytest.api_key}
        headers.update(additional_headers)
        url = aqueduct.api_client.__GLOBAL_API_CLIENT__.construct_full_url(endpoint)
        r = requests.get(url, headers=headers)
        return r

    def test_endpoint_list_saved_objects(self):
        endpoint = self.LIST_WORKFLOW_SAVED_OBJECTS_TEMPLATE % self.flows["changing_saves"]
        data = self.get_response_class(endpoint).json()["object_details"]

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
        assert len(set([item["integration_name"] for item in data])) == 1
        assert len(set([item["service"] for item in data])) == 1

    def test_endpoint_test_integration(self):
        resp = self.get_response_class(
            self.GET_TEST_INTEGRATION_TEMPLATE % self.integration._metadata.id
        )
        assert resp.ok
