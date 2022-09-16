package lambda

type LambdaFunctionType string

const (
	FunctionExecutor37Type LambdaFunctionType = "function37"
	FunctionExecutor38Type LambdaFunctionType = "function38"
	FunctionExecutor39Type LambdaFunctionType = "function39"
	ParamExecutorType      LambdaFunctionType = "param"
	SystemMetricType       LambdaFunctionType = "system-metric"

	AthenaConnectorType    LambdaFunctionType = "athena-connector"
	BigQueryConnectorType  LambdaFunctionType = "bigquery-connector"
	PostgresConnectorType  LambdaFunctionType = "postgres-connector"
	SnowflakeConnectorType LambdaFunctionType = "snowflake-connector"
	S3ConnectorType        LambdaFunctionType = "s3-connector"
)
