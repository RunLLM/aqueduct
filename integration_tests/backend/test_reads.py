import os
import subprocess
import sys
from pathlib import Path
from time import sleep

import pytest
import requests

import aqueduct


class TestReads:
    GET_WORKFLOW_TABLES_TEMPLATE = "/api/workflow/%s/objects"
    WORKFLOW_PATH = Path(__file__).parent / "setup"

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

    def test_endpoint_get_workflow_tables(self):
        endpoint = self.GET_WORKFLOW_TABLES_TEMPLATE % self.flows["changing_saves.py"]
<<<<<<< HEAD
        data = self.get_response_class(endpoint).json()["object_details"]
=======
        data = self.get_response_class(endpoint)
      
        data = data.json()["object_details"]
>>>>>>> 5a64ce5 (Rename)

        assert len(data) == 3
        
        # table_name, update_mode
        data_set = set(
            [
                ("table_1", "append"),
                ("table_1", "replace"),
                ("table_2", "append"),
            ]
        )
        assert set([(item["object_name"], item["update_mode"]) for item in data]) == data_set

        # Check all in same integration
        assert len(set([item["integration_id"] for item in data])) == 1
        assert len(set([item["service"] for item in data])) == 1

    def test_sdk_get_workflow_tables(self):
        data = self.client.get_workflow_writes(self.flows["changing_saves.py"])

        # Check all in same integration
        assert len(data.keys()) == 1

        # table_name, update_mode
        data_set = set(
            [
                ("table_1", "append"),
                ("table_1", "replace"),
                ("table_2", "append"),
            ]
        )
        integration_id = list(data.keys())[0]
        assert len(data[integration_id]) == 3
        assert set([(item.name, item.update_mode) for item in data[integration_id]]) == data_set
<<<<<<< HEAD
=======
    
    def test_sdk_delete_workflow_invalid(self):
        tables = self.client.get_workflow_writes(self.flows["changing_saves.py"])
        integration_id = list(tables.keys())[0]
        tables[integration_id][0].name = 'I_DON_T_EXIST'
        tables[integration_id] = [tables[integration_id][0]]
       
        with pytest.raises(Exception) as e_info:
            data = self.client.delete_flow(self.flows["changing_saves.py"], writes_to_delete=tables, force=True)
>>>>>>> 5a64ce5 (Rename)
