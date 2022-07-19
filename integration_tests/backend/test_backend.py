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
    r = requests.get(url, headers=headers)
    return r

class TestBackend:
    GET_WORKFLOW_TABLES_TEMPLATE = "/api/workflow/%s/tables"
    WORKFLOW_PATH = Path(__file__).parent / "setup"
    
    @classmethod
    def setup_class(cls):
        if pytest.server_address.endswith('/'):
            pytest.server_address = pytest.server_address[:-1]

        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)

        if not pytest.server_address.startswith('http'):
            if cls.client._api_client.use_https:
                pytest.server_address = 'https://'+pytest.server_address
            else:
                pytest.server_address = 'http://'+pytest.server_address
        
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
        return
        endpoint = self.GET_WORKFLOW_TABLES_TEMPLATE % self.flows["changing_saves.py"]
        data = get_response(endpoint).json()["table_details"]

        assert len(data) == 3

        # table_name, update_mode
        data_set = set([
            ('table_1', 'append'),
            ('table_1', 'replace'),
            ('table_2', 'append'),
        ])
        assert set([(item['table_name'], item['update_mode'])for item in data]) == data_set

        # Check all in same integration
        assert len(set([item['integration_id'] for item in data])) == 1
        assert len(set([item['service'] for item in data])) == 1
