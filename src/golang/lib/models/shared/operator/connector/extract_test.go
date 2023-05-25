package connector

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMarshalAndUnmarshallExtract(t *testing.T) {
	postgresParams := generateExtractPostgresParams()
	originalExtract := Extract{
		Service:    shared.Postgres,
		ResourceId: uuid.New(),
		Parameters: postgresParams,
	}

	data, err := json.Marshal(&originalExtract)
	require.Nil(t, err)

	var newExtract Extract
	err = json.Unmarshal(data, &newExtract)
	require.Nil(t, err)

	require.True(t, reflect.DeepEqual(originalExtract, newExtract))

	// Modify the parameters of originalExtract to confirm the change is detected
	postgresParams.Query = "SELECT * FROM reviews;"
	require.False(t, reflect.DeepEqual(originalExtract, newExtract))
}

func generateExtractPostgresParams() *PostgresExtractParams {
	return &PostgresExtractParams{
		RelationalDBExtractParams: RelationalDBExtractParams{
			Query: "SELECT * FROM wine;",
		},
	}
}
