package _000016_add_artifact_type_column_to_artifact

import (
	"bytes"
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
	log "github.com/sirupsen/logrus"
)

const (
	pythonExecutorPackage = "aqueduct_executor"
	migrationPythonPath   = "migrators.artifact_migration_000016.main"
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
	ArtifactType      NewArtifactType   `json:"artifact_type,omitempty"`
}

type NullMetadata struct {
	Metadata
	IsNull bool
}

func (m *Metadata) Value() (driver.Value, error) {
	return utils.ValueJsonB(*m)
}

func (m *Metadata) Scan(value interface{}) error {
	return utils.ScanJsonB(value, m)
}

func (n *NullMetadata) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.Metadata).Value()
}

func (n *NullMetadata) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	metadata := &Metadata{}
	if err := metadata.Scan(value); err != nil {
		return err
	}

	n.Metadata, n.IsNull = *metadata, false
	return nil
}

type ExecutionStatus string

const (
	SucceededExecutionStatus ExecutionStatus = "succeeded"
)

type artifactResult struct {
	Id          uuid.UUID       `db:"id"`
	Metadata    NullMetadata    `db:"metadata"`
	ContentPath string          `db:"content_path" json:"content_path"`
	Status      ExecutionStatus `db:"status" json:"status"`
}

func getArtifactResult(ctx context.Context, db database.Database, artifactId uuid.UUID) ([]artifactResult, error) {
	query := "SELECT id, metadata, content_path, status FROM artifact_result WHERE artifact_id = $1;"

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

type NewArtifactType string

type SerializationType string

type TypeMetadata struct {
	ArtifactType      NewArtifactType   `json:"artifact_type"`
	SerializationType SerializationType `json:"serialization_type"`
}

func updateTypeInArtifact(
	ctx context.Context,
	id uuid.UUID,
	artifactType NewArtifactType,
	db database.Database,
) error {
	changes := map[string]interface{}{
		"type": artifactType,
	}
	return utils.UpdateRecord(ctx, changes, "artifact", "id", id, db)
}

func migrateArtifact(ctx context.Context, db database.Database) error {
	if err := config.Init(confPath); err != nil {
		return err
	}

	artifactResultMigrated := 0

	artifactSpecs, err := getArtifactSpec(ctx, db)
	if err != nil {
		return err
	}

	for _, artifactSpec := range artifactSpecs {
		artifactResults, err := getArtifactResult(ctx, db, artifactSpec.Id)
		if err != nil {
			return err
		}

		newArtifactType := NewArtifactType("")

		for _, artifactResult := range artifactResults {
			if artifactResult.Metadata.ArtifactType != "" && artifactResult.Metadata.SerializationType != "" {
				log.Infof("Skipping data migration for artifact result %s since its content has already been migrated.", artifactResult.Id)
				continue
			}
			// Temporaty file to store the updated metadata dict that contains
			// the serialization type and artifact type.
			metadataPath := fmt.Sprintf("%s_%s", artifactResult.Id, "metadata")
			sConfig := config.Storage()
			storageConfig := &sConfig

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

			// Before launching storage migration, copy the content file in case the database migration fails
			// and we need to revert the content change.
			originalContent, err := storage.NewStorage(storageConfig).Get(ctx, artifactResult.ContentPath)
			if err != nil {
				if artifactResult.Status != SucceededExecutionStatus && err == storage.ErrObjectDoesNotExist {
					log.Infof("Skipping data migration for artifact result %s since its content wasn't generated.", artifactResult.Id)
					continue
				} else {
					log.Errorf("Unexpected error while migrating artifact result %s: %s.", artifactResult.Id, err)
					return err
				}
			}

			defer func() {
				if err != nil {
					putOperatorError := storage.NewStorage(storageConfig).Put(ctx, artifactResult.ContentPath, originalContent)
					if putOperatorError != nil {
						log.Errorf("Storage migration rollback failed due to error %s", putOperatorError)
					}
				}
			}()

			// Launch the Python migration job with the spec constructed above.
			cmd := exec.Command(
				"python3",
				"-m",
				fmt.Sprintf("%s.%s", pythonExecutorPackage, migrationPythonPath),
				"--spec",
				base64.StdEncoding.EncodeToString(specData),
			)
			cmd.Env = os.Environ()

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb

			err = cmd.Run()
			if err != nil {
				log.Errorf("Error running Python migration job. Stdout: %s, Stderr: %s.", outb.String(), errb.String())
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

			// Garbage collect the temp file.
			err = storage.NewStorage(storageConfig).Delete(ctx, metadataPath)
			if err != nil {
				return err
			}

			artifactResult.Metadata.ArtifactType = typeMetadata.ArtifactType
			artifactResult.Metadata.SerializationType = typeMetadata.SerializationType

			// Update artifact_result table's metadata column with the serialization type and
			// artifact type.
			err = updateMetadataInArtifactResult(ctx, artifactResult.Id, &artifactResult.Metadata, db)
			if err != nil {
				return err
			}

			newArtifactType = typeMetadata.ArtifactType

			artifactResultMigrated += 1
		}

		if newArtifactType != "" {
			// Update artifact table's type column with the artifact type.
			err = updateTypeInArtifact(ctx, artifactSpec.Id, newArtifactType, db)
			if err != nil {
				return err
			}
		} else {
			// If we reach here, it means the artifact has no result available, so we do a best-effort
			// mapping netween the original artifact type and the new type.
			if artifactSpec.Spec.spec.Type == TableType {
				err = updateTypeInArtifact(ctx, artifactSpec.Id, "table", db)
				if err != nil {
					return err
				}
			} else if artifactSpec.Spec.spec.Type == FloatType {
				err = updateTypeInArtifact(ctx, artifactSpec.Id, "numeric", db)
				if err != nil {
					return err
				}
			} else if artifactSpec.Spec.spec.Type == BoolType {
				err = updateTypeInArtifact(ctx, artifactSpec.Id, "boolean", db)
				if err != nil {
					return err
				}
			} else if artifactSpec.Spec.spec.Type == JsonType {
				// Since we don't know the real type of a parameter, we put untyped for now and
				// later when the workflow is executed, we will update this field from untyped to
				// its real type.
				err = updateTypeInArtifact(ctx, artifactSpec.Id, "untyped", db)
				if err != nil {
					return err
				}
			} else {
				return errors.Newf("Unexpected original artifact type %s", artifactSpec.Spec.spec.Type)
			}
		}
	}

	log.Infof("A total of %d artifact results have been migrated.", artifactResultMigrated)

	return nil
}
