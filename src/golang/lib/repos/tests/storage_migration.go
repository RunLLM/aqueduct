package tests

import (
	"github.com/aqueducthq/aqueduct/lib/database"
	aq_errors "github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestStorageMigrationList() {
	migrations := ts.seedStorageMigration()

	actualMigrations, err := ts.storageMigration.List(ts.ctx, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), migrations, actualMigrations)
}

func (ts *TestSuite) TestStorageMigrationCurrent() {
	current, err := ts.storageMigration.Current(ts.ctx, ts.DB)
	require.True(ts.T(), aq_errors.Is(err, database.ErrNoRows()))
	require.Nil(ts.T(), current)

	migrations := ts.seedStorageMigration()

	current, err = ts.storageMigration.Current(ts.ctx, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), migrations[0], *current)
}
