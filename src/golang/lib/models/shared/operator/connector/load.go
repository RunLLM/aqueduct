package connector

import (
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

// Load defines the spec for a Load operator.
type Load struct {
	Service       shared.Service `json:"service"`
	IntegrationId uuid.UUID      `json:"integration_id"`
	Parameters    LoadParams     `json:"parameters"`
}

// UnmarshalJSON overrides the default unmarshalling, so that Load.Parameters
// can be unmarshalled to the correct LoadParams implementation.
func (l *Load) UnmarshalJSON(data []byte) error {
	// Unmarshal data to an alias of Load. Unmarshalling to loadAlias defers unmarshalling of
	// Parameters, since it is defined as a *json.RawMessage.
	var loadAlias struct {
		Service       shared.Service   `json:"service"`
		IntegrationId uuid.UUID        `json:"integration_id"`
		Parameters    *json.RawMessage `json:"parameters"`
	}
	if err := json.Unmarshal(data, &loadAlias); err != nil {
		return err
	}

	// Set fields that were not deferred when unmarshalling
	l.Service = loadAlias.Service
	l.IntegrationId = loadAlias.IntegrationId

	// Initialize correct destination struct for this operator's Load.Parameters
	var params LoadParams
	switch l.Service {
	case shared.Postgres, shared.AqueductDemo:
		params = &PostgresLoadParams{}
	case shared.Snowflake:
		params = &SnowflakeLoadParams{}
	case shared.MySql:
		params = &MySqlLoadParams{}
	case shared.Redshift:
		params = &RedshiftLoadParams{}
	case shared.MariaDb:
		params = &MariaDbLoadParams{}
	case shared.SqlServer:
		params = &SqlServerLoadParams{}
	case shared.BigQuery:
		params = &BigQueryLoadParams{}
	case shared.Sqlite:
		params = &SqliteLoadParams{}
	case shared.GoogleSheets:
		params = &GoogleSheetsLoadParams{}
	case shared.Salesforce:
		params = &SalesforceLoadParams{}
	case shared.S3:
		params = &S3LoadParams{}
	case shared.MongoDB:
		params = &MongoDBLoadParams{}
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

func (l *Load) Scan(value interface{}) error {
	return utils.ScanJSONB(value, l)
}
