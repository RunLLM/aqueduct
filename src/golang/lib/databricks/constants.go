package databricks

const (
	SparkVersion         = "10.4.x-scala2.12"
	DefaultMinNumWorkers = 1
	DefaultMaxNumWorkers = DefaultMinNumWorkers + 10
	DefaultNodeTypeID    = "m5d.large"

	DatabricksFunctionScript = "aqscript.py"
	DatabricksParamScript    = "paramScript.py"
	DatabricksMetricScript   = "metricScript.py"
	DatabricksDataScript     = "dataScript.py"
)
