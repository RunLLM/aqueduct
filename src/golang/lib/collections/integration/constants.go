package integration

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
	AqueductDemo Service = "Aqueduct Demo"
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

	DemoDbIntegrationName = "aqueduct_demo"
)

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
		Slack:
		return svc, nil
	default:
		return "", errors.Newf("Unknown service: %s", s)
	}
}

func GetRelationalDatabaseIntegrations() map[Service]bool {
	return map[Service]bool{
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
