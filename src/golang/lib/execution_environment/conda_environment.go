package execution_environment

import (
	"fmt"

	"github.com/aqueducthq/aqueduct/lib"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	log "github.com/sirupsen/logrus"
)

var pythonVersions = [...]string{
	"3.7",
	"3.8",
	"3.9",
	"3.10",
}

func baseEnvNameByVersion(pythonVersion string) string {
	return fmt.Sprintf("aqueduct_python%s", pythonVersion)
}

// createBaseEnvs creates base python environments.
func createBaseEnvs() error {
	for _, pythonVersion := range pythonVersions {
		envName := baseEnvNameByVersion(pythonVersion)
		args := []string{
			"create",
			"-n",
			envName,
			fmt.Sprintf("python==%s", pythonVersion),
			"-y",
		}
		_, _, err := lib_utils.RunCmd(CondaCmdPrefix, args...)
		if err != nil {
			return err
		}

		args = []string{
			"run",
			"-n",
			envName,
			"pip3",
			"install",
			fmt.Sprintf("aqueduct-ml==%s", lib.ServerVersionNumber),
		}
		_, _, err = lib_utils.RunCmd(CondaCmdPrefix, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateCondaEnv(e *ExecutionEnvironment, condaPath string) error {
	// First, we create a conda env with the env object's Python version.
	createArgs := []string{
		"create",
		"-n",
		e.Name(),
		fmt.Sprintf("python==%s", e.PythonVersion),
		"-y",
	}

	_, _, err := lib_utils.RunCmd(CondaCmdPrefix, createArgs...)
	if err != nil {
		return err
	}

	forkEnvPath := fmt.Sprintf(
		"%s/envs/aqueduct_python%s/lib/python%s/site-packages",
		condaPath,
		e.PythonVersion,
		e.PythonVersion,
	)
	forkArgs := []string{
		"develop",
		"-n",
		e.Name(),
		forkEnvPath,
	}

	_, _, err = lib_utils.RunCmd(CondaCmdPrefix, forkArgs...)
	if err != nil {
		return err
	}

	// Then, we use pip3 to install dependencies inside this new Conda env.
	installArgs := append([]string{
		"run",
		"-n",
		e.Name(),
		"pip3",
		"install",
	}, e.Dependencies...)

	_, _, err = lib_utils.RunCmd(CondaCmdPrefix, installArgs...)
	if err != nil {
		return err
	}

	return nil
}

func deleteCondaEnv(name string) error {
	args := []string{
		"env",
		"remove",
		"-n",
		name,
	}

	_, _, err := lib_utils.RunCmd(CondaCmdPrefix, args...)
	return err
}

func DeleteCondaEnv(e *ExecutionEnvironment) error {
	return deleteCondaEnv(e.Name())
}

func DeleteBaseEnvs() error {
	for _, pythonVersion := range pythonVersions {
		err := deleteCondaEnv(baseEnvNameByVersion(pythonVersion))
		if err != nil {
			return err
		}
	}

	return nil
}

// Best-effort to delete all envs and log any error
func DeleteCondaEnvs(envs []ExecutionEnvironment) {
	for _, env := range envs {
		err := DeleteCondaEnv(&env)
		if err != nil {
			log.Errorf("Failed to delete env %s: %v", env.ID.String(), err)
		}
	}
}
