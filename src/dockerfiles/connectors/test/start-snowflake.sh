#!/bin/bash

git clone https://github.com/aqueducthq/aqueduct.git
git checkout -t origin/eng-1510-add-k8s-engine-integration

cd aqueduct/src/python

python3 setup.py

python3 -m aqueduct_executor.operators.connectors.tabular.main --spec "$JOB_SPEC"
