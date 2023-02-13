package job

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/stretchr/testify/require"
)

func TestExtractAwsCredentials(t *testing.T) {
	credentialsFilepath := filepath.Join(t.TempDir(), "credentials_test")
	f, err := os.Create(credentialsFilepath)
	require.Nil(t, err)
	f.WriteString(
		`[default]
aws_access_key_id=dummyid
aws_secret_access_key=dummykey`,
	)
	f.Close()

	config := &shared.S3Config{
		Region:             "us-east-2",
		Bucket:             "dummybucket",
		CredentialsPath:    credentialsFilepath,
		CredentialsProfile: "default",
	}
	// Expect proper extraction
	id, key, err := extractAwsCredentials(config)
	require.Nil(t, err)
	require.Equal(t, "dummyid", id)
	require.Equal(t, "dummykey", key)

	config.CredentialsProfile = "user"
	// Expect error to be thrown for unknown profile
	id, key, err = extractAwsCredentials(config)
	require.Equal(t, "", id)
	require.Equal(t, "", key)
	require.Error(t, err)
}
