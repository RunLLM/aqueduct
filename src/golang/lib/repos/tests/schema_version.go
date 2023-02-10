package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestSchemaVersion_Get() {
	schemaVersions := ts.seedSchemaVersion(1)
	expectedSchemaVersion := &schemaVersions[0]

	actualSchemaVersion, err := ts.schemaVersion.Get(ts.ctx, expectedSchemaVersion.Version, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedSchemaVersion, actualSchemaVersion)
}

func (ts *TestSuite) TestSchemaVersion_GetCurrent() {
	schemaVersions := ts.seedSchemaVersion(5)
	expectedSchemaVersion := &schemaVersions[4]

	actualSchemaVersion, err := ts.schemaVersion.GetCurrent(ts.ctx, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedSchemaVersion, actualSchemaVersion)
}

func (ts *TestSuite) TestSchemaVersion_Create() {
	expectedSchemaVersion := &models.SchemaVersion{
		Version: int64(models.CurrentSchemaVersion + 1),
		Dirty:   true,
		Name:    randString(10),
	}

	actualSchemaVersion, err := ts.schemaVersion.Create(ts.ctx, expectedSchemaVersion.Version, expectedSchemaVersion.Name, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), expectedSchemaVersion, actualSchemaVersion)
}

func (ts *TestSuite) TestSchemaVersion_Delete() {
	schemaVersions := ts.seedSchemaVersion(1)

	err := ts.schemaVersion.Delete(ts.ctx, schemaVersions[0].Version, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestSchemaVersion_Update() {
	schemaVersions := ts.seedSchemaVersion(1)
	expectedSchemaVersion := schemaVersions[0]

	name := randString(10)

	changes := map[string]interface{}{
		models.SchemaVersionName: name,
	}

	actualSchemaVersion, err := ts.schemaVersion.Update(ts.ctx, expectedSchemaVersion.Version, changes, ts.DB)
	require.Nil(ts.T(), err)

	expectedSchemaVersion.Name = name

	requireDeepEqual(ts.T(), &expectedSchemaVersion, actualSchemaVersion)
}
