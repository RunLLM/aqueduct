from collections import defaultdict

import pytest
import requests
from setup.load_workflow import create_test_endpoint_getworkflowtables_flow

import aqueduct


def get_response(endpoint, additional_headers={}):
    headers = {"api-key": pytest.api_key}
    headers.update(additional_headers)
    url = f"{pytest.adapter}{pytest.server_address}{endpoint}"
    r = requests.get(url, headers=headers)
    return r

class TestBackend:
    GET_WORKFLOW_TABLES_TEMPLATE = "/api/workflow/%s/tables"
    
    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)
        cls.flows = []

        # Workflow that loads a table from the `aqueduct_demo` then saves it to `table_1` in append mode.
        # This save operator is then replaced by one that saves to `table_1` in replace mode.
        # In the next deployment of this run, it saves to `table_1` in append mode.
        # In the last deployment, it saves to `table_2` in append mode.
        cls.changing_saves_flow_table_names = [
            "table_1",
            "table_1",
            "table_1",
            "table_2",
        ]
        cls.changing_saves_flow_update_modes = [
            "append",
            "replace",
            "append",
            "append",
        ]
        cls.changing_saves_flow = create_changing_saves_flow(
            cls.client,
            "changing saves flow",
            cls.changing_saves_flow_table_names,
            cls.changing_saves_flow_update_modes,
        )
        cls.flows.append(cls.changing_saves_flow)

    @classmethod
    def teardown_class(cls):
        for flow in cls.flows:
            cls.client.delete_flow(flow.id())

    def test_endpoint_getworkflowtables(self):
        endpoint = self.GET_WORKFLOW_TABLES_TEMPLATE % self.changing_saves_flow.id()
        data = get_response(endpoint).json()["table_details"]

        expected_table_names_update_modes = defaultdict(int)
        for table_name, update_mode in zip(
            self.changing_saves_flow_table_names,
            self.changing_saves_flow_update_modes,
        ):
            # Should de-dup exact duplicates.
            expected_table_names_update_modes[(table_name, update_mode)] = 1

        # Should contain all except for exact duplicates
        n_saves = len(data)
        assert n_saves == len(expected_table_names_update_modes.keys())

        # Check structure, values
        actual_integration_ids = defaultdict(int)
        actual_services = defaultdict(int)
        actual_table_names_update_modes = defaultdict(int)
        for details in data:
            assert set(details.keys()) == set(
                ["name", "integration_id", "service", "table_name", "update_mode"]
            )
            actual_integration_ids[details["integration_id"]] += 1
            actual_services[details["service"]] += 1
            actual_table_names_update_modes[(details["table_name"], details["update_mode"])] += 1

        assert len(actual_integration_ids) == 1
        assert actual_integration_ids[list(actual_integration_ids.keys())[0]] == n_saves

        assert len(actual_services) == 1
        assert actual_services[list(actual_services.keys())[0]] == n_saves

        assert len(actual_table_names_update_modes) == len(expected_table_names_update_modes)
        for key in expected_table_names_update_modes.keys():
            assert key in actual_table_names_update_modes
            assert actual_table_names_update_modes[key] == expected_table_names_update_modes[key]
