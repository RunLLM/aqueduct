package shared

import (
	"github.com/dropbox/godropbox/errors"
)

// Service specifies the name of the integration.
type Service string

// Supported integrations
const (
	Postgres     Service = "Postgres"
	Snowflake    Service = "Snowflake"
	MySql        Service = "MySQL"
	Redshift     Service = "Redshift"
	MariaDb      Service = "MariaDB"
	SqlServer    Service = "SQL Server"
	BigQuery     Service = "BigQuery"
	GoogleSheets Service = "Google Sheets"
	Salesforce   Service = "Salesforce"
	S3           Service = "S3"
	Github       Service = "Github"
	Sqlite       Service = "SQLite"
	Airflow      Service = "Airflow"
	Kubernetes   Service = "Kubernetes"
	GCS          Service = "GCS"
	Athena       Service = "Athena"
	Lambda       Service = "Lambda"
	MongoDB      Service = "MongoDB"
	Conda        Service = "Conda"
	Databricks   Service = "Databricks"
	Email        Service = "Email"
	Slack        Service = "Slack"
	Spark        Service = "Spark"

	// Cloud integrations
	AWS Service = "AWS"

	// Container registry integrations
	ECR Service = "ECR"

	// Built-in integrations
	Aqueduct     Service = "Aqueduct"
	AqueductDemo Service = "Aqueduct Demo"

	AqueductComputeIntegrationName = "Aqueduct Server"
	DemoDbIntegrationName          = "aqueduct_demo"
)

var relationalDatabaseIntegrations map[Service]bool = map[Service]bool{
	Postgres:     true,
	Snowflake:    true,
	MySql:        true,
	Redshift:     true,
	MariaDb:      true,
	SqlServer:    true,
	BigQuery:     true,
	AqueductDemo: true,
	Sqlite:       true,
	Athena:       true,
	MongoDB:      true,
}

var computeIntegrations map[Service]bool = map[Service]bool{
	Airflow:    true,
	Lambda:     true,
	Conda:      true,
	Databricks: true,
	Kubernetes: true,
	Spark:      true,
	AWS:        true,
}

// ServiceToEngineConfigField contains
// all services with `integration_id` in its 'engine_config' field.
// This is used in SQL queries to retrieve engine configs (workflow or operator)
// based on integration ID.
//
// The key should be the service type, and value should be the json tag
// for the corresponding field that contains the integration ID.
var ServiceToEngineConfigField map[Service]string = map[Service]string{
	Lambda:     "lambda_config",
	Airflow:    "airflow_config",
	Kubernetes: "k8s_config",
	Databricks: "databricks_config",
}

// ParseService decodes s into a Service or an error.
func ParseService(s string) (Service, error) {
	svc := Service(s)
	switch svc {
	case Postgres,
		Snowflake,
		MySql,
		Redshift,
		MariaDb,
		SqlServer,
		BigQuery,
		GoogleSheets,
		Salesforce,
		S3,
		Athena,
		AqueductDemo,
		Github,
		Sqlite,
		Airflow,
		Kubernetes,
		GCS,
		Lambda,
		MongoDB,
		Conda,
		Databricks,
		Email,
		Slack,
		Spark,
		AWS,
		ECR:
		return svc, nil
	default:
		return "", errors.Newf("Unknown service: %s", s)
	}
}

func IsRelationalDatabaseIntegration(service Service) bool {
	_, ok := relationalDatabaseIntegrations[service]
	return ok
}

func IsDatabaseIntegration(service Service) bool {
	if IsRelationalDatabaseIntegration(service) {
		return true
	}

	return service == MongoDB
}

func IsComputeIntegration(service Service) bool {
	_, ok := computeIntegrations[service]
	return ok
}

// IsUserOnlyIntegration returns whether the specified service is only accessible by the user.
func IsUserOnlyIntegration(svc Service) bool {
	userSpecific := []Service{GoogleSheets, Github}
	for _, s := range userSpecific {
		if s == svc {
			return true
		}
	}
	return false
}
