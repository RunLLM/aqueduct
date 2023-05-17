package execution_environment

import (
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetServerPythonVersion performs a best-effort fetch of the server's python version. This should only be used when conda is not available.
func GetServerPythonVersion() string {
	var version strings.Builder
	cmd := exec.Command(
		"python3",
		"--version",
	)
	cmd.Stdout = &version
	err := cmd.Run()
	if err != nil {
		log.Errorf("Could not get Python version on server: %v", err)
		return ""
	}
	return strings.ReplaceAll(version.String(), "\n", "")
}
