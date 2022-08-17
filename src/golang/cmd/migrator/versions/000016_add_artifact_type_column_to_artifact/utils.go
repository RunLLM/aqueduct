package _000016_add_artifact_type_column_to_artifact

import (
	"context"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

const (
	pythonExecutorPackage = "aqueduct_executor"
	migrationPythonPath   = "migrators.artifact_migration.main"
)

var confPath = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "config", "config.yml")

type Type string

const (
	TableType Type = "table"
	FloatType Type = "float"
	BoolType  Type = "boolean"
	JsonType  Type = "json"
)

type Table struct{}

type Float struct{}

type Bool struct{}

type Json struct{}

type specUnion struct {
	Type  Type   `json:"type"`
	Table *Table `json:"table,omitempty"`

	// TODO(ENG-1119): The float artifact currently also represents integers.
	Float *Float `json:"float,omitempty"`
	Bool  *Bool  `json:"bool,omitempty"`
	Json  *Json  `json:"jsonable,omitempty"`
}

type Spec struct {
	spec specUnion
}

func (s Spec) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.spec)
}

func (s *Spec) UnmarshalJSON(rawMessage []byte) error {
	var spec specUnion
	err := json.Unmarshal(rawMessage, &spec)
	if err != nil {
		return err
	}

	// Overwrite the spec type based on the data.
	var typeCount int
	if spec.Table != nil {
		spec.Type = TableType
		typeCount++
	} else if spec.Float != nil {
		spec.Type = FloatType
		typeCount++
	} else if spec.Bool != nil {
		spec.Type = BoolType
		typeCount++
	} else if spec.Json != nil {
		spec.Type = JsonType
		typeCount++
	}

	if typeCount != 1 {
		return errors.Newf("Artifact Spec can only be of one type. Number of types: %d", typeCount)
	}

	s.spec = spec
	return nil
}

func (s *Spec) Value() (driver.Value, error) {
	return utils.ValueJsonB(*s)
}

func (s *Spec) Scan(value interface{}) error {
	return utils.ScanJsonB(value, s)
}

type artifactSpec struct {
	Id   uuid.UUID `db:"id"`
	Spec Spec      `db:"spec"`
}

func getArtifactSpec(ctx context.Context, db database.Database) ([]artifactSpec, error) {
	query := "SELECT id, spec FROM artifact;"

	var result []artifactSpec
	err := db.Query(ctx, &result, query)
	return result, err
}

type Metadata struct {
	Schema []map[string]string // Table Schema from Pandas
	// Metrics from the system regarding the op used to create the artifact result.
	// A key/value pair of [metricname]metricvalue e.g. SystemMetric["runtime"] -> "3.65"
	SystemMetrics     map[string]string `json:"system_metadata,omitempty"`
	SerializationType SerializationType `json:"serialization_type,omitempty"`
	ArtifactType      NewType           `json:"artifact_type,omitempty"`
}

type NullMetadata struct {
	Metadata
	IsNull bool
}

type artifactResult struct {
	Id          uuid.UUID    `db:"id"`
	Metadata    NullMetadata `db:"metadata"`
	ContentPath string       `db:"content_path" json:"content_path"`
}

func getArtifactResult(ctx context.Context, db database.Database, artifactId uuid.UUID) ([]artifactResult, error) {
	query := "SELECT id, metadata, content_path FROM artifact_result WHERE artifact_id = $1;"

	var result []artifactResult
	err := db.Query(ctx, &result, query, artifactId)
	return result, err
}

func updateMetadataInArtifactResult(
	ctx context.Context,
	id uuid.UUID,
	metadata *NullMetadata,
	db database.Database,
) error {
	changes := map[string]interface{}{
		"metadata": metadata,
	}
	return utils.UpdateRecord(ctx, changes, "artifact_result", "id", id, db)
}

type MigrationSpec struct {
	ArtifactType  string               `json:"artifact_type"`
	StorageConfig shared.StorageConfig `json:"storage_config"`
	MetadataPath  string               `json:"metadata_path"`
	ContentPath   string               `json:"content_path"`
}

type NewType string

const (
	Untyped          NewType = "untyped"
	NewStringType    NewType = "string"
	NewBoolType      NewType = "bool"
	NewNumericType   NewType = "numeric"
	NewDictType      NewType = "dictionary"
	NewTupleType     NewType = "tuple"
	NewTableType     NewType = "table"
	NewJsonType      NewType = "json"
	NewBytesType     NewType = "bytes"
	NewImageType     NewType = "image"
	NewPicklableType NewType = "picklable"
)

type SerializationType string

const (
	StringSerializationType    SerializationType = "string"
	TableSerializationType     SerializationType = "table"
	JsonSerializationType      SerializationType = "json"
	BytesSerializationType     SerializationType = "bytes"
	ImageSerializationType     SerializationType = "image"
	PicklableSerializationType SerializationType = "picklable"
)

type TypeMetadata struct {
	ArtifactType      NewType           `json:"artifact_type"`
	SerializationType SerializationType `json:"serialization_type"`
}

func updateTypeInArtifact(
	ctx context.Context,
	id uuid.UUID,
	artifactType NewType,
	db database.Database,
) error {
	changes := map[string]interface{}{
		"type": artifactType,
	}
	return utils.UpdateRecord(ctx, changes, "artifact", "id", id, db)
}

func migrateArtifact(ctx context.Context, db database.Database) error {
	artifactSpecs, err := getArtifactSpec(ctx, db)
	if err != nil {
		return err
	}

	for _, artifactSpec := range artifactSpecs {
		artifactResults, err := getArtifactResult(ctx, db, artifactSpec.Id)
		if err != nil {
			return err
		}

		newArtifactType := ""

		for _, artifactResult := range artifactResults {
			metadataPath := fmt.Sprintf("%s_%s", artifactResult.Id, "metadata")
			storageConfig := config.ParseServerConfiguration(confPath).StorageConfig

			migrationSpec := MigrationSpec{
				ArtifactType:  string(artifactSpec.Spec.spec.Type),
				StorageConfig: *storageConfig,
				MetadataPath:  metadataPath,
				ContentPath:   artifactResult.ContentPath,
			}

			specData, err := json.Marshal(migrationSpec)
			if err != nil {
				return err
			}

			cmd := exec.Command(
				"python3",
				"-m",
				fmt.Sprintf("%s.%s", pythonExecutorPackage, migrationPythonPath),
				"--spec",
				base64.StdEncoding.EncodeToString(specData),
			)
			cmd.Env = os.Environ()

			err = cmd.Run()
			if err != nil {
				return err
			}

			serializedTypeMetadata, err := storage.NewStorage(storageConfig).Get(ctx, metadataPath)
			if err != nil {
				return err
			}

			var typeMetadata TypeMetadata
			err = json.Unmarshal(serializedTypeMetadata, &typeMetadata)
			if err != nil {
				return err
			}

			err = storage.NewStorage(storageConfig).Delete(ctx, metadataPath)
			if err != nil {
				return err
			}

			artifactResult.Metadata.ArtifactType = typeMetadata.ArtifactType
			artifactResult.Metadata.SerializationType = typeMetadata.SerializationType

			err = updateMetadataInArtifactResult(ctx, artifactResult.Id, &artifactResult.Metadata, db)
			if err != nil {
				return err
			}

			newArtifactType = string(typeMetadata.ArtifactType)
		}

		if newArtifactType != "" {
			err = updateTypeInArtifact(ctx, artifactSpec.Id, NewType(newArtifactType), db)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
