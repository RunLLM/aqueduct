package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type executionEnvironmentRepo struct {
	executionEnvironmentReader
	executionEnvironmentWriter
}

type executionEnvironmentReader struct{}

type executionEnvironmentWriter struct{}

func NewExecutionEnvironmentRepo() repos.ExecutionEnvironment {
	return &executionEnvironmentRepo{
		executionEnvironmentReader: executionEnvironmentReader{},
		executionEnvironmentWriter: executionEnvironmentWriter{},
	}
}

func (*executionEnvironmentReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.ExecutionEnvironment, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM execution_environment WHERE id = $1;`,
		models.ExecutionEnvironmentCols(),
	)
	args := []interface{}{ID}

	return getExecutionEnvironment(ctx, DB, query, args...)
}

func (*executionEnvironmentReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.ExecutionEnvironment, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM execution_environment WHERE id IN (%s);`,
		models.ExecutionEnvironmentCols(),
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getExecutionEnvironments(ctx, DB, query, args...)
}

func (*executionEnvironmentReader) GetByHash(ctx context.Context, hash uuid.UUID, DB database.Database) (*models.ExecutionEnvironment, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM execution_environment WHERE hash = $1;`,
		models.ExecutionEnvironmentCols(),
	)
	args := []interface{}{hash}

	return getExecutionEnvironment(ctx, DB, query, args...)
}

func (*executionEnvironmentReader) GetByOperatorBatch(ctx context.Context, opIDs []uuid.UUID, DB database.Database) (map[uuid.UUID]models.ExecutionEnvironment, error) {
	type resultRow struct {
		ID               uuid.UUID                       `db:"id"`
		OpID             uuid.UUID                       `db:"operator_id"`
		Hash             uuid.UUID                       `db:"hash"`
		Spec             shared.ExecutionEnvironmentSpec `db:"spec"`
		GarbageCollected bool                            `db:"garbage_collected"`
	}

	query := fmt.Sprintf(
		`SELECT operator.id AS operator_id, %s
		FROM execution_environment, operator
		WHERE operator.execution_environment_id = execution_environment.id
		AND operator.id IN (%s);`,
		models.ExecutionEnvironmentColsWithPrefix(),
		stmt_preparers.GenerateArgsList(len(opIDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(opIDs)

	var results []resultRow
	err := DB.Query(ctx, &results, query, args...)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[uuid.UUID]models.ExecutionEnvironment, len(results))
	for _, row := range results {
		resultMap[row.OpID] = models.ExecutionEnvironment{
			ID:   row.ID,
			Spec: row.Spec,
			Hash: row.Hash,
		}
	}

	return resultMap, nil
}

func (*executionEnvironmentWriter) Create(
	ctx context.Context,
	spec *shared.ExecutionEnvironmentSpec,
	hash uuid.UUID,
	DB database.Database,
) (*models.ExecutionEnvironment, error) {
	cols := []string{
		models.ExecutionEnvironmentID,
		models.ExecutionEnvironmentSpec,
		models.ExecutionEnvironmentHash,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.ExecutionEnvironmentTable, cols, models.ExecutionEnvironmentCols())

	ID, err := GenerateUniqueUUID(ctx, models.ExecutionEnvironmentTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		spec,
		hash,
	}
	return getExecutionEnvironment(ctx, DB, query, args...)
}

func (w *executionEnvironmentWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	return w.DeleteBatch(ctx, []uuid.UUID{ID}, DB)
}

func (*executionEnvironmentWriter) DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error {
	if len(IDs) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"DELETE FROM execution_environment WHERE id IN (%s);",
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(IDs)
	return DB.Execute(ctx, query, args...)
}

func (*executionEnvironmentWriter) Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.ExecutionEnvironment, error) {
	var executionEnvironment models.ExecutionEnvironment
	err := repos.UpdateRecordToDest(ctx, &executionEnvironment, changes, models.ExecutionEnvironmentTable, models.ExecutionEnvironmentID, ID, models.ExecutionEnvironmentCols(), DB)
	return &executionEnvironment, err
}

func getExecutionEnvironments(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.ExecutionEnvironment, error) {
	var executionEnvironments []models.ExecutionEnvironment
	err := DB.Query(ctx, &executionEnvironments, query, args...)
	return executionEnvironments, err
}

func getExecutionEnvironment(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.ExecutionEnvironment, error) {
	executionEnvironments, err := getExecutionEnvironments(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(executionEnvironments) == 0 {
		return nil, database.ErrNoRows()
	}

	if len(executionEnvironments) != 1 {
		return nil, errors.Newf("Expected 1 execution environment but got %v", len(executionEnvironments))
	}

	return &executionEnvironments[0], nil
}
