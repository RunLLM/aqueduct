package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/utils"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
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

var ErrUnsupportedAuthMethod = errors.New("Auth method is not supported.")

// Logs the full internal error message and sends the external error message back to the client.
func HandleError(
	ctx context.Context,
	server Server,
	handlerName string,
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	err error,
) {
	if err == nil {
		log.Fatal("Cannot pass no error into handleError()!")
	}
	server.Log(ctx, handlerName, r, statusCode, err)

	var externalMsg string
	dbxErr, ok := err.(errors.DropboxError)
	if ok {
		// Only return the top-level error message to clients.
		externalMsg = dbxErr.GetMessage()
	} else {
		externalMsg = err.Error()
	}
	utils.SendErrorResponse(w, externalMsg, statusCode)
}

func HandleSuccess(
	ctx context.Context,
	server Server,
	handler Handler,
	w http.ResponseWriter,
	r *http.Request,
	resp interface{},
) {
	server.Log(ctx, handler.Name(), r, http.StatusOK, nil)
	handler.SendResponse(w, resp)
}

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

func ExecuteHandler(server Server, handler Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		args, statusCode, err := handler.Prepare(r)
		ctx := r.Context()
		if err != nil {
			HandleError(ctx, server, handler.Name(), w, r, statusCode, err)
			return
		}
		resp, statusCode, err := handler.Perform(ctx, args)
		if err != nil {
			HandleError(ctx, server, handler.Name(), w, r, statusCode, err)
			return
		}
		HandleSuccess(ctx, server, handler, w, r, resp)
	}
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
	utils.SendJsonResponse(w, resp, http.StatusOK)
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
	utils.SendJsonResponse(w, resp, http.StatusOK)
}
