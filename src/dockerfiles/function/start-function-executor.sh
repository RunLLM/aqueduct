#!/bin/bash
#!/bin/bash
echo $JOB_SPEC
FUNCTION_EXTRACT_PATH=$(python3 -m aqueduct_executor.operators.function_executor.get_extract_path --spec "$JOB_SPEC")
EXIT_CODE=$?
if [ $EXIT_CODE != "0" ]; then exit $(($EXIT_CODE)); fi

python3 -m aqueduct_executor.operators.function_executor.extract_function --spec "$JOB_SPEC"
EXIT_CODE=$?
if [ $EXIT_CODE != "0" ]; then exit $(($EXIT_CODE)); fi

if test -f "$FUNCTION_EXTRACT_PATH/op/requirements.txt"
then
      pip freeze >> "$FUNCTION_EXTRACT_PATH/op/local_deps.txt"
      python3 -m aqueduct_executor.operators.function_executor.prune_requirements --local_path="$FUNCTION_EXTRACT_PATH/op/local_deps.txt" --requirements_path="$FUNCTION_EXTRACT_PATH/op/requirements.txt" --missing_path="$FUNCTION_EXTRACT_PATH/op/missing.txt"
      EXIT_CODE=$?
      if [ $EXIT_CODE != "0" ]; then exit $(($EXIT_CODE)); fi
      if test -f "$FUNCTION_EXTRACT_PATH/op/missing.txt"
      then
            pip3 install -r "$FUNCTION_EXTRACT_PATH/op/missing.txt" --no-cache-dir
      fi
fi

python3 -m aqueduct_executor.operators.function_executor.main --spec "$JOB_SPEC"
EXIT_CODE=$?

# Exit after cleanup, regardless of execution success / failure.
exit $(($EXIT_CODE))
