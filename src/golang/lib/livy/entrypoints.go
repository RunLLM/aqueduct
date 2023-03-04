package livy

const (
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

spec_json = base64.b64decode("%s")
spec = parse_spec(spec_json)
extract_function.run(spec)
spark_session_obj = spark
run(spec, spark_session_obj)
`

	ParamEntrypoint = `import argparse
import base64

from aqueduct_executor.operators.param_executor import execute
from aqueduct_executor.operators.param_executor.spec import parse_spec

spec_json = base64.b64decode("%s")
spec = parse_spec(spec_json)

execute.run(spec)
`

	SystemMetricEntrypoint = `import argparse
import base64

from aqueduct_executor.operators.system_metric_executor import execute
from aqueduct_executor.operators.system_metric_executor.spec import parse_spec

spec_json = base64.b64decode("%s")
spec = parse_spec(spec_json)

execute.run(spec)

`

	DataEntrypoint = `import argparse
import base64

from aqueduct_executor.operators.spark.execute_data import run
from aqueduct_executor.operators.connectors.data.spec import parse_spec

spec_json = base64.b64decode("%s")
spec = parse_spec(spec_json)
spark_session_obj = spark

run(spec, spark_session_obj)
`
)
