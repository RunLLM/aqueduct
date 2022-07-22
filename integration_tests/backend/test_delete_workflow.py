import os
import subprocess
import sys
from pathlib import Path

import pytest
import requests

import aqueduct


class TestDeleteWorkflow:
    GET_WORKFLOW_TABLES_TEMPLATE = "/api/workflow/%s/tables"
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

    @classmethod
    def teardown_class(cls):
        for flow in cls.flows:
            cls.client.delete_flow(cls.flows[flow])
    
    def test_sdk_deleteworkflow_invalid(self):
        tables = self.client.get_workflow_writes(self.flows["changing_saves.py"])
        integration_id = list(tables.keys())[0]
        tables[integration_id][0].name = 'I_DON_T_EXIST'
        tables[integration_id] = [tables[integration_id][0]]
       
        with pytest.raises(Exception) as e_info:
            data = self.client.delete_flow(self.flows["changing_saves.py"], writes_to_delete=tables, force=True)

    def test_delete_workflow(self):
        tables = self.client.get_workflow_writes(self.flows["changing_saves.py"])
        integration_id = list(tables.keys())[0]
        tables[integration_id][0].name = 'I_DON_T_EXIST'
        tables[integration_id] = [tables[integration_id][0]]
       
        with pytest.raises(Exception) as e_info:
            data = self.client.delete_flow(self.flows["changing_saves.py"], writes_to_delete=tables, force=True)
