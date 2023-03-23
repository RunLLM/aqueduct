package request

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

// Given an http request, this helper function extract its payload as a
// bytestring from its `Body` field. Currently, this function supports two
// `contentType`s: `application/octet-stream` and `multipart/form-data`.
// For `multipart/form-data`, since the request's `Body` comes from a file
// upload, the caller should specify the name of the file in `fileName`.
// This argument is ignored for `application/octet-stream`.
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
			err = r.ParseMultipartForm(32 << 20)
			if err != nil {
				return nil, err
			}
			var buf bytes.Buffer

			file, header, err := r.FormFile(fileName)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to read file and header from request.")
			}

			log.Printf("filename is %v, size is %v", header.Filename, header.Size)

			defer func() {
				err = file.Close()
				if err != nil {
					log.Errorf("Unable to close file descriptor used to extract HTTP payload from request.")
				}
			}()

			_, err = io.Copy(&buf, file)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to read file from request.")
			}
			payload = buf.Bytes()
		} else {
			// If we reach here, it means the value is sent as `String` instead of `File`,
			// so we use `FormValue` to extract the payload and return its byte form.
			// The caller can later parse the returned value to string. By default, the
			// request's body size is capped at `defaultMaxMemory = 32 << 20`. In addition
			// `multipartReader` gives an addition of 10MB upon defaultMaxMemory for
			// nonfile data. Therefore this adds up to a maximum request body size of 42MB.
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
