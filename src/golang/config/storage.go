package config

import "github.com/aqueducthq/aqueduct/lib/collections/shared"

var globalStorageConfig *shared.StorageConfig

func GetStorage() shared.StorageConfig {
	return *globalStorageConfig
}

func UpdateStorage(newConfig *shared.StorageConfig) error {
	globalStorageConfig = newConfig
	return dumpGlobalStorage()
}

// dumpGlobalStorage writes `globalStorageConfig` to the server config.yml file.
func dumpGlobalStorage() error {

}
