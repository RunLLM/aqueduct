package _000022_backfill_python_type

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

const (
	pythonExecutorPackage = "aqueduct_executor"
	migrationPythonPath   = "migrators.backfill_python_type_000022.main"
	metadataColumn        = "metadata"
	artifactResultTable   = "artifact_result"
)

type SerializationType string

type ArtifactType string

type Metadata struct {
	Schema []map[string]string // Table Schema from Pandas
	// Metrics from the system regarding the op used to create the artifact result.
	// A key/value pair of [metricname]metricvalue e.g. SystemMetric["runtime"] -> "3.65"
	SystemMetrics     map[string]string `json:"system_metadata,omitempty"`
	SerializationType SerializationType `json:"serialization_type,omitempty"`
	ArtifactType      ArtifactType      `json:"artifact_type,omitempty"`
	PythonType        string            `json:"python_type,omitempty"`
}

type NullMetadata struct {
	Metadata
	IsNull bool
}

func (m *Metadata) Value() (driver.Value, error) {
	return utils.ValueJSONB(*m)
}

func (m *Metadata) Scan(value interface{}) error {
	return utils.ScanJSONB(value, m)
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

type ArtifactResult struct {
	Id          uuid.UUID    `db:"id" json:"id"`
	ContentPath string       `db:"content_path" json:"content_path"`
	Metadata    NullMetadata `db:"metadata" json:"metadata"`
}

type MigrationSpec struct {
	SerializationType SerializationType    `json:"serialization_type"`
	StorageConfig     shared.StorageConfig `json:"storage_config"`
	ContentPath       string               `json:"content_path"`
}

func getAllArtifactResults(
	ctx context.Context,
	db database.Database,
) ([]ArtifactResult, error) {
	query := "SELECT id, content_path, metadata FROM artifact_result;"

	var response []ArtifactResult
	err := db.Query(ctx, &response, query)
	return response, err
}

func backfillPythonType(
	ctx context.Context,
	id uuid.UUID,
	metadata *Metadata,
	contentPath string,
	storageConfig *shared.StorageConfig,
	db database.Database,
) error {
	migrationSpec := MigrationSpec{
		SerializationType: metadata.SerializationType,
		StorageConfig:     *storageConfig,
		ContentPath:       contentPath,
	}

	specData, err := json.Marshal(migrationSpec)
	if err != nil {
		return err
	}

	// Launch the Python job to infer the type of the parameter value
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
		return errors.Newf("Error running Python migration job. Stdout: %s, Stderr: %s.", outb.String(), errb.String())
	}

	outputs := strings.Split(outb.String(), "\n")
	if len(outputs) < 2 {
		return errors.Newf("Unexpected Python migration script output: %v", outputs)
	}
	// The python type info appears in the second token.
	metadata.PythonType = outputs[1]

	changes := map[string]interface{}{
		metadataColumn: metadata,
	}

	return repos.UpdateRecord(ctx, changes, artifactResultTable, "id", id, db)
}
