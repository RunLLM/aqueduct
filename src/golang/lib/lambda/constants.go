package lambda

const (
	FunctionLambdaFunction     = "aqueduct-function"
	ParameterLambdaFunction    = "aqueduct-parameter"
	SystemMetricLambdaFunction = "aqueduct-systemmetric"

	AthenaLambdaFunction    = "aqueduct-athena"
	BigQueryLambdaFunction  = "aqueduct-bigquery"
	PostgresLambdaFunction  = "aqueduct-postgres"
	S3LambdaFunction        = "aqueduct-s3"
	SnowflakeLambdaFunction = "aqueduct-snowflake"

	FunctionLambdaImage     = "aqueducthq/lambda-function:0.0.13"
	ParameterLambdaImage    = "aqueducthq/lambda-param:0.0.13"
	SystemMetricLambdaImage = "aqueducthq/lambda-system-metric:0.0.13"

	AthenaConnectorLambdaImage    = "aqueducthq/lambda-athena-connector:0.0.13"
	BigQueryConnectorLambdaImage  = "aqueducthq/lambda-bigquery-connector:0.0.13"
	PostgresConnectorLambdaImage  = "aqueducthq/lambda-postgres-connector:0.0.13"
	S3ConnectorLambdaImage        = "aqueducthq/lambda-s3-connector:0.0.13"
	SnowflakeConnectorLambdaImage = "aqueducthq/lambda-snowflake-connector:0.0.13"
)
