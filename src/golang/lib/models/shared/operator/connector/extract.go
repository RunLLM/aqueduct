package connector

import (
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Extract defines the spec for an Extract operator.
type Extract struct {
	Service       shared.Service `json:"service"`
	IntegrationId uuid.UUID      `json:"integration_id"`
	Parameters    ExtractParams  `json:"parameters"`
}

// UnmarshalJSON overrides the default unmarshalling, so that Extract.Parameters
// can be unmarshalled to the correct ExtractParams implementation.
func (e *Extract) UnmarshalJSON(data []byte) error {
	// Unmarshal data to an alias of Extract. Unmarshalling to extractAlias defers unmarshalling of
	// Parameters, since it is defined as a *json.RawMessage.
	var extractAlias struct {
		Service       shared.Service   `json:"service"`
		IntegrationId uuid.UUID        `json:"integration_id"`
		Parameters    *json.RawMessage `json:"parameters"`
	}
	if err := json.Unmarshal(data, &extractAlias); err != nil {
		return err
	}

	// Set fields that were not deferred when unmarshalling
	e.Service = extractAlias.Service
	e.IntegrationId = extractAlias.IntegrationId

	// Initialize correct destination struct for this operator's Extract.Parameters
	var params ExtractParams
	switch e.Service {
	case shared.Postgres, shared.AqueductDemo:
		params = &PostgresExtractParams{}
	case shared.Athena:
		params = &AthenaExtractParams{}
	case shared.Snowflake:
		params = &SnowflakeExtractParams{}
	case shared.MySql:
		params = &MySqlExtractParams{}
	case shared.Redshift:
		params = &RedshiftExtractParams{}
	case shared.MariaDb:
		params = &MariaDbExtractParams{}
	case shared.SqlServer:
		params = &SqlServerExtractParams{}
	case shared.BigQuery:
		params = &BigQueryExtractParams{}
	case shared.Sqlite:
		params = &SqliteExtractParams{}
	case shared.GoogleSheets:
		params = &GoogleSheetsExtractParams{}
	case shared.Salesforce:
		params = &SalesforceExtractParams{}
	case shared.S3:
		params = &S3ExtractParams{}
	case shared.MongoDB:
		params = &MongoDBExtractParams{}
	default:
		return errors.Newf("Unknown Service type: %s, unable to unmarshal ExtractParams", e.Service)
	}

	// Unmarshal extractAlias.Parameters to `params`, which is a specific implementation of ExtractParams
	if err := json.Unmarshal(*extractAlias.Parameters, params); err != nil {
		return err
	}

	e.Parameters = params
	return nil
}
