package lib_utils

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This function appends a prefix to the resource name
// so that it conforms to the k8s's accepted format (name must start with an alphabet).
func AppendPrefix(name string) string {
	return fmt.Sprintf("aqueduct-%s", name)
}

func ParseStatus(st *status.Status) (string, int) {
	var errorMsg string
	var ok bool

	if len(st.Details()) == 0 {
		errorMsg = st.Message()
	} else {
		errorMsg, ok = st.Details()[0].(string) // Details should only have one object, and it should be a string.
		if !ok {
			log.Errorf("Unable to correctly parse gRPC status: %v\n", st)
		}
	}

	var errorCode int
	if st.Code() == codes.InvalidArgument {
		errorCode = http.StatusBadRequest
	} else if st.Code() == codes.Internal {
		errorCode = http.StatusInternalServerError
	} else if st.Code() == codes.NotFound {
		errorCode = http.StatusNotFound
	} else {
		errorCode = http.StatusInternalServerError
	}

	return errorMsg, errorCode
}
