package server

import (
	"net/http"

	"github.com/aqueducthq/aqueduct/internal/server/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Every request MUST contain these arguments.
type CommonArgs struct {
	user.User
	RequestId string
}

func (c *CommonArgs) getOrganizationId() string {
	return c.User.OrganizationId
}

func (c *CommonArgs) getRequestId() string {
	return c.RequestId
}

// Most routes will first go through `RequireApiKey` middleware, which assigns the user-related
// fields to the context, so the parsing below should never fail in these cases.
func ParseCommonArgs(r *http.Request) (*CommonArgs, int, error) {
	userIdRaw := r.Context().Value(utils.UserIdKey)
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

	organizationIdRaw := r.Context().Value(utils.OrganizationIdKey)
	if organizationIdRaw == nil {
		return nil, http.StatusBadRequest, errors.New("No Organization ID supplied on request context.")
	}

	organizationId, ok := organizationIdRaw.(string)
	if !ok {
		return nil, http.StatusBadRequest, errors.New("Unable to convert Organization ID to string.")
	}

	auth0IdRaw := r.Context().Value(utils.UserAuth0IdKey)
	if auth0IdRaw == nil {
		return nil, http.StatusBadRequest, errors.New("No Auth0 ID supplied on request context.")
	}

	auth0Id, ok := auth0IdRaw.(string)
	if !ok {
		return nil, http.StatusBadRequest, errors.New("Unable to convert Auth0 ID to string.")
	}

	// No valid request ID is not a blocking issue.
	requestId, ok := r.Context().Value(utils.UserRequestIdKey).(string)
	if !ok {
		log.Warning("Seems that request ID is not properly generated.")
	}

	return &CommonArgs{
		User: user.User{
			Id:             userId,
			OrganizationId: organizationId,
			Auth0Id:        auth0Id,
		},
		RequestId: requestId,
	}, http.StatusOK, nil
}
