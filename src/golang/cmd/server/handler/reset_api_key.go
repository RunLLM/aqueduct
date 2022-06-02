package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/collections/user"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
)

type resetApiKeyArgs struct {
	*aq_context.AqContext
}

type resetApiKeyResponse struct {
	ApiKey string `json:"apiKey"`
}

type ResetApiKeyHandler struct {
	PostHandler

	Database   database.Database
	UserWriter user.Writer
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

	userObject, err := h.UserWriter.ResetApiKey(ctx, args.Id, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to reset API key.")
	}

	return resetApiKeyResponse{
		ApiKey: userObject.ApiKey,
	}, http.StatusOK, nil
}
