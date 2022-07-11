package artifact_test

import (
	"encoding/json"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact/table"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSerializingAndDeserializingArtifact(t *testing.T) {
	id := uuid.New()

	atf := artifact.DBArtifact{
		Id:   id,
		Name: "test",
		Spec: *artifact.NewSpecFromTable(
			table.Table{},
		),
	}

	rawAtf, err := json.Marshal(atf)
	require.Nil(t, err)

	var reconstructedAtf artifact.DBArtifact
	err = json.Unmarshal(rawAtf, &reconstructedAtf)
	require.Nil(t, err)
	require.True(t, reconstructedAtf.Spec.IsTable())
	require.NotNil(t, reconstructedAtf.Spec.Table())
}
