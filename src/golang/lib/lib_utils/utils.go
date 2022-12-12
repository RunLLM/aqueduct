package lib_utils

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"

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

// RunCmd executes command with arg.
// It returns the stdout, stderr, and an error object that contains an informative message that
// combines stdout and stderr.
func RunCmd(command string, arg ...string) (string, string, error) {
	cmd := exec.Command(command, arg...)
	cmd.Env = os.Environ()

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("Error running command: %s. Stdout: %s, Stderr: %s.", command, outb.String(), errb.String())
		log.Errorf(errMsg)
		return outb.String(), errb.String(), errors.New(errMsg)
	}

	return outb.String(), errb.String(), nil
}
