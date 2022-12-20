package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/response"
)

type RequestMethod string

const (
	GetRequestMethod  RequestMethod = "GET"
	PostRequestMethod RequestMethod = "POST"
)

type AuthMethod string

const (
	ApiKeyAuthMethod AuthMethod = "ApiKey"
)

type Handler interface {
	Name() string
	// Extra headers additional to those required by auth.
	Headers() []string
	// 'GET' or 'POST'
	Method() RequestMethod
	// Auth on this route. For now, we supports APIKey.
	AuthMethod() AuthMethod
	// Parse the request and returns structured arguments of the request as an `interface{}`
	Prepare(r *http.Request) (interface{}, int, error)
	// Takes the parsed request and actually handle the request.
	// Implementation of this method should cast `interfaceArgs` to the proper struct type
	// matching what returned by `Prepare()`.
	Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error)
	// Send response back.
	SendResponse(w http.ResponseWriter, resp interface{})
}

type GetHandler struct{}

func (*GetHandler) Method() RequestMethod {
	return GetRequestMethod
}

func (*GetHandler) AuthMethod() AuthMethod {
	return ApiKeyAuthMethod
}

func (*GetHandler) Headers() []string {
	return nil
}

func (*GetHandler) SendResponse(w http.ResponseWriter, resp interface{}) {
	response.SendJsonResponse(w, resp, http.StatusOK)
}

type PostHandler struct{}

func (*PostHandler) Method() RequestMethod {
	return PostRequestMethod
}

func (*PostHandler) AuthMethod() AuthMethod {
	return ApiKeyAuthMethod
}

func (*PostHandler) Headers() []string {
	return nil
}

func (*PostHandler) SendResponse(w http.ResponseWriter, resp interface{}) {
	response.SendJsonResponse(w, resp, http.StatusOK)
}
