#!/bin/bash
# NOTE: Keep this in sync with the `start-function-executor.sh` in `/src/dockerfiles/function`.

JOB_SPEC=$1
FUNCTION_EXTRACT_PATH=$(python3 -m aqueduct_executor.operators.function_executor.get_extract_path --spec "$JOB_SPEC")
EXIT_CODE=$?
if [ $EXIT_CODE != "0" ]; then exit $(($EXIT_CODE)); fi

python3 -m aqueduct_executor.operators.function_executor.extract_function --spec "$JOB_SPEC"
EXIT_CODE=$?
if [ $EXIT_CODE != "0" ]; then exit $(($EXIT_CODE)); fi

if test -f "$FUNCTION_EXTRACT_PATH/op/requirements.txt"
then
      python3 -m pip freeze >> "$FUNCTION_EXTRACT_PATH/op/local_deps.txt"
      python3 -m aqueduct_executor.operators.function_executor.install_requirements --local_path="$FUNCTION_EXTRACT_PATH/op/local_deps.txt" --requirements_path="$FUNCTION_EXTRACT_PATH/op/requirements.txt" --missing_path="$FUNCTION_EXTRACT_PATH/op/missing.txt" --spec "$JOB_SPEC"
      EXIT_CODE=$?
      if [ $EXIT_CODE != "0" ]; then exit $(($EXIT_CODE)); fi
fi

python3 -m aqueduct_executor.operators.function_executor.main --spec "$JOB_SPEC"
EXIT_CODE=$?

# Exit after cleanup, regardless of execution success / failure.
exit $(($EXIT_CODE))
