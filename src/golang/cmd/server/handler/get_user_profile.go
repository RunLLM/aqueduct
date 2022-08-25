package handler

import (
	"context"
	"net/http"

	aq_context "github.com/aqueducthq/aqueduct/lib/context"
)

// Route: /api/user
// Method: GET
// Params: None
//
// Request
//
//	Headers:
//		`api-key`: user's API Key
//
// Response:
//	Body:
//		serialized `User` object.
type GetUserProfileHandler struct {
	GetHandler
}

func (*GetUserProfileHandler) Name() string {
	return "GetUserProfile"
}

func (*GetUserProfileHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return aq_context.ParseAqContext(r.Context())
}

func (*GetUserProfileHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*aq_context.AqContext)
	return args.User, http.StatusOK, nil
}
