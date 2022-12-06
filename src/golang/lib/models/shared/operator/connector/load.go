package connector

import (
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Load defines the spec for a Load operator.
type Load struct {
	Service       integration.Service `json:"service"`
	IntegrationID uuid.UUID           `json:"integration_id"`
	Parameters    LoadParams          `json:"parameters"`
}

// UnmarshalJSON overrides the default unmarshalling, so that Load.Parameters
// can be unmarshalled to the correct LoadParams implementation.
func (l *Load) UnmarshalJSON(data []byte) error {
	// Unmarshal data to an alias of Load. Unmarshalling to loadAlias defers unmarshalling of
	// Parameters, since it is defined as a *json.RawMessage.
	var loadAlias struct {
		Service       integration.Service `json:"service"`
		IntegrationID uuid.UUID           `json:"integration_id"`
		Parameters    *json.RawMessage    `json:"parameters"`
	}
	if err := json.Unmarshal(data, &loadAlias); err != nil {
		return err
	}

	// Set fields that were not deferred when unmarshalling
	l.Service = loadAlias.Service
	l.IntegrationID = loadAlias.IntegrationID

	// Initialize correct destination struct for this operator's Load.Parameters
	var params LoadParams
	switch l.Service {
	case integration.Postgres, integration.AqueductDemo:
		params = &PostgresLoadParams{}
	case integration.Snowflake:
		params = &SnowflakeLoadParams{}
	case integration.MySql:
		params = &MySqlLoadParams{}
	case integration.Redshift:
		params = &RedshiftLoadParams{}
	case integration.MariaDb:
		params = &MariaDbLoadParams{}
	case integration.SqlServer:
		params = &SqlServerLoadParams{}
	case integration.BigQuery:
		params = &BigQueryLoadParams{}
	case integration.Sqlite:
		params = &SqliteLoadParams{}
	case integration.GoogleSheets:
		params = &GoogleSheetsLoadParams{}
	case integration.Salesforce:
		params = &SalesforceLoadParams{}
	case integration.S3:
		params = &S3LoadParams{}
	default:
		return errors.Newf("Unknown Service type: %s, unable to unmarshal LoadParams", l.Service)
	}

	// Unmarshal loadAlias.Parameters to `params`, which is a specific implementation of LoadParams
	if err := json.Unmarshal(*loadAlias.Parameters, params); err != nil {
		return err
	}

	l.Parameters = params
	return nil
}
