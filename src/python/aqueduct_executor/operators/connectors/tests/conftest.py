import os
import sys

_REQUIREMENTS_FILE = "./python/aqueduct_executor/operators/connectors/tests/requirements.txt"


def pytest_configure(config):
    # Install required packages
    os.system(f"{sys.executable} -m pip install -r {_REQUIREMENTS_FILE}")
