#!/bin/bash

git clone https://github.com/aqueducthq/aqueduct.git
cd aqueduct/src/python

pip install .

python3 -m aqueduct_executor.operators.connectors.data.main --spec "$JOB_SPEC"
