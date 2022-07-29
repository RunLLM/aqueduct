import pytest
import requests
from setup.changing_saves_workflow import setup_changing_saves

import aqueduct


class TestBackend:
    GET_WORKFLOW_TABLES_TEMPLATE = "/api/workflow/%s/tables"
    GET_TEST_INTEGRATION_TEMPLATE = "/api/integration/%s/test"

    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)
        integration = cls.client.integration(name=pytest.integration)
        cls.integration = integration
        cls.flows = {"changing_saves": setup_changing_saves(cls.client, pytest.integration)}

    @classmethod
    def teardown_class(cls):
        for flow in cls.flows:
            cls.client.delete_flow(cls.flows[flow])

    @classmethod
    def get_response_class(cls, endpoint, additional_headers={}):
        headers = {"api-key": pytest.api_key}
        headers.update(additional_headers)
        url = cls.client._api_client.construct_full_url(endpoint)
        r = requests.get(url, headers=headers)
        return r

    def test_endpoint_getworkflowtables(self):
        endpoint = self.GET_WORKFLOW_TABLES_TEMPLATE % self.flows["changing_saves"]
        data = self.get_response_class(endpoint).json()["table_details"]

        assert len(data) == 3

        # table_name, update_mode
        data_set = set(
            [
                ("table_1", "append"),
                ("table_1", "replace"),
                ("table_2", "append"),
            ]
        )
        assert set([(item["table_name"], item["update_mode"]) for item in data]) == data_set

        # Check all in same integration
        assert len(set([item["integration_id"] for item in data])) == 1
        assert len(set([item["service"] for item in data])) == 1

    def test_testintegration(self):
        resp = self.get_response_class(
            self.GET_TEST_INTEGRATION_TEMPLATE % self.integration._metadata.id
        )
        assert resp.ok
