package server

import (
	"context"
	"net/http"
)

type GetUserProfileHandler struct {
	GetHandler
}

func (*GetUserProfileHandler) Name() string {
	return "GetUserProfile"
}

func (*GetUserProfileHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return ParseCommonArgs(r)
}

func (*GetUserProfileHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*CommonArgs)
	return args.User, http.StatusOK, nil
}
