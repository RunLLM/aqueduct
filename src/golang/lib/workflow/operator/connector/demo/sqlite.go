package demo

import (
	"os"
	"path"

	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
)

const (
	databasePathKey = "database"
)

var sqliteDatabasePath = path.Join(os.Getenv("HOME"), ".aqueduct/server/db/demo.db")

func GetSqliteIntegrationConfig() auth.Config {
	configMap := map[string]string{
		databasePathKey: sqliteDatabasePath,
	}

	config := auth.NewStaticConfig(configMap)
	return config
}
