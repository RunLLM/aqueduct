package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/context_parsing"
)

type GetUserProfileHandler struct {
	GetHandler
}

func (*GetUserProfileHandler) Name() string {
	return "GetUserProfile"
}

func (*GetUserProfileHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return context_parsing.ParseAqContext(r.Context())
}

func (*GetUserProfileHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*context_parsing.AqContext)
	return args.User, http.StatusOK, nil
}
