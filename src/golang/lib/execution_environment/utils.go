package execution_environment

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

// runCmd executes the command and on error, returns an informative error message that combines
// outputs from stdout and stderr.
func runCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Env = os.Environ()

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("Error running command: %s. Stdout: %s, Stderr: %s.", name, outb.String(), errb.String())
		log.Errorf(errMsg)
		return errors.New(errMsg)
	}

	return nil
}
