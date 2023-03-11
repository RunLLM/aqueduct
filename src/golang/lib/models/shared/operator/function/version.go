package function

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
)

type PythonVersion string

const (
	PythonVersionUnknown PythonVersion = "unknown"
	PythonVersion37      PythonVersion = "3.7"
	PythonVersion38      PythonVersion = "3.8"
	PythonVersion39      PythonVersion = "3.9"
	PythonVersion310     PythonVersion = "3.10"
)

const pythonVersionFile = "python_version.txt"

// GetPythonVersion returns the Python version required by the function code stored
// at `path`.
func GetPythonVersion(ctx context.Context, path string, storageConfig *shared.StorageConfig) (PythonVersion, error) {
	program, err := storage.NewStorage(storageConfig).Get(ctx, path)
	if err != nil {
		return PythonVersionUnknown, err
	}

	return readPythonVersion(program)
}

// readPythonVersion looks for a file named `python_version.txt` in `program`, which
// is zipped file. It returns the Python version found in the file.
func readPythonVersion(program []byte) (PythonVersion, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(program), int64(len(program)))
	if err != nil {
		return PythonVersionUnknown, err
	}

	var versionFile *zip.File
	for _, zipFile := range zipReader.File {
		parts := strings.Split(zipFile.Name, "/")
		if len(parts) == 2 && parts[1] == pythonVersionFile {
			versionFile = zipFile
			break
		}
	}

	if versionFile == nil {
		return PythonVersionUnknown, errors.New("Unable to find python_version.txt in serialized function.")
	}

	f, err := versionFile.Open()
	if err != nil {
		return PythonVersionUnknown, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return PythonVersionUnknown, err
	}

	version := string(content)
	version = strings.TrimSpace(version)

	pythonVersion := PythonVersion(version)
	switch pythonVersion {
	case PythonVersion37, PythonVersion38, PythonVersion39, PythonVersion310:
		return pythonVersion, nil
	default:
		return PythonVersionUnknown, errors.Newf("Unknown Python version %v", pythonVersion)
	}
}
