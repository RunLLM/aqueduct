package sqlite

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

const (
	apiKeyLength = 60
)

type userRepo struct {
	userReader
	userWriter
}

type userReader struct{}

type userWriter struct{}

func NewUserRepo() repos.User {
	return &userRepo{
		userReader: userReader{},
		userWriter: userWriter{},
	}
}

func (*userReader) GetByAPIKey(ctx context.Context, apiKey string, DB database.Database) (*models.User, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM app_user WHERE api_key = $1;`,
		models.UserCols(),
	)
	args := []interface{}{apiKey}
	return getUser(ctx, DB, query, args...)
}

func (*userWriter) Create(
	ctx context.Context,
	orgID string,
	apiKey string,
	DB database.Database,
) (*models.User, error) {
	cols := []string{
		models.UserID,
		models.UserEmail,
		models.UserOrgID,
		models.UserRole,
		models.UserAPIKey,
		models.UserAuth0ID,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.UserTable, cols, models.UserCols())

	ID, err := GenerateUniqueUUID(ctx, models.UserTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		"", /* email */
		orgID,
		"", /* role */
		apiKey,
		"", /* auth0_id */
	}
	return getUser(ctx, DB, query, args...)
}

func (*userWriter) ResetAPIKey(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.User, error) {
	TX, err := DB.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer database.TxnRollbackIgnoreErr(ctx, TX)

	newAPIKey, err := generateAPIKey(ctx, TX)
	if err != nil {
		return nil, err
	}

	cols := []string{models.UserAPIKey}
	query := TX.PrepareUpdateWhereWithReturnAllStmt(models.UserTable, cols, models.UserID, models.UserCols())
	args := []interface{}{newAPIKey, ID}

	user, err := getUser(ctx, TX, query, args...)
	if err != nil {
		return nil, err
	}

	if err := TX.Commit(ctx); err != nil {
		return nil, err
	}

	return user, err
}

func getUsers(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.User, error) {
	var users []models.User
	err := DB.Query(ctx, &users, query, args...)
	return users, err
}

func getUser(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.User, error) {
	users, err := getUsers(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, database.ErrNoRows()
	}

	if len(users) != 1 {
		return nil, errors.Newf("Expected 1 user but got %v", len(users))
	}

	return &users[0], nil
}

// generateAPIKey generates a unique API key.
func generateAPIKey(ctx context.Context, DB database.Database) (string, error) {
	for {
		b := make([]byte, apiKeyLength/2)
		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}
		apiKey := fmt.Sprintf("%x", b)

		r := &userReader{}
		_, err = r.GetByAPIKey(ctx, apiKey, DB)
		if err != nil && errors.Is(err, database.ErrNoRows()) {
			// No row with this API key was found
			return apiKey, nil
		}

		if err != nil {
			return "", err
		}
	}
}
