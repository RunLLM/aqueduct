#!/bin/bash
# NOTE: Keep this in sync with the `start-function-executor.sh` in `/src/python/aqueduct_executor`.

# Get the version number using `aqueduct`
aqueduct_version=$(aqueduct version | tr -d '\n')

# Get the expected version number from the environment variable
expected_version=$AQUEDUCT_EXPECTED_VERSION

# If expected version is not empty, compare version numbers
if [ -n "$expected_version" ]; then
  echo "Comparing aqueduct version ($aqueduct_version) to expected version ($expected_version)"
  
  # Check if the version numbers match
  if [ "$aqueduct_version" != "$expected_version" ]; then
    echo "Error: aqueduct version ($aqueduct_version) does not match expected version ($expected_version)"
    exit 1
  fi
fi

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
