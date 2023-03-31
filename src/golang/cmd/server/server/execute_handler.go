package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/response"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

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

	// errors.GetMessage(err) returns both the outer and the inner error messages, excluding stack trace.
	response.SendErrorResponse(w, errors.GetMessage(err), statusCode)
}

func HandleSuccess(
	ctx context.Context,
	server Server,
	handlerObj handler.Handler,
	w http.ResponseWriter,
	r *http.Request,
	resp interface{},
) {
	server.Log(ctx, handlerObj.Name(), r, http.StatusOK, nil)
	handlerObj.SendResponse(w, resp)
}

func ExecuteHandler(server *AqServer, handlerObj handler.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if handlerObj.Name() != new(handler.ConfigureStorageHandler).Name() {
			// ConfigureStorageHandler requests an exclusive Lock on RequestMutex,
			// so there would be dead-lock if this request first acquired a shared lock
			server.RequestMutex.RLock()
			defer server.RequestMutex.RUnlock()
		}

		args, statusCode, err := handlerObj.Prepare(r)
		ctx := r.Context()
		if err != nil {
			HandleError(ctx, server, handlerObj.Name(), w, r, statusCode, err)
			return
		}
		resp, statusCode, err := handlerObj.Perform(ctx, args)
		if err != nil {
			HandleError(ctx, server, handlerObj.Name(), w, r, statusCode, err)
			return
		}
		HandleSuccess(ctx, server, handlerObj, w, r, resp)
	}
}
