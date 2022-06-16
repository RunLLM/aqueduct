package user

import (
	"context"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
)

type sqliteReaderImpl struct {
	standardReaderImpl
}

type sqliteWriterImpl struct {
	standardWriterImpl
}

func newSqliteReader() Reader {
	return &sqliteReaderImpl{standardReaderImpl{}}
}

func newSqliteWriter() Writer {
	return &sqliteWriterImpl{standardWriterImpl{}}
}

func (w *sqliteWriterImpl) CreateUser(
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
		IdColumn, EmailColumn, OrganizationIdColumn, RoleColumn, ApiKeyColumn, Auth0IdColumn,
	}
	insertUserStmt := tx.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		id, email, organizationId, role, apiKey, auth0Id,
	}

	var user User
	err = tx.Query(ctx, &user, insertUserStmt, args...)

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &user, err
}

func (w *sqliteWriterImpl) CreateUserWithApiKey(
	ctx context.Context,
	email, organizationId, role, auth0Id, apiKey string,
	db database.Database,
) (*User, error) {
	insertColumns := []string{
		IdColumn, EmailColumn, OrganizationIdColumn, RoleColumn, ApiKeyColumn, Auth0IdColumn,
	}
	insertUserStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		id, email, organizationId, role, apiKey, auth0Id,
	}

	var user User
	err = db.Query(ctx, &user, insertUserStmt, args...)
	return &user, err
}
