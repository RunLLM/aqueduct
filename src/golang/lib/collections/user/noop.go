package user

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type noopReaderImpl struct {
	throwError bool
}

type noopWriterImpl struct {
	throwError bool
}

func NewNoopReader(throwError bool) Reader {
	return &noopReaderImpl{throwError: throwError}
}

func NewNoopWriter(throwError bool) Writer {
	return &noopWriterImpl{throwError: throwError}
}

func (w *noopWriterImpl) CreateUser(
	ctx context.Context,
	email, organizationId, role, auth0Id string,
	db database.Database,
) (*User, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) CreateUserWithApiKey(
	ctx context.Context,
	email, organizationId, role, auth0Id, apiKey string,
	db database.Database,
) (*User, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) CreateOrganizationAdmin(
	ctx context.Context,
	organizationId, organizationName string,
	db database.Database,
) (*User, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

// GetUser will return database.ErrNoRows if no User is found
func (r *noopReaderImpl) GetUser(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*User, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

// GetUserFromApiKey will return database.ErrNoRows if no User is found
func (r *noopReaderImpl) GetUserFromApiKey(
	ctx context.Context,
	apiKey string,
	db database.Database,
) (*User, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

// GetUserFromEmail will return database.ErrNoRows if no User is found
func (r *noopReaderImpl) GetUserFromEmail(
	ctx context.Context,
	email string,
	db database.Database,
) (*User, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

// GetUserFromAuth0Id will return database.ErrNoRows if no User is found
func (r *noopReaderImpl) GetUserFromAuth0Id(
	ctx context.Context,
	auth0Id string,
	db database.Database,
) (*User, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

// GetOrganizationAdmin will return database.ErrNoRows if no admin User is found
func (r *noopReaderImpl) GetOrganizationAdmin(
	ctx context.Context,
	organizationId string,
	db database.Database,
) (*User, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetUsersInOrganization(
	ctx context.Context,
	organizationId string,
	db database.Database,
) ([]User, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetWatchersByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) ([]User, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (w *noopWriterImpl) ResetApiKey(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*User, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteUser(ctx context.Context, id uuid.UUID, db database.Database) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteUserWithAuth0Id(
	ctx context.Context,
	auth0Id string,
	db database.Database,
) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}
