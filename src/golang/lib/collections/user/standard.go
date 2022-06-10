package user

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateUser(
	ctx context.Context,
	email, organizationId, role, auth0Id string,
	db database.Database,
) (*User, error) {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer database.TxnRollbackIgnoreErr(ctx, tx)

	reader := standardReaderImpl{}
	if _, err := reader.GetOrganizationAdmin(ctx, organizationId, tx); role != string(AdminRole) && err != nil {
		if err != database.ErrNoRows {
			return nil, err
		}

		// Create an admin for the organization, since one does not exist
		s := strings.Split(email, "@") // Parse the organization name from user email
		if len(s) < 2 {
			return nil, errors.Newf("Unable to parse organization name from user email: %s", email)
		}

		organizationName := s[1]

		// We skip creating an admin for gmail users, as they don't belong to an org.
		if organizationName != "gmail.com" {
			if _, err := w.CreateOrganizationAdmin(ctx, organizationId, organizationName, tx); err != nil {
				return nil, err
			}
		}
	}

	apiKey, err := generateApiKey(ctx, tx)
	if err != nil {
		return nil, err
	}

	insertColumns := []string{
		EmailColumn, OrganizationIdColumn, RoleColumn, ApiKeyColumn, Auth0IdColumn,
	}
	insertUserStmt := tx.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{
		email, organizationId, role, apiKey, auth0Id,
	}

	var user User
	err = tx.Query(ctx, &user, insertUserStmt, args...)

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &user, err
}

func (w *standardWriterImpl) CreateUserWithApiKey(
	ctx context.Context,
	email, organizationId, role, auth0Id, apiKey string,
	db database.Database,
) (*User, error) {
	insertColumns := []string{
		EmailColumn, OrganizationIdColumn, RoleColumn, ApiKeyColumn, Auth0IdColumn,
	}
	insertUserStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{
		email, organizationId, role, apiKey, auth0Id,
	}

	var user User
	err := db.Query(ctx, &user, insertUserStmt, args...)
	return &user, err
}

func (w *standardWriterImpl) CreateOrganizationAdmin(
	ctx context.Context,
	organizationId, organizationName string,
	db database.Database,
) (*User, error) {
	adminEmail := fmt.Sprintf("admin@%s", organizationName)
	adminAuth0Id := organizationName
	return w.CreateUser(ctx, adminEmail, organizationId, string(AdminRole), adminAuth0Id, db)
}

// GetUser will return database.ErrNoRows if no User is found
func (r *standardReaderImpl) GetUser(ctx context.Context, id uuid.UUID, db database.Database) (*User, error) {
	getUserQuery := fmt.Sprintf(
		"SELECT %s FROM app_user WHERE id = $1;",
		allColumns(),
	)
	var user User

	err := db.Query(ctx, &user, getUserQuery, id)
	return &user, err
}

// GetUserFromApiKey will return database.ErrNoRows if no User is found
func (r *standardReaderImpl) GetUserFromApiKey(
	ctx context.Context,
	apiKey string,
	db database.Database,
) (*User, error) {
	getUserQuery := fmt.Sprintf(
		"SELECT %s FROM app_user WHERE api_key = $1;",
		allColumns(),
	)
	var user User

	err := db.Query(ctx, &user, getUserQuery, apiKey)
	return &user, err
}

// GetUserFromEmail will return database.ErrNoRows if no User is found
func (r *standardReaderImpl) GetUserFromEmail(
	ctx context.Context,
	email string,
	db database.Database,
) (*User, error) {
	getUserQuery := fmt.Sprintf(
		"SELECT %s FROM app_user WHERE email = $1;",
		allColumns(),
	)
	var user User

	err := db.Query(ctx, &user, getUserQuery, email)
	return &user, err
}

// GetUserFromAuth0Id will return database.ErrNoRows if no User is found
func (r *standardReaderImpl) GetUserFromAuth0Id(
	ctx context.Context,
	auth0Id string,
	db database.Database,
) (*User, error) {
	getUserQuery := fmt.Sprintf(
		"SELECT %s FROM app_user WHERE auth0_id = $1;",
		allColumns(),
	)
	var user User

	err := db.Query(ctx, &user, getUserQuery, auth0Id)
	return &user, err
}

// GetOrganizationAdmin will return database.ErrNoRows if no admin User is found
func (r *standardReaderImpl) GetOrganizationAdmin(
	ctx context.Context,
	organizationId string,
	db database.Database,
) (*User, error) {
	getAdminQuery := fmt.Sprintf(
		"SELECT %s FROM app_user WHERE organization_id = $1 AND role = $2;",
		allColumns(),
	)
	var user User

	err := db.Query(ctx, &user, getAdminQuery, organizationId, AdminRole)
	return &user, err
}

func (r *standardReaderImpl) GetUsersInOrganization(
	ctx context.Context,
	organizationId string,
	db database.Database,
) ([]User, error) {
	getUsersInOrgQuery := fmt.Sprintf(
		"SELECT %s FROM app_user WHERE organization_id = $1;",
		allColumns(),
	)
	var users []User

	err := db.Query(ctx, &users, getUsersInOrgQuery, organizationId)
	return users, err
}

func (r *standardReaderImpl) GetWatchersByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) ([]User, error) {
	getWatchersStmt := fmt.Sprintf(
		`SELECT %s FROM workflow_watcher INNER JOIN app_user 
				ON workflow_watcher.user_id = app_user.id 
				WHERE workflow_watcher.workflow_id = $1;`,
		allColumnsWithPrefix(),
	)
	var users []User

	err := db.Query(ctx, &users, getWatchersStmt, workflowId)
	return users, err
}

func (w *standardWriterImpl) ResetApiKey(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*User, error) {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer database.TxnRollbackIgnoreErr(ctx, tx)

	newApiKey, err := generateApiKey(ctx, tx)
	if err != nil {
		return nil, err
	}

	updateColumns := []string{ApiKeyColumn}
	updateApiKeyStmt := tx.PrepareUpdateWhereWithReturnAllStmt(tableName, updateColumns, IdColumn, allColumns())

	args := []interface{}{
		newApiKey, id,
	}

	var user User

	if err := tx.Query(ctx, &user, updateApiKeyStmt, args...); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &user, nil
}

func (w *standardWriterImpl) DeleteUser(ctx context.Context, id uuid.UUID, db database.Database) error {
	deleteUserStmt := `DELETE FROM app_user WHERE id = $1;`
	return db.Execute(ctx, deleteUserStmt, id)
}

func (w *standardWriterImpl) DeleteUserWithAuth0Id(
	ctx context.Context,
	auth0Id string,
	db database.Database,
) error {
	deleteUserStmt := `DELETE FROM app_user WHERE auth0_id = $1;`
	return db.Execute(ctx, deleteUserStmt, auth0Id)
}

// Helper function to generate a unique API key
func generateApiKey(ctx context.Context, db database.Database) (string, error) {
	checkApiKeyQuery := fmt.Sprintf(
		"SELECT %s FROM app_user WHERE api_key = $1;",
		allColumns(),
	)

	for {
		apiKey, err := tokenGenerator()
		if err != nil {
			return "", err
		}

		err = db.Query(ctx, &User{}, checkApiKeyQuery, apiKey)

		if err == database.ErrNoRows {
			// Generated API key is unique
			return apiKey, nil
		}

		if err != nil {
			return "", err
		}
	}
}

func tokenGenerator() (string, error) {
	b := make([]byte, apiKeyLength/2)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
