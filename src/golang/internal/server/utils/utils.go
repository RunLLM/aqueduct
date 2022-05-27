package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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

	w.Header().Set(ContentTypeHeader, "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonBlob)
}

//	Given an http request, this helper function extract its payload as a
//	bytestring from its `Body` field. Currently, this function supports two
//	`contentType`s: `application/octet-stream` and `multipart/form-data`.
//	For `multipart/form-data`, since the request's `Body` comes from a file
//	upload, the caller should specify the name of the file in `fileName`.
//	This argument is ignored for `application/octet-stream`.
func ExtractHttpPayload(contentType, fileName string, isFile bool, r *http.Request) ([]byte, error) {
	payload := []byte{}
	var err error

	if strings.Contains(contentType, "multipart/form-data") {
		if isFile {
			//	The request comes from the UI as file upload. We use
			//	`strings.Contains` instead of an equality check because
			//	`multipart/form-data` is typically followed by a boundary string that
			//	varies across requests, so we want to omit that part.
			//	Limit max input length.
			r.ParseMultipartForm(32 << 20)
			var buf bytes.Buffer

			file, header, err := r.FormFile(fileName)
			if err != nil {
				return payload, err
			}

			log.Printf("filename is %v, size is %v", header.Filename, header.Size)

			defer file.Close()

			io.Copy(&buf, file)
			payload = buf.Bytes()
		} else {
			// If we reach here, it means the value is sent as `String` instead of `File`,
			// so we use `FormValue` to extract the payload and return its byte form.
			// The caller can later parse the returned value to string.
			value := r.FormValue(fileName)
			payload = []byte(value)
		}
	} else if contentType == "application/x-www-form-urlencoded" {
		// This is for cases where the HTTP body doesn't contain any files, in which case the SDK's `requests` lib
		// will send the request with this content type.
		value := r.FormValue(fileName)
		payload = []byte(value)
	} else {
		return payload, errors.Newf("Unsupported content type header: %s.", contentType)
	}

	return payload, err
}

// Send small content (that fits in single-machine memory) in binary
func SendSmallFileResponse(w http.ResponseWriter, fileName string, content *bytes.Buffer) {
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set(ContentTypeHeader, "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	io.Copy(w, content)
}

func ValidateDagOperatorIntegrationOwnership(
	ctx context.Context,
	operators map[uuid.UUID]operator.Operator,
	organizationId string,
	integrationReader integration.Reader,
	db database.Database,
) (bool, error) {
	for _, operator := range operators {
		var integrationId uuid.UUID
		if operator.Spec.IsExtract() {
			integrationId = operator.Spec.Extract().IntegrationId
		} else if operator.Spec.IsLoad() {
			integrationId = operator.Spec.Load().IntegrationId
		} else {
			continue
		}

		ok, err := integrationReader.ValidateIntegrationOwnership(
			ctx,
			integrationId,
			organizationId,
			db,
		)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}
