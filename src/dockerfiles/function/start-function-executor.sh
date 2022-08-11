#!/bin/bash
OP_PATH=$(python3 -m aqueduct_executor_enterprise.operators.function_executor.get_extract_path --spec "$JOB_SPEC")
python3 -m aqueduct_executor_enterprise.operators.function_executor.extract_function --spec "$JOB_SPEC"

PYTHON_VERSION=$(python3 -m aqueduct_executor_enterprise.operators.function_executor.set_conda_version "$OP_PATH")
echo "Python version is $PYTHON_VERSION"

if test -f "$OP_PATH/requirements.txt"; then conda run -n $PYTHON_VERSION python3 -m pip install -r "$OP_PATH/requirements.txt" --no-cache-dir; fi

conda run -n $PYTHON_VERSION python3 -m aqueduct_executor_enterprise.operators.function_executor.execute_function --spec "$JOB_SPEC"
