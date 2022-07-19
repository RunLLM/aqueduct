import pytest
import requests
import subprocess
from pathlib import Path
import os
import sys

import aqueduct


def get_response(endpoint, additional_headers={}):
    headers = {"api-key": pytest.api_key}
    headers.update(additional_headers)
    url = f"{pytest.server_address}{endpoint}"
    if url[:4] != "http":
        try:
           r = requests.get("https://"+url, headers=headers) 
        except:
            try:
                r = requests.get("http://"+url, headers=headers) 
            except:
                raise Exception(f"Cannot connect to {url}")
    else:
        try:
            r = requests.get(url, headers=headers)
        except:
            raise Exception(f"Cannot connect to {url}")
    return r

class TestBackend:
    GET_WORKFLOW_TABLES_TEMPLATE = "/api/workflow/%s/tables"
    WORKFLOW_PATH = Path(__file__).parent / "setup"
    
    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)
        cls.flows = {}

        workflow_files = [f for f in os.listdir(cls.WORKFLOW_PATH) if os.path.isfile(os.path.join(cls.WORKFLOW_PATH, f))]
        for workflow in workflow_files:
            proc = subprocess.Popen(["python3", os.path.join(cls.WORKFLOW_PATH, workflow), pytest.api_key, pytest.server_address],
                                    stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            out, err = proc.communicate()
            out = out.decode("utf-8") 
            err = err.decode("utf-8") 
            if err:
                raise Exception(f"Could not run workflow {workflow}.\n\n{err}")
            else:
                cls.flows[workflow] = out.strip().split()[-1]

    @classmethod
    def teardown_class(cls):
        for flow in cls.flows:
            cls.client.delete_flow(cls.flows[flow])

    def test_endpoint_getworkflowtables(self):
        endpoint = self.GET_WORKFLOW_TABLES_TEMPLATE % self.flows["changing_saves.py"]
        data = get_response(endpoint).json()["table_details"]
        print(data)

    #     expected_table_names_update_modes = defaultdict(int)
    #     for table_name, update_mode in zip(
    #         self.changing_saves_flow_table_names,
    #         self.changing_saves_flow_update_modes,
    #     ):
    #         # Should de-dup exact duplicates.
    #         expected_table_names_update_modes[(table_name, update_mode)] = 1

    #     # Should contain all except for exact duplicates
    #     n_saves = len(data)
    #     assert n_saves == len(expected_table_names_update_modes.keys())

    #     # Check structure, values
    #     actual_integration_ids = defaultdict(int)
    #     actual_services = defaultdict(int)
    #     actual_table_names_update_modes = defaultdict(int)
    #     for details in data:
    #         assert set(details.keys()) == set(
    #             ["name", "integration_id", "service", "table_name", "update_mode"]
    #         )
    #         actual_integration_ids[details["integration_id"]] += 1
    #         actual_services[details["service"]] += 1
    #         actual_table_names_update_modes[(details["table_name"], details["update_mode"])] += 1

    #     assert len(actual_integration_ids) == 1
    #     assert actual_integration_ids[list(actual_integration_ids.keys())[0]] == n_saves

    #     assert len(actual_services) == 1
    #     assert actual_services[list(actual_services.keys())[0]] == n_saves

    #     assert len(actual_table_names_update_modes) == len(expected_table_names_update_modes)
    #     for key in expected_table_names_update_modes.keys():
    #         assert key in actual_table_names_update_modes
    #         assert actual_table_names_update_modes[key] == expected_table_names_update_modes[key]
