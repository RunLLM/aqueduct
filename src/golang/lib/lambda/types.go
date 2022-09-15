package lambda

type LambdaFunctionType string

const (
	FunctionExecutorType LambdaFunctionType = "function"
	ParamExecutorType    LambdaFunctionType = "param"
	SystemMetricType     LambdaFunctionType = "system-metric"

	AthenaConnectorType    LambdaFunctionType = "athena-connector"
	BigQueryConnectorType  LambdaFunctionType = "bigquery-connector"
	PostgresConnectorType  LambdaFunctionType = "postgres-connector"
	SnowflakeConnectorType LambdaFunctionType = "snowflake-connector"
	S3ConnectorType        LambdaFunctionType = "s3-connector"
)
