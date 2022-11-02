package connector

import (
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Extract defines the spec for an Extract operator.
type Extract struct {
	Service       integration.Service `json:"service"`
	IntegrationId uuid.UUID           `json:"integration_id"`
	Parameters    ExtractParams       `json:"parameters"`
}

// UnmarshalJSON overrides the default unmarshalling, so that Extract.Parameters
// can be unmarshalled to the correct ExtractParams implementation.
func (e *Extract) UnmarshalJSON(data []byte) error {
	// Unmarshal data to an alias of Extract. Unmarshalling to extractAlias defers unmarshalling of
	// Parameters, since it is defined as a *json.RawMessage.
	var extractAlias struct {
		Service       integration.Service `json:"service"`
		IntegrationId uuid.UUID           `json:"integration_id"`
		Parameters    *json.RawMessage    `json:"parameters"`
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
	case integration.Postgres, integration.AqueductDemo:
		params = &PostgresExtractParams{}
	case integration.Athena:
		params = &AthenaExtractParams{}
	case integration.Snowflake:
		params = &SnowflakeExtractParams{}
	case integration.MySql:
		params = &MySqlExtractParams{}
	case integration.Redshift:
		params = &RedshiftExtractParams{}
	case integration.MariaDb:
		params = &MariaDbExtractParams{}
	case integration.SqlServer:
		params = &SqlServerExtractParams{}
	case integration.BigQuery:
		params = &BigQueryExtractParams{}
	case integration.Sqlite:
		params = &SqliteExtractParams{}
	case integration.GoogleSheets:
		params = &GoogleSheetsExtractParams{}
	case integration.Salesforce:
		params = &SalesforceExtractParams{}
	case integration.S3:
		params = &S3ExtractParams{}
	case integration.MongoDb:
		params = &MongoDbExtractParams{}
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
