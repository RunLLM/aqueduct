package aq_context

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type contextKeyType string

const (
	UserIdKey         contextKeyType = "userId"
	OrganizationIdKey contextKeyType = "organizationId"
	UserRequestIdKey  contextKeyType = "userRequestId"
	UserAuth0IdKey    contextKeyType = "userAuth0Id"
)

type AqContext struct {
	models.User
	RequestID string
	// StorageConfig is a copy of the global storage config as of the creation of this AqContext
	StorageConfig *shared.StorageConfig
}

// Most routes will first go through `RequireApiKey` middleware, which assigns the user-related
// fields to the context, so the parsing below should never fail in these cases.
func ParseAqContext(ctx context.Context) (*AqContext, int, error) {
	userIdRaw := ctx.Value(UserIdKey)
	if userIdRaw == nil {
		return nil, http.StatusBadRequest, errors.New("No UserID supplied on request context.")
	}

	userIdStr, ok := userIdRaw.(string)
	if !ok {
		return nil, http.StatusBadRequest, errors.New("Unable to convert UserID to string.")
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to convert UserID to uuid.")
	}

	organizationIdRaw := ctx.Value(OrganizationIdKey)
	if organizationIdRaw == nil {
		return nil, http.StatusBadRequest, errors.New("No Organization ID supplied on request context.")
	}

	organizationId, ok := organizationIdRaw.(string)
	if !ok {
		return nil, http.StatusBadRequest, errors.New("Unable to convert Organization ID to string.")
	}

	auth0IdRaw := ctx.Value(UserAuth0IdKey)
	if auth0IdRaw == nil {
		return nil, http.StatusBadRequest, errors.New("No Auth0 ID supplied on request context.")
	}

	auth0Id, ok := auth0IdRaw.(string)
	if !ok {
		return nil, http.StatusBadRequest, errors.New("Unable to convert Auth0 ID to string.")
	}

	// No valid request ID is not a blocking issue.
	requestId, ok := ctx.Value(UserRequestIdKey).(string)
	if !ok {
		log.Warning("Seems that request ID is not properly generated.")
	}

	// The same StorageConfig should be used for the context of a request
	storageConfig := config.Storage()

	return &AqContext{
		User: models.User{
			ID:      userId,
			OrgID:   organizationId,
			Auth0ID: auth0Id,
		},
		RequestID:     requestId,
		StorageConfig: &storageConfig,
	}, http.StatusOK, nil
}
