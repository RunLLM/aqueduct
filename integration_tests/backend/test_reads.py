import pytest
import requests
import utils
from setup.changing_saves_workflow import setup_changing_saves

import aqueduct


class TestBackend:
    GET_WORKFLOW_TABLES_TEMPLATE = "/api/workflow/%s/objects"

    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)
        cls.flows = {"changing_saves": setup_changing_saves(cls.client)}
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

    def test_endpoint_getworkflowtables(self):
        endpoint = self.GET_WORKFLOW_TABLES_TEMPLATE % self.flows["changing_saves"]
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
        assert len(set([item["integration_id"] for item in data])) == 1
        assert len(set([item["service"] for item in data])) == 1
