import os
import subprocess
import sys
from pathlib import Path
from time import sleep

import pytest
import requests

import aqueduct


class TestDeleteWorkflow:
    LIST_WORKFLOW_SAVED_OBJECTS_TEMPLATE = "/api/workflow/%s/objects"
    INTEGRATION_OBJECTS_TEMPLATE = "/api/integration/%s/objects"
    WORKFLOW_PATH = Path(__file__).parent / "setup"

    @classmethod
    def setup_class(cls):
        cls.client = aqueduct.Client(pytest.api_key, pytest.server_address)
        cls.flows = {}

        workflow_files = ['simple_saves.py']
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
                cls.flows[workflow] = out.strip().split()[-1]
            sleep(10)

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

    def test_delete_workflow(self):
        endpoint = self.INTEGRATION_OBJECTS_TEMPLATE % integration_id

        tables_response = self.get_response_class(endpoint).json()
        assert 'delete_table' in set(tables_response['table_names'])
       
        with pytest.raises(InvalidRequestError) as e_info:
            self.client.delete_flow(self.flows["simple_saves.py"], writes_to_delete=tables, force=False)
        print(e_info)
        data = self.client.delete_flow(self.flows["simple_saves.py"], writes_to_delete=tables, force=True)
        # Wait for deletion to occur
        sleep(1)

        tables_response = self.get_response_class(endpoint).json()
        assert 'delete_table' not in set(tables_response['table_names'])

        del self.flows["simple_saves.py"]
   
        assert len(data) == 1
        assert len(data[integration_id]) == 1
        assert data[integration_id][0].succeeded == True