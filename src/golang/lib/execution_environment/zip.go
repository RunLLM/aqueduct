package execution_environment

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/dropbox/godropbox/errors"
)

const (
	ReqFileName           = "requirements.txt"
	PythonVersionFileName = "python_version"
)

var (
	ErrPythonVersionMissing    = errors.New("Python version file is missing.")
	ErrRequirementsFileMissing = errors.New("Requirement file is missing.")
)

func InferDependenciesFromZipFile(zipball []byte) (*ExecutionEnvironment, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipball), int64(len(zipball)))
	if err != nil {
		return nil, err
	}

	env := &ExecutionEnvironment{}
	hasReqFile := false
	hasVersionFile := false

	for _, zipFile := range zipReader.File {
		if strings.Contains(zipFile.Name, ReqFileName) || strings.Contains(zipFile.Name, PythonVersionFileName) {
			isReqFile := strings.Contains(zipFile.Name, ReqFileName)
			reader, err := zipFile.Open()
			if err != nil {
				return nil, err
			}

			buf, err := ioutil.ReadAll(reader)
			if err != nil {
				return nil, err
			}

			contents := string(buf)
			if isReqFile {
				rows := strings.Split(contents, "\n")
				normalizedRows := make([]string, 0, len(rows))
				for _, row := range rows {
					normalizedRow := strings.TrimSpace(row)
					// Deal with empty rows
					if normalizedRow != "" {
						normalizedRows = append(
							normalizedRows, normalizedRow,
						)
					}
				}

				sort.Strings(normalizedRows)
				env.Dependencies = normalizedRows
				hasReqFile = true
				continue
			}

			// otherwise it's a python version file
			env.PythonVersion = strings.TrimSpace(contents)
			hasVersionFile = true
		}
	}

	if !hasReqFile {
		return nil, ErrRequirementsFileMissing
	}

	if !hasVersionFile {
		return nil, ErrPythonVersionMissing
	}

	return env, nil
}
