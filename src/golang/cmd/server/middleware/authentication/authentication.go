package authentication

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/response"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/repos"
)

// RequireApiKey expects a request whose header contains key `api-key`
// for authorization purposes. If the authorization is successful,
// it forwards the request to the controller. Otherwise, it sends an http response
// in JSON format with an `error` message.
func RequireApiKey(userRepo repos.User, db database.Database) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(routes.ApiKeyHeader)

			user, err := userRepo.GetByAPIKey(r.Context(), apiKey, db)
			if errors.Is(err, database.ErrNoRows()) {
				response.SendErrorResponse(w, "Invalid API key credentials.", http.StatusForbidden)
			} else if err != nil {
				// Something went wrong with accessing the database
				response.SendErrorResponse(w, "Unable to validate API key credentials.", http.StatusForbidden)
			} else {
				// Create a new context with userId and organizationId.
				contextWithUserId := context.WithValue(r.Context(), aq_context.UserIdKey, user.ID.String())
				contextWithOrganizationId := context.WithValue(contextWithUserId, aq_context.OrganizationIdKey, user.OrgID)
				contextWithUserAuth0Id := context.WithValue(contextWithOrganizationId, aq_context.UserAuth0IdKey, user.Auth0ID)
				h.ServeHTTP(w, r.WithContext(contextWithUserAuth0Id))
			}
		})
	}
}
