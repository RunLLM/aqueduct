package connector

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMarshalAndUnmarshallLoad(t *testing.T) {
	postgresParams := generateLoadPostgresParams()
	originalLoad := Load{
		Service:       shared.Postgres,
		IntegrationId: uuid.New(),
		Parameters:    postgresParams,
	}

	data, err := json.Marshal(&originalLoad)
	require.Nil(t, err)

	var newLoad Load
	err = json.Unmarshal(data, &newLoad)
	require.Nil(t, err)

	require.True(t, reflect.DeepEqual(originalLoad, newLoad))

	// Modify the parameters of originalLoad to confirm the change is detected
	postgresParams.Table = "prod_table"
	require.False(t, reflect.DeepEqual(originalLoad, newLoad))
}

func generateLoadPostgresParams() *PostgresLoadParams {
	return &PostgresLoadParams{
		RelationalDBLoadParams: RelationalDBLoadParams{
			Table: "test_table",
		},
	}
}
