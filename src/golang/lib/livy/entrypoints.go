package livy

const (
	InstallAqueductEntrypoint = `import sys
import subprocess

print(1)
subprocess.check_call([sys.executable, '-m', 'pip', 'install', 'aqueduct-ml==0.2.1'])
`

	FunctionEntrypoint = `import base64
import argparse
import subprocess
import sys

from aqueduct_executor.operators.spark.execute_function import run
from aqueduct_executor.operators.function_executor import (
	extract_function,
	install_requirements,
)
from aqueduct_executor.operators.function_executor.spec import parse_spec


def pip_freeze(local_deps_path):
	subprocess.run([sys.executable, "-m", "pip", "freeze", ">>", local_deps_path])


def main():
	"""
	1. extract function
	2. download required packages
	3. execute function.
	"""

	spec_json = base64.b64decode("%s")
	spec = parse_spec(spec_json)

	extract_function.run(spec)
	open(spec.function_extract_path + "op/local_deps.txt", 'w')
	open(spec.function_extract_path + "op/missing.txt", 'w')
	pip_freeze(spec.function_extract_path + "op/local_deps.txt")
	install_requirements.run(
		spec.function_extract_path + "op/local_deps.txt",
		spec.function_extract_path + "op/requirements.txt",
		spec.function_extract_path + "op/missing.txt",
		spec,
	)
	spark_session_obj = spark
	run(spec, spark_session_obj)


if __name__ == "__main__":
	main()

`
	ParamEntrypoint = `import argparse
import base64

from aqueduct_executor.operators.param_executor import execute
from aqueduct_executor.operators.param_executor.spec import parse_spec

if __name__ == "__main__":

	spec_json = base64.b64decode("%s")
	spec = parse_spec(spec_json)

	execute.run(spec)
`

	SystemMetricEntrypoint = `import argparse
import base64

from aqueduct_executor.operators.system_metric_executor import execute
from aqueduct_executor.operators.system_metric_executor.spec import parse_spec

if __name__ == "__main__":

	spec_json = base64.b64decode("%s")
	spec = parse_spec(spec_json)

	execute.run(spec)

`

	DataEntrypoint = `import argparse
import base64

from aqueduct_executor.operators.spark.execute_data import run
from aqueduct_executor.operators.connectors.data.spec import parse_spec

if __name__ == "__main__":

	spec_json = base64.b64decode("%s")
	spec = parse_spec(spec_json)
	spark_session_obj = spark

	run(spec, spark_session_obj)
`
)
