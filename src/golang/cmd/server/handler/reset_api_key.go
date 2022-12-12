package handler

import (
	"context"
	"net/http"

	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
)

type resetApiKeyArgs struct {
	*aq_context.AqContext
}

type resetApiKeyResponse struct {
	ApiKey string `json:"apiKey"`
}

// Route: /api/keys/reset
// Method: POST
// Request:
//
//	Headers:
//		`api-key`: user's current API Key
//
// Response: serialized `resetApiKeyResponse` object containing the new key.
type ResetApiKeyHandler struct {
	PostHandler

	Database database.Database
	UserRepo repos.User
}

func (*ResetApiKeyHandler) Name() string {
	return "ResetApiKey"
}

func (*ResetApiKeyHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to reset API key.")
	}

	return &resetApiKeyArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

func (h *ResetApiKeyHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*resetApiKeyArgs)
	emptyResp := resetApiKeyResponse{}

	user, err := h.UserRepo.ResetAPIKey(ctx, args.ID, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to reset API key.")
	}

	return resetApiKeyResponse{
		ApiKey: user.APIKey,
	}, http.StatusOK, nil
}
