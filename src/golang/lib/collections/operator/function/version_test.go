package function

import (
	"archive/zip"
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadPythonVersion(t *testing.T) {
	type test struct {
		version PythonVersion
	}

	tests := []test{
		{PythonVersion37},
		{PythonVersion38},
		{PythonVersion39},
		{PythonVersion310},
	}

	for _, tc := range tests {
		data := createZipWithVersion(t, tc.version)
		version, err := readPythonVersion(data)
		require.Nil(t, err)
		require.Equal(t, version, tc.version)
	}
}

func TestReadPythonVersionUnknown(t *testing.T) {
	data := createZipWithVersion(t, PythonVersionUnknown)
	version, err := readPythonVersion(data)
	require.NotNil(t, err)
	require.Equal(t, version, PythonVersionUnknown)
}

// createZipWithVersion creates a zipped file containing `python_version.txt`
// with `version` written inside. It returns the zipped file.
func createZipWithVersion(t *testing.T, version PythonVersion) []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	versionFile, err := zipWriter.Create(fmt.Sprintf("test/%s", pythonVersionFile))
	require.Nil(t, err)

	content := []byte(version)
	_, err = versionFile.Write(content)
	require.Nil(t, err)

	err = zipWriter.Close()
	require.Nil(t, err)

	return buf.Bytes()
}
