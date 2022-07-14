import pytest
import requests
from collections import defaultdict
import aqueduct
from setup.load_workflow import create_test_endpoint_GetWorkflowTables_flow

class TestBackend:
    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)
        cls.flows = []

        # For test_endpoint_GetWorkflowTables
        cls.test_endpoint_GetWorkflowTables_flow_table_names = ["table_1", "table_1", "table_1", "table_2"]
        cls.test_endpoint_GetWorkflowTables_flow_update_modes = ["append", "append", "replace", "append"]
        cls.test_endpoint_GetWorkflowTables_flow = create_test_endpoint_GetWorkflowTables_flow(
            cls.client, 
            "test_endpoint_GetWorkflowTables flow",
            cls.test_endpoint_GetWorkflowTables_flow_table_names,
            cls.test_endpoint_GetWorkflowTables_flow_update_modes 
        )
        cls.flows.append(cls.test_endpoint_GetWorkflowTables_flow)

    @classmethod
    def teardown_class(cls):
        for flow in cls.flows:
            cls.client.delete_flow(flow.id())

    def test_endpoint_GetWorkflowTables(self):
        headers = {
            "api-key": pytest.api_key
        }
        url = f"{pytest.adapter}{pytest.server_address}/api/workflow/{self.test_endpoint_GetWorkflowTables_flow.id()}/tables"
        r = requests.get(url, headers=headers)
        data = r.json()["table_details"]

        expected_table_names_update_modes = defaultdict(int)
        for table_name, update_mode in zip(
            self.test_endpoint_GetWorkflowTables_flow_table_names, 
            self.test_endpoint_GetWorkflowTables_flow_update_modes
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
            assert set(details.keys()) == set(['name', 'integration_id', 'service', 'table_name', 'update_mode'])
            actual_integration_ids[details['integration_id']] += 1
            actual_services[details['service']] += 1
            actual_table_names_update_modes[(details['table_name'], details['update_mode'])] += 1
        
        assert len(actual_integration_ids) == 1
        assert actual_integration_ids[list(actual_integration_ids.keys())[0]] == n_saves

        assert len(actual_services) == 1
        assert actual_services[list(actual_services.keys())[0]] == n_saves

        assert len(actual_table_names_update_modes) == len(expected_table_names_update_modes)
        for key in expected_table_names_update_modes.keys():
            assert key in actual_table_names_update_modes
            assert actual_table_names_update_modes[key] == expected_table_names_update_modes[key]
