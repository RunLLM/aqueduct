package response

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
)

type EmptyResponse struct{}

type ErrorResponse struct {
	Error string `json:"error"`
}

func SendErrorResponse(w http.ResponseWriter, errorMsg string, errorCode int) {
	response := ErrorResponse{Error: errorMsg}
	SendJsonResponse(w, response, errorCode)
}

//	Internal utility function to send a JSON response by automatically
//	serializing the provided response object and adding the provided statusCode
//	as the response's status code. The response object is expected to be
//	JSON-serializable.
func SendJsonResponse(w http.ResponseWriter, response interface{}, statusCode int) {
	jsonBlob, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(routes.ContentTypeHeader, "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonBlob)
}

// Send small content (that fits in single-machine memory) in binary
func SendSmallFileResponse(w http.ResponseWriter, fileName string, content *bytes.Buffer) {
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set(routes.ContentTypeHeader, "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	io.Copy(w, content)
}
