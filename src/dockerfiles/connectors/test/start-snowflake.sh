#!/bin/bash

git clone https://github.com/aqueducthq/aqueduct.git
cd aqueduct
git checkout -t origin/eng-1510-add-k8s-engine-integration

cd src/python

pip install .

python3 -m aqueduct_executor.operators.connectors.tabular.main --spec "$JOB_SPEC"
